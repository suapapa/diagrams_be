package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
)

// this type should be matched with sandbox
type result struct {
	Err string `json:"err"`
	Img string `json:"img"`
}

const (
	diagramsNodesJSON = "diagrams_nodes.json"
	containerName     = "suapapa/diagrams-server-gvisor:latest"
)

var listenAddr string

func main() {
	flag.StringVar(&listenAddr, "l", ":8888", "listen address")
	flag.Parse()

	http.HandleFunc("/diagram", handleDiagram)
	http.HandleFunc("/nodes", handleNodes)
	http.Handle("/", http.FileServer(http.Dir("./dist")))

	log.Println("listen and serve at :8888...")
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		panic(err)
	}
}

func handleDiagram(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	// read diagram python code from frontend
	// io.Copy(os.Stdout, r.Body)
	defer r.Body.Close()
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// pass it to diagrams container (gVisor)
	// write result png to respone writer

	name := "diagram_run_" + randHex(8)
	cmd := exec.Command("docker", "run",
		"--name="+name,
		"--rm",
		"-i", // read stdin

		// "--runtime=runsc",
		"--network=none",
		"--memory="+fmt.Sprint(memoryLimitBytes),

		containerName,
	)
	cmd.Stdin = r.Body
	cmd.Stdout = w
	cmd.Run()
}

func handleNodes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// check if diagramsNodesJSON exists
	_, err := os.Stat(diagramsNodesJSON)
	// if not create one
	if os.IsNotExist(err) {
		fw, err := os.Create(diagramsNodesJSON)
		if err != nil {
			panic(err) // TODO: fix to internal server error
		}
		cmd := exec.Command("docker", "run",
			// "-it",
			"--rm",
			// "--runtime=runsc",
			// "--network=none",
			"--entrypoint=/usr/local/bin/python",
			containerName,
			"listup_nodes.py",
		)
		cmd.Stdout = fw
		cmd.Run()
		fw.Close()
	} else if err != nil {
		panic(err) // TODO: fix to internal server error
	}
	// and serve
	http.ServeFile(w, r, diagramsNodesJSON)
}
