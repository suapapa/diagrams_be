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

var (
	listenAddr string
	dev        bool
)

func main() {
	flag.StringVar(&listenAddr, "l", ":8888", "listen address")
	flag.BoolVar(&dev, "d", false, "run in dev mode")
	flag.Parse()

	http.HandleFunc("/diagram", handleDiagram)
	http.HandleFunc("/nodes", handleNodes)
	if !dev {
		http.Handle("/", http.FileServer(http.Dir("./dist")))
	}

	log.Println("listen and serve at :8888...")
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		panic(err)
	}
}

func handleDiagram(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "expected a POST", http.StatusBadRequest)
		return
	}

	// read diagram python code from frontend
	defer r.Body.Close()
	if true /* dev */ {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
	w.Header().Set("Content-Type", "application/json")
	log.Println("hit diagram")
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
	if true /* dev */ {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
	log.Println("hit nodes")
	w.Header().Set("Content-Type", "application/json")
	// check if diagramsNodesJSON exists
	_, err := os.Stat(diagramsNodesJSON)
	// if not create one
	if os.IsNotExist(err) {
		fw, err := os.Create(diagramsNodesJSON)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cmd := exec.Command("docker", "run",
			"--rm",
			"--network=none",
			"--entrypoint=/usr/local/bin/python",
			containerName,
			"listup_nodes.py",
		)
		cmd.Stdout = fw
		cmd.Run()
		fw.Close()
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// and serve
	http.ServeFile(w, r, diagramsNodesJSON)
}
