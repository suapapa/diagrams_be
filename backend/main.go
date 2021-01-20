package main

import (
	"flag"
	"net/http"
)

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
	// read diagram python code from frontend
	// pass it to diagrams container (gVisor)
	// write result png to respone writer
}
