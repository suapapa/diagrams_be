package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

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
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	cmd := exec.Command("python", diagramIn)
	cmd.Stdout = outBuf
	cmd.Stderr = errBuf
	err = cmd.Run()
	outStr := outBuf.String()
	errStr := errBuf.String()
	if err != nil {
		printJson(&diagrams.Output{Msg: outStr, Err: errStr})
		return
	}

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

	content, err := io.ReadAll(f)
	checkErr(err)
	encoded := base64.StdEncoding.EncodeToString(content)
	printJson(&diagrams.Output{Img: encoded, Msg: outStr, Err: errStr})
}

func printJson(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func checkErr(err error) {
	checkErrMsg("", err)
}

func checkErrMsg(msg string, err error) {
	if err != nil {
		ret := diagrams.Output{
			Msg: msg,
			Err: err.Error(),
		}

		json.NewEncoder(os.Stdout).Encode(&ret)
		os.Exit(-1)
	}
}
