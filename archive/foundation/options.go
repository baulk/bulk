package foundation

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mattn/go-runewidth"
	"golang.org/x/crypto/ssh/terminal"
)

// ExtractOptions todo
type ExtractOptions struct {
	OverwriteExisting bool
	MkdirAll          bool
	IgnoreError       bool
	FilenameEncoding  string
	Password          string
	PassworldCallback func() string
	ProgressBar       io.Writer
	width             int
}

// UpdateWidth todo
func (es *ExtractOptions) UpdateWidth() {
	w, _, err := terminal.GetSize(int(os.Stdout.Fd()))
	if err == nil {
		es.width = w
		return
	}
	es.width = 80
}

// OnEntry on entry
func (es *ExtractOptions) OnEntry(name string) {
	if es == nil || es.ProgressBar == nil {
		return
	}
	n := runewidth.StringWidth(name)
	if n+8 <= es.width {
		fmt.Fprintf(es.ProgressBar, "\x1b[2K\r\x1b[33mx %s\x1b[0m", name)
		return
	}
	basename := filepath.Base(name)
	if n = runewidth.StringWidth(basename); n+8 <= es.width {
		fmt.Fprintf(es.ProgressBar, "\x1b[2K\r\x1b[33mx .../%s\x1b[0m", basename)
		return
	}
	fmt.Fprintf(es.ProgressBar, "\x1b[2K\r\x1b[33mx ...%s\x1b[0m", basename[len(basename)-es.width+8:])
}

// Compression Algorithm
const (
	None = iota
	Deflate
	GZ
	BZip2
	LZMA
	XZ
	LZ4
	Brotli
	Zstandard
)

// Error define
var (
	ErrRelativePathEscape = errors.New("relative path escape")
	ErrResponseFilesField = errors.New("response files field error")
)
