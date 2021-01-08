package main

import (
	"fmt"
	"os"

	"github.com/klauspost/compress/zip"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s zipfile\n", os.Args[0])
		os.Exit(1)
	}
	fd, err := zip.OpenReader(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable open reader: %v\n", err)
		os.Exit(1)
	}
	defer fd.Close()
	for _, file := range fd.File {
		fmt.Fprintf(os.Stderr, "file: %s\n", file.Name)
	}
}
