// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Original code is borrowed from Go sandbox program which
// is part of Go playground (https://play.golang.org/).

package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

const (
	maxBinarySize    = 100 << 20
	startTimeout     = 30 * time.Second
	runTimeout       = 5 * time.Second
	maxOutputSize    = 100 << 20
	memoryLimitBytes = 100 << 20

	containerName = "suapapa/diagrams-server-gvisor"
)

var (
	errTooMuchOutput = errors.New("Output too large")
	errRunTimeout    = errors.New("timeout running program")
)

// containedStartMessage is the first thing written to stdout by the
// gvisor-contained process when it starts up. This lets the parent HTTP
// server know that a particular container is ready to run a binary.
const containedStartMessage = "golang-gvisor-process-started\n"

// containedStderrHeader is written to stderr after the gvisor-contained process
// successfully reads the processMeta JSON line + executable binary from stdin,
// but before it's run.
var containedStderrHeader = []byte("golang-gvisor-process-got-input\n")

var (
	readyContainer chan *Container
	runSem         chan struct{}
)

// Container represents a docker container
type Container struct {
	name string

	stdin  io.WriteCloser
	stdout *limitedWriter
	stderr *limitedWriter

	cmd       *exec.Cmd
	cancelCmd context.CancelFunc

	waitErr chan error // 1-buffered; receives error from WaitOrStop(..., cmd, ...)
}

func (c *Container) Close() {
	setContainerWanted(c.name, false)

	c.cancelCmd()
	if err := c.Wait(); err != nil {
		log.Printf("error in c.Wait() for %q: %v", c.name, err)
	}
}

func (c *Container) Wait() error {
	err := <-c.waitErr
	c.waitErr <- err
	return err
}

var (
	wantedMu        sync.Mutex
	containerWanted = map[string]bool{}
)

// setContainerWanted records whether a named container is wanted or
// not. Any unwanted containers are cleaned up asynchronously as a
// sanity check against leaks.
//
// TODO(bradfitz): add leak checker (background docker ps loop)
func setContainerWanted(name string, wanted bool) {
	wantedMu.Lock()
	defer wantedMu.Unlock()
	if wanted {
		containerWanted[name] = true
	} else {
		delete(containerWanted, name)
	}
}

func isContainerWanted(name string) bool {
	wantedMu.Lock()
	defer wantedMu.Unlock()
	return containerWanted[name]
}

