package main

import (
	"debug/elf"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s elf-file\n", os.Args[0])
		os.Exit(1)
	}
	r, err := elf.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open elf file: %s error %v\n", os.Args[1], err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Machine: %v Version: %v\n", r.Machine, r.Version)
	for _, s := range r.Sections {
		fmt.Fprintf(os.Stderr, "%v\n", s.Name)
	}
	if runpath, err := r.DynString(elf.DT_RUNPATH); err == nil {
		fmt.Fprintf(os.Stderr, "runpath %v\n", runpath)
	}
	if sonname, err := r.DynString(elf.DT_SONAME); err == nil {
		fmt.Fprintf(os.Stderr, "sonname %v\n", sonname)
	}
	if rpath, err := r.DynString(elf.DT_RPATH); err == nil {
		fmt.Fprintf(os.Stderr, "rpath %v\n", rpath)
	}
	if needed, err := r.DynString(elf.DT_NEEDED); err == nil {
		fmt.Fprintf(os.Stderr, "needed %v\n", needed)
	}
}
