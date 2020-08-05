package main

import (
	"fmt"
	"os"

	"github.com/baulk/bulk/netutils"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s url\n", os.Args[0])
		os.Exit(1)
	}
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "usage: %s url. unable Getwd %v\n", os.Args[0], err)
		os.Exit(1)
	}
	executor := netutils.NewExecutor(cwd)
	executor.IsDebugMode = true
	var hsx string
	if len(os.Args) > 2 {
		hsx = os.Args[2]
	}
	file, err := executor.WebGet(os.Args[1], hsx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "download %s error %v\n", os.Args[1], err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "download success save to %s\n", file)
}
