package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

type result struct {
	Err string `json:"err"`
	Img string `json:"img"`
}

var (
	diagramIn  = "diagram.py"
	diagramOut = "diagrams_image.png"
)

func main() {
	// read diagrams code from stdin
	w, err := os.Create(diagramIn)
	checkErr(err)
	io.Copy(w, os.Stdin)
	w.Close()

	// run diagrams code with python (this program should run in gVisor)
	cmd := exec.Command("python", diagramIn)
	err = cmd.Run()
	checkErr(err)

	// find out diagramOut exists
	match, err := filepath.Glob("*.png")
	checkErr(err)
	if len(match) != 1 {
		checkErr(fmt.Errorf("fail to gen diagram png"))
	}
	diagramOut = match[0]

	_, err = os.Stat(diagramOut)
	checkErr(err)
	defer os.RemoveAll(diagramOut)

	f, err := os.Open(diagramOut)
	checkErr(err)
	defer f.Close()

	content, err := ioutil.ReadAll(f)
	encoded := base64.StdEncoding.EncodeToString(content)
	json.NewEncoder(os.Stdout).Encode(result{Img: encoded})
}

func checkErr(err error) {
	if err != nil {
		json.NewEncoder(os.Stdout).Encode(result{Err: err.Error()})
		os.Exit(-1)
	}
}
