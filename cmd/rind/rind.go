package main

import (
	"fmt"
	"os"

	"github.com/baulk/bulk/netutils"
)

// rind golang base download utils

type rindOptions struct {
	opt netutils.Options
	eu  netutils.EnhanceURL
}

func (ro *rindOptions) Invoke(val int, oa, raw string) error {

	return nil
}

func (ro *rindOptions) ParseArgv() error {

	return nil
}

func main() {
	var ro rindOptions
	if err := ro.ParseArgv(); err != nil {
		fmt.Fprintf(os.Stderr, "parse argv \x1b[31m%v\x1b[0m\n", err)
		os.Exit(1)
	}
	e := netutils.NewExecutor(&ro.opt)
	filename, err := e.WebGet(&ro.eu)
	if err != nil {
		fmt.Fprintf(os.Stderr, "download \x1b[31m%v\x1b[0m\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "download %s success\n", filename)
}
