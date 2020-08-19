// +build windows

package base

import (
	"os"

	"golang.org/x/sys/windows"
)

// const
const (
	EnableVirtualTerminalProcessingMode = 0x4
)

func init() {
	FixupTerminal()
}

// FixupTerminal becasue
func FixupTerminal() {
	var mode uint32
	// becasue we print message to stderr
	fd := os.Stderr.Fd()
	if windows.GetConsoleMode(windows.Handle(fd), &mode) == nil {
		_ = windows.SetConsoleMode(windows.Handle(fd), mode|EnableVirtualTerminalProcessingMode)
	}
}
