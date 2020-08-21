package main

import (
	"fmt"
	"os"

	"github.com/baulk/bulk/base"
	"github.com/baulk/bulk/netutils"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s url\n", os.Args[0])
		os.Exit(1)
	}
	base.IsDebugMode = true
	executor := netutils.NewExecutor(nil)
	var hsx string
	if len(os.Args) > 2 {
		hsx = os.Args[2]
	}
	file, err := executor.WebGet(&netutils.EnhanceURL{URL: os.Args[1], HashValue: hsx})
	if err != nil {
		fmt.Fprintf(os.Stderr, "download %s error %v\n", os.Args[1], err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "download success save to %s\n", file)
}
