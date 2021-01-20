package main

import (
	"flag"
	"io"
	"net/http"
	"os"
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
	io.Copy(os.Stdout, r.Body)
	defer r.Body.Close()

	// pass it to diagrams container (gVisor)
	// write result png to respone writer
	w.Write([]byte("hello world\n"))
}
