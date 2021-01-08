package main

import (
	"fmt"
	"os"
	"time"
)

const (
	a = ^uint32(0)
	b = int64(^uint32(0))
)

func main() {
	fmt.Fprintf(os.Stderr, "%08X\n", a)
	now := time.Now()
	utc := now.UTC()
	fmt.Fprintf(os.Stderr, "%s\n", now.Format(time.RFC3339))
	fmt.Fprintf(os.Stderr, "%s\n", now.Format(time.RFC3339Nano))
	fmt.Fprintf(os.Stderr, "%s\n", utc.Format(time.RFC3339))
	fmt.Fprintf(os.Stderr, "%s\n", utc.Format(time.RFC3339Nano))
}
