package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os/exec"
	"sync/atomic"
)

// this type should be matched with sandbox
const (
	diagramsNodesJSON = "diagrams_nodes.json"

	memoryLimitBytes = 128 * 1024 * 1024
)

var (
	listenAddr       string
	sandboxContainer string
	urlPrefix        string
	maxContentLength int64
	ready            atomic.Bool
)

func main() {
	flag.StringVar(&listenAddr, "l", ":8080", "listen address")
	flag.StringVar(&sandboxContainer, "c", "suapapa/diagrams:latest", "diagrams container image")
	flag.StringVar(&urlPrefix, "p", "/diagrams-srv", "url prefix")
	flag.Int64Var(&maxContentLength, "m", 2048, "max input length")
	flag.Parse()

	go func() {
		if err := prepare(); err != nil {
			log.Fatal(err)
		}
	}()

	if urlPrefix[0] != '/' {
		urlPrefix = "/" + urlPrefix
	}

	http.HandleFunc(urlPrefix+"/diagram", handleDiagram)
	http.HandleFunc(urlPrefix+"/nodes", handleNodes)
	http.HandleFunc(urlPrefix+"/ready", handleReady)
	// http.Handle("/", http.FileServer(http.Dir("./dist")))

	log.Printf("listen and serve at %s...", listenAddr)
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		panic(err)
	}
}

func handleDiagram(w http.ResponseWriter, r *http.Request) {
	if !ready.Load() {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "expected a POST", http.StatusBadRequest)
		return
	}

	if maxContentLength > 0 && r.ContentLength > maxContentLength {
		http.Error(w, "too big input", http.StatusRequestEntityTooLarge)
		return
	}

	// read diagram python code from frontend
	defer r.Body.Close()

	// if true /* dev */ {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// }

	w.Header().Set("Content-Type", "application/json")
	// pass it to diagrams container (gVisor)
	// write diagrams.Result png to respone writer
	name := "diagrams_" + randHex(8)
	log.Printf("running %s", name)
	cmd := exec.Command("docker", "run",
		"--name="+name,
		"--rm",
		"-i", // read stdin
		// "--runtime=runsc",
		"--network=none",
		"--memory="+fmt.Sprint(memoryLimitBytes),

		sandboxContainer,
	)
	cmd.Stdin = r.Body
	cmd.Stdout = w // http.ResponseWriter로 JSON 출력
	cmd.Run()
}

func handleNodes(w http.ResponseWriter, r *http.Request) {
	if !ready.Load() {
		return
	}

	// if true /* dev */ {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// }
	// log.Println("hit nodes")
	w.Header().Set("Content-Type", "application/json")
	w.Write(diagramsNodesBytes)
}

func handleReady(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if ready.Load() {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusProcessing)
}
