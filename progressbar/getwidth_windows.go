// +build windows

package progressbar

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/mattn/go-isatty"
	"golang.org/x/sys/windows"
)

// getWinSize get console size or cygwin terminal size
func getWinSize() (w int, h int, err error) {
	if isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		cmd := exec.Command("stty", "size")
		cmd.Stdin = os.Stdin
		buf, err := cmd.CombinedOutput()
		if err != nil {
			return 0, 0, err
		}
		bvv := strings.Split(strings.TrimSpace(string(buf)), " ")
		if len(bvv) < 2 {
			return 0, 0, errors.New("invaild stty result:'" + string(buf) + "'")
		}
		y, err := strconv.Atoi(bvv[0])
		if err != nil {
			return 0, 0, errors.New("invaild stty rows: " + bvv[0])
		}
		x, err := strconv.Atoi(bvv[1])
		if err != nil {
			return 0, 0, errors.New("invaild stty columns: " + bvv[1])
		}
		return x, y, nil
	}
	var info windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(windows.Handle(os.Stderr.Fd()), &info); err != nil {
		return 0, 0, err
	}
	return int(info.Size.X), int(info.Size.Y), nil
}

func getWidth() int {
	if w, _, err := getWinSize(); err == nil {
		return w
	}
	return 80
}

// const
const (
	EnableVirtualTerminalProcessingMode = 0x4
)

func init() {
	var mode uint32
	// becasue we print message to stderr
	fd := os.Stderr.Fd()
	if windows.GetConsoleMode(windows.Handle(fd), &mode) == nil {
		_ = windows.SetConsoleMode(windows.Handle(fd), mode|EnableVirtualTerminalProcessingMode)
	}
}
