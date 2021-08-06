// +build !windows

package base

import "os"

func FixupTerminal() {
	//
}

//MoveFile file
func MoveFile(oldpath, newpath string) error {
	if _, err := os.Stat(newpath); err == nil {
		if err := os.Remove(newpath); err != nil {
			return err
		}
	}
	return os.Rename(oldpath, newpath)
}
