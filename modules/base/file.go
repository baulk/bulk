package base

import (
	"io"
	"os"
	"path/filepath"
)

// SaveFile save file to filesystem
func SaveFile(in io.Reader, path string, fm os.FileMode, replace bool) error {
	if _, err := os.Stat(path); err == nil {
		if replace {
			return os.ErrExist
		}
		os.Remove(path)
	} else {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
	}
	fd, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fm)
	if err != nil {
		return err
	}
	defer fd.Close()
	if _, err := io.Copy(fd, in); err != nil {
		return err
	}
	return nil
}
