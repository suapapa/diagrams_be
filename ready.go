package main

import (
	"bytes"
	"log"
	"os/exec"

	"github.com/pkg/errors"
)

const (
	diagramsNodesJSON = "diagrams_nodes.json"
)

var (
	diagramsNodesBytes []byte
)

// This will pull diagrams image and extract node list
func prepare() error {
	cmd := exec.Command("docker", "pull",
		sandboxContainer,
	)
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "fail to pull image")
	}

	buff := &bytes.Buffer{}
	cmd = exec.Command("docker", "run",
		"--rm",
		"--network=none",
		"--entrypoint=/usr/local/bin/python",
		sandboxContainer,
		"listup_nodes.py",
	)
	cmd.Stdout = buff
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "fail to gen "+diagramsNodesJSON)
	}

	diagramsNodesBytes = buff.Bytes()

	log.Println("ready!")
	return nil
}
