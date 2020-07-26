package basics

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// IsRelativePath todo
func IsRelativePath(basepath, childpath string) bool {
	childpath += string(os.PathSeparator)
	return strings.HasPrefix(childpath, basepath)
}

// SymbolicLink todo
func SymbolicLink(oldname string, newname string) error {
	if err := os.MkdirAll(filepath.Dir(newname), 0755); err != nil {
		return fmt.Errorf("%s: making directory for file: %v", newname, err)
	}

	if _, err := os.Lstat(newname); err == nil {
		if err = os.Remove(newname); err != nil {
			return fmt.Errorf("%s: failed to unlink: %+v", newname, err)
		}
	}

	if err := os.Symlink(oldname, newname); err != nil {
		return fmt.Errorf("%s: making symbolic link for: %v", newname, err)
	}
	return nil
}

// HardLink todo
func HardLink(oldname string, newname string) error {
	if err := os.MkdirAll(filepath.Dir(newname), 0755); err != nil {
		return fmt.Errorf("%s: making directory for file: %v", newname, err)
	}

	if _, err := os.Lstat(newname); err == nil {
		if err = os.Remove(newname); err != nil {
			return fmt.Errorf("%s: failed to unlink: %+v", newname, err)
		}
	}

	if err := os.Link(oldname, newname); err != nil {
		return fmt.Errorf("%s: making hard link for: %v", newname, err)
	}
	return nil
}

// PathIsExists todo
func PathIsExists(p string) bool {
	if _, err := os.Stat(p); err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
