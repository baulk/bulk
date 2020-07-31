package main

import (
	"fmt"
	"os"
	"path"

	"github.com/baulk/bulk/base"
)

func main() {
	pv := []string{
		"bin/../lib/libc.a",
		"bin/../../lib/libc.a",
		"bin/../jack/../readme.a",
		"../../../../jack/some",
		"",
	}
	for _, p := range pv {
		if base.PathIsSlipVulnerability(p) {
			// skip relative path
			fmt.Fprintf(os.Stderr, "\x1b[31mVulnerability: [%s] --> [%s]\x1b[0m\n", p, path.Clean(p))
			continue
		}
		fmt.Fprintf(os.Stderr, "\x1b[32m[%s] --> [%s]\x1b[0m\n", p, path.Clean(p))
	}
}
