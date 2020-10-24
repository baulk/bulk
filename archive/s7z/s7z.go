package s7z

import (
	"io"
	"os"
	"path/filepath"

	"github.com/baulk/bulk/archive/foundation"
	"github.com/baulk/bulk/base"
	"github.com/baulk/bulk/go7z"
)

// Extractor type
type Extractor struct {
	fd  *os.File
	szr *go7z.Reader
	es  *foundation.ExtractOptions
}

// Matched magic
func Matched(buf []byte) bool {
	return len(buf) > 5 &&
		buf[0] == 0x37 && buf[1] == 0x7A && buf[2] == 0xBC &&
		buf[3] == 0xAF && buf[4] == 0x27 && buf[5] == 0x1C
}

//NewExtractor new tar extractor
func NewExtractor(fd *os.File, es *foundation.ExtractOptions) (*Extractor, error) {
	st, err := fd.Stat()
	if err != nil {
		fd.Close()
		return nil, err
	}
	r, err := go7z.NewReader(fd, st.Size())
	if err != nil {
		fd.Close()
		return nil, err
	}
	e := &Extractor{szr: r, fd: fd, es: es}
	e.szr.Options.SetPassword(es.Password)
	e.szr.Options.SetPasswordCallback(es.PassworldCallback)
	return e, nil
}

// Close fd
func (e *Extractor) Close() error {
	return e.fd.Close()
}

func (e *Extractor) extractFile(p string) error {
	if foundation.PathIsExists(p) {
		if !e.es.OverwriteExisting {
			return base.ErrorCat("file already exists: ", p)
		}
	}
	return base.SaveFile(e.szr, p, 0664, true)
}

// Extract file
func (e *Extractor) Extract(destination string) error {
	for {
		hdr, err := e.szr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		p := filepath.Join(destination, hdr.Name)
		if !foundation.IsRelativePath(destination, p) {
			if e.es.IgnoreError {
				continue
			}
			return foundation.ErrRelativePathEscape
		}
		//name := path.Clean(hdr.Name)
		if hdr.IsEmptyStream && !hdr.IsEmptyFile {
			if err := os.MkdirAll(p, 0775); err != nil {
				if !e.es.IgnoreError {
					return err
				}
			}
			continue
		}
		if err := e.extractFile(p); err != nil {
			if !e.es.IgnoreError {
				return err
			}
		}
	}
	return nil
}
