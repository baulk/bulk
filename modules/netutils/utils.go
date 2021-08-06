package netutils

import (
	"path/filepath"
	"strings"
)

var complexExtensions = []string{
	".tar.gz",
	".tar.xz",
	".tar.sz",
	".tar.bz2",
	".tar.br",
	".tar.zst",
}

// StripExtension stripExtension
func StripExtension(name string) (string, string) {
	lowername := strings.ToLower(name)
	for _, s := range complexExtensions {
		if strings.HasSuffix(lowername, s) {
			return name[0 : len(name)-len(s)], s
		}
	}
	ext := filepath.Ext(name)
	return name[0 : len(name)-len(ext)], ext
}
