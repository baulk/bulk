package tar

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/baulk/bulk/archive/basics"
	"github.com/baulk/bulk/base"
	"github.com/dsnet/compress/bzip2"
	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4/v3"
	"github.com/ulikunitz/xz"
)

// tar

// Extractor type
type Extractor struct {
	fd *os.File
	r  *tar.Reader
	es *basics.ExtractSetting
}

// Matched todo
func Matched(buf []byte) bool {
	return len(buf) > 261 &&
		buf[257] == 0x75 && buf[258] == 0x73 &&
		buf[259] == 0x74 && buf[260] == 0x61 &&
		buf[261] == 0x72
}

//NewExtractor new tar extractor
func NewExtractor(fd *os.File, es *basics.ExtractSetting) (*Extractor, error) {
	return &Extractor{r: tar.NewReader(fd), fd: fd, es: es}, nil
}

// Close fd
func (e *Extractor) Close() error {
	return e.fd.Close()
}

func (e *Extractor) extractSymlink(p, destination string, hdr *tar.Header) error {
	linkname := hdr.Linkname
	if filepath.IsAbs(linkname) {
		return basics.SymbolicLink(filepath.Clean(linkname), p)
	}
	oldname := filepath.Join(filepath.Dir(p), linkname)
	return basics.SymbolicLink(oldname, p)
}

func (e *Extractor) extractHardLink(p, destination string, hdr *tar.Header) error {
	linkname := hdr.Linkname
	if filepath.IsAbs(linkname) {
		return basics.HardLink(filepath.Clean(linkname), p)
	}
	oldname := filepath.Join(filepath.Dir(p), linkname)
	return basics.HardLink(oldname, p)
}

// Extract file
func (e *Extractor) Extract(destination string) error {
	for {
		hdr, err := e.r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		p := filepath.Join(destination, hdr.Name)
		if !basics.IsRelativePath(destination, p) {
			if e.es.IgnoreError {
				continue
			}
			return basics.ErrRelativePathEscape
		}
		if hdr.Typeflag != tar.TypeDir && basics.PathIsExists(p) && !e.es.OverwriteExisting {
			return base.ErrorCat("file already exists: ", p)
		}
		fi := hdr.FileInfo()
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(p, fi.Mode()); err != nil {
				if !e.es.IgnoreError {
					return err
				}
			}
		case tar.TypeReg, tar.TypeRegA, tar.TypeChar, tar.TypeBlock, tar.TypeFifo, tar.TypeGNUSparse:
			e.es.OnEntry(hdr.Name)
			if err := basics.WriteDisk(e.r, p, fi.Mode()); err != nil {
				if !e.es.IgnoreError {
					return err
				}
			}
		case tar.TypeSymlink:
			if err := e.extractSymlink(p, destination, hdr); err != nil {
				if !e.es.IgnoreError {
					return err
				}
			}
		case tar.TypeLink:
			if err := e.extractHardLink(p, destination, hdr); err != nil {
				if !e.es.IgnoreError {
					return err
				}
			}
		case tar.TypeXGlobalHeader:
		}
	}
	return nil
}

// BrewingExtractor todo
type BrewingExtractor struct {
	extractor *Extractor
	mwr       io.ReadCloser
}

// MatchExtension todo
func MatchExtension(name string) int {
	if runtime.GOOS == "windows" {
		name = strings.ToLower(name)
	}
	if strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".tgz") {
		return basics.GZ
	}
	if strings.HasSuffix(name, ".tar.bz2") || strings.HasSuffix(name, ".tbz2") {
		return basics.BZip2
	}
	if strings.HasSuffix(name, ".tar.br") || strings.HasSuffix(name, ".tbr") {
		return basics.Brotli
	}
	if strings.HasSuffix(name, ".tar.zst") {
		return basics.Zstandard
	}
	if strings.HasSuffix(name, ".tar.xz") || strings.HasSuffix(name, ".txz") {
		return basics.XZ
	}
	if strings.HasSuffix(name, ".tar.lz4") || strings.HasSuffix(name, ".tlz4") {
		return basics.LZ4
	}
	return basics.None
}

// NewBrewingExtractor todo
func NewBrewingExtractor(fd *os.File, es *basics.ExtractSetting, alg int) (*BrewingExtractor, error) {
	var err error
	e := &BrewingExtractor{extractor: &Extractor{es: es}}
	switch alg {
	case basics.GZ:
		e.mwr, err = gzip.NewReader(fd)
		if err != nil {
			fd.Close()
			return nil, err
		}
	case basics.LZ4:
		e.mwr = ioutil.NopCloser(lz4.NewReader(fd))
	case basics.Brotli:
		e.mwr = ioutil.NopCloser(brotli.NewReader(fd))
	case basics.BZip2:
		e.mwr, err = bzip2.NewReader(fd, nil)
		if err != nil {
			fd.Close()
			return nil, err
		}
	case basics.XZ:
		r, err := xz.NewReader(fd)
		if err != nil {
			fd.Close()
			return nil, err
		}
		e.mwr = ioutil.NopCloser(r)
	case basics.Zstandard:
		dec, err := zstd.NewReader(fd)
		if err != nil {
			fd.Close()
			return nil, err
		}
		e.mwr = dec.IOReadCloser()
	default:
		fd.Close()
		return nil, fmt.Errorf("unsupport compress algorithm %d", alg)
	}
	e.extractor.fd = fd
	e.extractor.r = tar.NewReader(e.mwr)
	return e, nil
}

// Close fd
func (e *BrewingExtractor) Close() error {
	_ = e.mwr.Close()
	return e.extractor.Close()
}

// Extract file
func (e *BrewingExtractor) Extract(destination string) error {
	return e.extractor.Extract(destination)
}
