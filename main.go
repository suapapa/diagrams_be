package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os/exec"
	"strings"
)

const (
	memoryLimitBytes = 64 * 1024 * 1024
)

var (
	listenAddr       string
	sandboxContainer string
	urlPrefix        string
	maxContentLength int64

	programName = "diagrams"
	programVer  = "dev"
)

func main() {
	log.WithField("alert", "telegram").Infof("%s start", programName)
	defer log.WithField("alert", "telegram").Infof("%s exit", programName)

	flag.StringVar(&listenAddr, "l", ":8080", "listen address")
	flag.StringVar(&sandboxContainer, "c", "suapapa/diagrams:latest", "diagrams container image")
	flag.StringVar(&urlPrefix, "p", "/diagrams-srv", "url prefix")
	flag.Int64Var(&maxContentLength, "m", 2048, "max input length")
	flag.Parse()

	prepareCh := make(chan any)
	go func() {
		if err := prepare(); err != nil {
			log.Fatal(err)
		}
		prepareCh <- struct{}{}
	}()
	<-prepareCh

	if urlPrefix[0] != '/' {
		urlPrefix = "/" + urlPrefix
	}

	http.HandleFunc(urlPrefix+"/diagram", handleDiagram)
	http.HandleFunc(urlPrefix+"/nodes", handleNodes)

	log.Infof("listen and serve at %s...", listenAddr)
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		log.Fatal(err)
	}
}

type codeReq struct {
	Code string `json:"code"`
	Hash string `json:"hash,omitempty"`
}

func handleDiagram(w http.ResponseWriter, r *http.Request) {
	// read diagram python code from frontend
	defer r.Body.Close()
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", "POST,OPTIONS")

	if r.Method != "POST" {
		http.Error(w, "expected a POST", http.StatusBadRequest)
		return
	}

	if maxContentLength > 0 && r.ContentLength > maxContentLength {
		http.Error(w, "too big input", http.StatusRequestEntityTooLarge)
		return
	}

	var req codeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
	}

	if req.Hash == "" {
		h := sha1.New()
		h.Write([]byte(req.Code))
		req.Hash = hex.EncodeToString(h.Sum(nil))
	}

	log.Debugf("code: %s", req.Code)
	log.Infof("hash: %s", req.Hash)
	inBuf := strings.NewReader(req.Code)
	// TODO: check db if hash exists
	// if exists return saved diagram

	w.Header().Set("Content-Type", "application/json")
	// pass it to diagrams container (gVisor)
	// write diagrams.Result png to respone writer
	name := "diagrams_" + randHex(8)
	log.Infof("running %s", name)
	cmd := exec.Command("docker", "run",
		"--name="+name,
		"--rm",
		"-i", // read stdin
		// "--runtime=runsc",
		"--network=none",
		"--memory="+fmt.Sprint(memoryLimitBytes),
		sandboxContainer,
	)

	outBuf := bytes.NewBuffer(nil)
	cmd.Stdin = inBuf
	cmd.Stdout = outBuf // http.ResponseWriter로 JSON 출력

	if err := cmd.Run(); err != nil {
		log.Errorf("docker run error: %s, hash=%s", err, req.Hash)
		http.Error(w, "docker run error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	io.Copy(w, outBuf)
}

func handleNodes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", "GET,OPTIONS")
	w.WriteHeader(http.StatusOK)
	w.Write(diagramsNodesBytes)
}
