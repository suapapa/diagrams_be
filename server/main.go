package main

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
)

var (
	diagramOut = "diagrams_image.png"
)

type result struct {
	Err string
	Img string
}

func main() {
	// check diagramOut exists
	_, err := os.Stat(diagramOut)
	if err != nil {
		json.NewEncoder(os.Stdout).Encode(result{Err: err.Error()})
		os.Exit(-1)
	}
	defer os.RemoveAll(diagramOut)

	f, err := os.Open(diagramOut)
	if err != nil {
		json.NewEncoder(os.Stdout).Encode(result{Err: err.Error()})
		os.Exit(-1)
	}
	defer f.Close()

	content, err := ioutil.ReadAll(f)
	encoded := base64.StdEncoding.EncodeToString(content)
	json.NewEncoder(os.Stdout).Encode(result{Img: encoded})
}