func getContainer(ctx context.Context) (*Container, error) {
	select {
	case c := <-readyContainer:
		return c, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func startContainer(ctx context.Context) (c *Container, err error) {
	/*
		start := time.Now()
		defer func() {
			status := "success"
			if err != nil {
				status = "error"
			}
			// Ignore error. The only error can be invalid tag key or value length, which we know are safe.
			_ = stats.RecordWithTags(ctx, []tag.Mutator{tag.Upsert(kContainerCreateSuccess, status)},
				mContainerCreateLatency.M(float64(time.Since(start))/float64(time.Millisecond)))
		}()
	*/

	name := "play_run_" + randHex(8)
	setContainerWanted(name, true)
	cmd := exec.Command("docker", "run",
		"--name="+name,
		"--rm",
		"--tmpfs=/tmpfs:exec",
		"-i", // read stdin

		"--runtime=runsc",
		"--network=none",
		"--memory="+fmt.Sprint(memoryLimitBytes),

		containerName,
		"--mode=contained")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	pr, pw := io.Pipe()
	stdout := &limitedWriter{dst: &bytes.Buffer{}, n: maxOutputSize + int64(len(containedStartMessage))}
	stderr := &limitedWriter{dst: &bytes.Buffer{}, n: maxOutputSize}
	cmd.Stdout = &switchWriter{switchAfter: []byte(containedStartMessage), dst1: pw, dst2: stdout}
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	c = &Container{
		name:      name,
		stdin:     stdin,
		stdout:    stdout,
		stderr:    stderr,
		cmd:       cmd,
		cancelCmd: cancel,
		waitErr:   make(chan error, 1),
	}
	go func() {
		c.waitErr <- waitOrStop(ctx, cmd, os.Interrupt, 250*time.Millisecond)
	}()
	defer func() {
		if err != nil {
			c.Close()
		}
	}()

	startErr := make(chan error, 1)
	go func() {
		buf := make([]byte, len(containedStartMessage))
		_, err := io.ReadFull(pr, buf)
		if err != nil {
			startErr <- fmt.Errorf("error reading header from sandbox container: %v", err)
		} else if string(buf) != containedStartMessage {
			startErr <- fmt.Errorf("sandbox container sent wrong header %q; want %q", buf, containedStartMessage)
		} else {
			startErr <- nil
		}
	}()

	timer := time.NewTimer(startTimeout)
	defer timer.Stop()
	select {
	case <-timer.C:
		err := fmt.Errorf("timeout starting container %q", name)
		cancel()
		<-startErr
		return nil, err

	case err := <-startErr:
		if err != nil {
			return nil, err
		}
	}

	log.Printf("started container %q", name)
	return c, nil
}

// limitedWriter is an io.Writer that returns an errTooMuchOutput when the cap (n) is hit.
type limitedWriter struct {
	dst *bytes.Buffer
	n   int64 // max bytes remaining
}

// Write is an io.Writer function that returns errTooMuchOutput when the cap (n) is hit.
//
// Partial data will be written to dst if p is larger than n, but errTooMuchOutput will be returned.
func (l *limitedWriter) Write(p []byte) (int, error) {
	defer func() { l.n -= int64(len(p)) }()

	if l.n <= 0 {
		return 0, errTooMuchOutput
	}

	if int64(len(p)) > l.n {
		n, err := l.dst.Write(p[:l.n])
		if err != nil {
			return n, err
		}
		return n, errTooMuchOutput
	}

	return l.dst.Write(p)
}

// switchWriter writes to dst1 until switchAfter is written, the it writes to dst2.
type switchWriter struct {
	dst1        io.Writer
	dst2        io.Writer
	switchAfter []byte
	buf         []byte
	found       bool
}

func (s *switchWriter) Write(p []byte) (int, error) {
	if s.found {
		return s.dst2.Write(p)
	}

	s.buf = append(s.buf, p...)
	i := bytes.Index(s.buf, s.switchAfter)
	if i == -1 {
		if len(s.buf) >= len(s.switchAfter) {
			s.buf = s.buf[len(s.buf)-len(s.switchAfter)+1:]
		}
		return s.dst1.Write(p)
	}

	s.found = true
	nAfter := len(s.buf) - (i + len(s.switchAfter))
	s.buf = nil

	n, err := s.dst1.Write(p[:len(p)-nAfter])
	if err != nil {
		return n, err
	}
	n2, err := s.dst2.Write(p[len(p)-nAfter:])
	return n + n2, err
}

func randHex(n int) string {
	b := make([]byte, n/2)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", b)
}

// waitOrStop waits for the already-started command cmd by calling its Wait method.
//
// If cmd does not return before ctx is done, WaitOrStop sends it the given interrupt signal.
// If killDelay is positive, WaitOrStop waits that additional period for Wait to return before sending os.Kill.
func waitOrStop(ctx context.Context, cmd *exec.Cmd, interrupt os.Signal, killDelay time.Duration) error {
	if cmd.Process == nil {
		panic("WaitOrStop called with a nil cmd.Process — missing Start call?")
	}
	if interrupt == nil {
		panic("WaitOrStop requires a non-nil interrupt signal")
	}

	errc := make(chan error)
	go func() {
		select {
		case errc <- nil:
			return
		case <-ctx.Done():
		}

		err := cmd.Process.Signal(interrupt)
		if err == nil {
			err = ctx.Err() // Report ctx.Err() as the reason we interrupted.
		} else if err.Error() == "os: process already finished" {
			errc <- nil
			return
		}

		if killDelay > 0 {
			timer := time.NewTimer(killDelay)
			select {
			// Report ctx.Err() as the reason we interrupted the process...
			case errc <- ctx.Err():
				timer.Stop()
				return
			// ...but after killDelay has elapsed, fall back to a stronger signal.
			case <-timer.C:
			}

			// Wait still hasn't returned.
			// Kill the process harder to make sure that it exits.
			//
			// Ignore any error: if cmd.Process has already terminated, we still
			// want to send ctx.Err() (or the error from the Interrupt call)
			// to properly attribute the signal that may have terminated it.
			_ = cmd.Process.Kill()
		}

		errc <- err
	}()

	waitErr := cmd.Wait()
	if interruptErr := <-errc; interruptErr != nil {
		return interruptErr
	}
	return waitErr
}
