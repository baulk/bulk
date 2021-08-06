package zip

import (
	"errors"
	"io"
	"sync"

	"github.com/dsnet/compress/bzip2"
	"github.com/klauspost/compress/zip"
	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
)

var zstdWriterPool sync.Pool

func newZstdWriter(w io.Writer) (io.WriteCloser, error) {
	zw, ok := zstdWriterPool.Get().(*zstd.Encoder)
	if ok {
		zw.Reset(w)
	} else {
		zw, _ = zstd.NewWriter(w)
	}
	return &pooledZstdWriter{zw: zw}, nil
}

type pooledZstdWriter struct {
	mu sync.Mutex // guards Close and Write
	zw *zstd.Encoder
}

func (w *pooledZstdWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.zw == nil {
		return 0, errors.New("Write after Close")
	}
	return w.zw.Write(p)
}

func (w *pooledZstdWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	var err error
	if w.zw != nil {
		err = w.zw.Close()
		zstdWriterPool.Put(w.zw)
		w.zw = nil
	}
	return err
}

func zipRegisterCompressor(zw *zip.Writer) {
	zw.RegisterCompressor(uint16(BZIP2), func(w io.Writer) (io.WriteCloser, error) {
		return bzip2.NewWriter(w, nil)
	})
	zw.RegisterCompressor(uint16(XZ), func(w io.Writer) (io.WriteCloser, error) {
		return xz.NewWriter(w)
	})
	zw.RegisterCompressor(uint16(ZSTD), newZstdWriter)
}
