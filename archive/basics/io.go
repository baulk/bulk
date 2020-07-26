package basics

import (
	"io"
	"os"
	"path/filepath"
)

// fs.go

// WriteDisk write to disk file
func WriteDisk(in io.Reader, path string, fm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	// Why not use OpenFile, if file always exists. cannot chmod
	// Windows chmod only can set readonly
	if err := os.Chmod(path, fm); err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}
