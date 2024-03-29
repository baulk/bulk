package zip

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/baulk/bulk/modules/archive/foundation"
	"github.com/baulk/bulk/modules/base"
	"github.com/klauspost/compress/zip"
	"golang.org/x/text/encoding"
)

// CompressionMethod compress method see https://pkware.cachefly.net/webdocs/casestudies/APPNOTE.TXT
type CompressionMethod uint16

// CompressionMethod
// value
const (
	Store   CompressionMethod = 0
	Deflate CompressionMethod = 8
	BZIP2   CompressionMethod = 12
	LZMA    CompressionMethod = 14
	LZMA2   CompressionMethod = 33
	ZSTD    CompressionMethod = 93
	XZ      CompressionMethod = 95
	JPEG    CompressionMethod = 96
	WavPack CompressionMethod = 97
	PPMd    CompressionMethod = 98
	AES     CompressionMethod = 99
)

// Matched magic
func Matched(buf []byte) bool {
	return (len(buf) > 3 && buf[0] == 0x50 && buf[1] == 0x4B &&
		(buf[2] == 0x3 || buf[2] == 0x5 || buf[2] == 0x7) &&
		(buf[3] == 0x4 || buf[3] == 0x6 || buf[3] == 0x8))
}

// Extractor todo
type Extractor struct {
	fd                    *os.File
	zr                    *zip.Reader
	dec                   *encoding.Decoder
	eo                    *foundation.ExtractOptions
	compressedSizeTotal   uint64
	uncompressedSizeTotal uint64
}

// NewExtractor new extractor
func NewExtractor(fd *os.File, eo *foundation.ExtractOptions) (*Extractor, error) {
	st, err := fd.Stat()
	if err != nil {
		fd.Close()
		return nil, err
	}
	zr, err := zip.NewReader(fd, st.Size())
	if err != nil {
		fd.Close()
		return nil, err
	}
	zipRegisterDecompressor(zr)
	e := &Extractor{fd: fd, zr: zr, eo: eo}
	if ens := os.Getenv("ZIP_ENCODING"); len(ens) != 0 {
		eo.FilenameEncoding = ens
	}
	return e, nil
}

// Close fd
func (e *Extractor) Close() error {
	return e.fd.Close()
}

func (e *Extractor) extractSymlink(p, destination string, zf *zip.File) error {
	r, err := zf.Open()
	if err != nil {
		return err
	}
	defer r.Close()
	lnk, err := ioutil.ReadAll(io.LimitReader(r, 32678))
	if err != nil {
		return err
	}
	lnkp := strings.TrimSpace(string(lnk))
	if filepath.IsAbs(lnkp) {
		return foundation.SymbolicLink(filepath.Clean(lnkp), p)
	}
	oldname := filepath.Join(filepath.Dir(p), lnkp)
	return foundation.SymbolicLink(oldname, p)
}

func (e *Extractor) extractFile(p, destination string, zf *zip.File) error {
	if foundation.PathIsExists(p) {
		if !e.eo.OverwriteExisting {
			return base.ErrorCat("file already exists: ", p)
		}
	}
	r, err := zf.Open()
	if err != nil {
		return err
	}
	defer r.Close()
	return base.SaveFile(r, p, zf.FileHeader.Mode(), true)
}

// Statistics todo
func (e *Extractor) Statistics() error {
	for _, file := range e.zr.File {
		e.uncompressedSizeTotal += file.UncompressedSize64
		e.compressedSizeTotal += file.CompressedSize64
	}
	return nil
}

// Extract file
func (e *Extractor) Extract(destination string) error {
	for _, file := range e.zr.File {
		name := e.DecodeFileName(file.FileHeader)
		p := filepath.Join(destination, name)
		if !foundation.IsRelativePath(destination, p) {
			if e.eo.IgnoreError {
				continue
			}
			return foundation.ErrRelativePathEscape
		}
		fi := file.FileInfo()
		if fi.IsDir() {
			if err := os.MkdirAll(p, fi.Mode()); err != nil {
				if !e.eo.IgnoreError {
					return err
				}
			}
			continue
		}
		e.eo.OnEntry(name)
		if fi.Mode()&os.ModeSymlink != 0 {
			if err := e.extractSymlink(p, destination, file); err != nil {
				if !e.eo.IgnoreError {
					return err
				}
			}
			continue
		}
		if err := e.extractFile(p, destination, file); err != nil {
			if !e.eo.IgnoreError {
				return err
			}
		}
	}
	return nil
}

// new archive plase call zipRegisterCompressor
