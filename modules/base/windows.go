// +build windows

package base

import (
	"os"
	"syscall"

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

//MoveFile file
func MoveFile(oldpath, newpath string) error {
	from, err := syscall.UTF16PtrFromString(oldpath)
	if err != nil {
		return err
	}
	to, err := syscall.UTF16PtrFromString(newpath)
	if err != nil {
		return err
	}
	return windows.MoveFileEx(from, to, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_COPY_ALLOWED)
}
