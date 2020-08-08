package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	dirs := []string{
		"out", //
		"out.exe", "C:\\Windows\\Notepad.exe", "some.tar.gz",
		".tgz", "C:\\Jacksome\\note\\", "C:\\Jacksome\\note\\zzzz.tar.gz",
	}
	for _, d := range dirs {
		fmt.Fprintf(os.Stderr, "Dir: [%s] --> [%s]\n", d, filepath.Dir(d))
	}
}
