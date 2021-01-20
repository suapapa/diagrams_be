package main

import (
	"flag"
	"fmt"
	"net/http"
	"os/exec"
)

// this type should be matched with sandbox
type result struct {
	Err string `json:"err"`
	Img string `json:"img"`
}

var listenAddr string

func main() {
	flag.StringVar(&listenAddr, "l", ":8080", "listen address")
	flag.Parse()

	http.HandleFunc("/diagram", handleDiagram)
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
