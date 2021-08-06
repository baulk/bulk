package zip

import (
	"errors"
	"io"
	"io/ioutil"
	"sync"

	"github.com/dsnet/compress/bzip2"
	"github.com/klauspost/compress/zip"
	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
	"github.com/ulikunitz/xz/lzma"
)

func newBzip2Reader(r io.Reader) io.ReadCloser {
	br, err := bzip2.NewReader(r, nil)
	if err != nil {
		return nil
	}
	return ioutil.NopCloser(br)
}

// support xz
func newXzReader(r io.Reader) io.ReadCloser {
	xr, err := xz.NewReader(r)
	if err != nil {
		return nil
	}
	return ioutil.NopCloser(xr)
}

func newLzmaReader(r io.Reader) io.ReadCloser {
	lr, err := lzma.NewReader(r)
	if err != nil {
		return nil
	}
	return ioutil.NopCloser(lr)
}

var zstdReaderPool sync.Pool

type pooledZstdReader struct {
	mu sync.Mutex // guards Close and Read
	zr io.ReadCloser
}

func (r *pooledZstdReader) Read(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.zr == nil {
		return 0, errors.New("Read after Close")
	}
	return r.zr.Read(p)
}

func (r *pooledZstdReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	var err error
	if r.zr != nil {
		err = r.zr.Close()
		zstdReaderPool.Put(r.zr)
		r.zr = nil
	}
	return err
}

// Resetter resets a ReadCloser returned by NewReader or NewReaderDict
// to switch to a new underlying Reader. This permits reusing a ReadCloser
// instead of allocating a new one.
type Resetter interface {
	// Reset discards any buffered data and resets the Resetter as if it was
	// newly initialized with the given reader.
	Reset(r io.Reader) error
}

// In order to improve the decompression efficiency of zstd, here we take the zstd decoder
func newZstdReader(r io.Reader) io.ReadCloser {
	zr, ok := zstdReaderPool.Get().(io.ReadCloser)
	if ok {
		zr.(Resetter).Reset(r)
	} else {
		ze, err := zstd.NewReader(r)
		if err != nil {
			return nil
		}
		zr = ze.IOReadCloser()
	}
	return &pooledZstdReader{zr: zr}
}

func zipRegisterDecompressor(zr *zip.Reader) {
	zr.RegisterDecompressor(uint16(BZIP2), newBzip2Reader)
	zr.RegisterDecompressor(uint16(XZ), newXzReader)
	zr.RegisterDecompressor(uint16(ZSTD), newZstdReader)
}
