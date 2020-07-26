package basics

import (
	"errors"
	"os"
)

// ExtractSetting todo
type ExtractSetting struct {
	OverwriteExisting bool
	MkdirAll          bool
	IgnoreError       bool
	Encoding          string
	Password          string
	PassworldCallback func() string
	OnEntry           func(name string, fi os.FileInfo) error
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
