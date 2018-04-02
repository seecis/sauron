package main

import (
	"flag"
	"os"
	"log"
	"bufio"
	"io"
	"bytes"
)

func main() {

	var filePath string
	var extractorName string
	flag.StringVar(&extractorName,
		"extractor",
		"",
		"Sets the extractor to run on the provided html filePath")

	flag.StringVar(
		&filePath,
		"f",
		"",
		"Sets the filePath that extractor will be ran on")

	flag.Parse()

	var file *os.File
	var fileErr error

	if filePath != "" {
		_, err := os.Stat(filePath)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		file, fileErr = os.Open(filePath)
	} else {
		fileStat, err := os.Stdin.Stat()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		if (fileStat.Mode() & os.ModeCharDevice) == os.ModeCharDevice {
			log.Println("I need a pipe or a filePath")
			os.Exit(2)
		}

		file, fileErr = os.Open(filePath)
	}

	if fileErr != nil {
		log.Fatal(fileErr)
	}

	handleStream(bufio.NewReader(file), extractorName)
}

func handleStream(reader *bufio.Reader, extractor string) {
	tr := io.TeeReader(reader, os.Stderr)
	buf := new(bytes.Buffer)
	buf.ReadFrom(tr)
	s := buf.String()
	_ = s
}
