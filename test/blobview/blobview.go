package main

import (
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"strconv"
)

// git blob view

// BlobViewer todo
type BlobViewer struct {
	w  io.Writer
	zr io.ReadCloser
	fd *os.File
	h  hash.Hash
}

// NewBlobViewer new blob viewer
func NewBlobViewer(p string, w io.Writer) (*BlobViewer, error) {
	fd, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	return &BlobViewer{fd: fd, w: w, h: sha1.New()}, nil
}

func (bv *BlobViewer) readUntil(delim byte) ([]byte, error) {
	var buf [1]byte
	value := make([]byte, 0, 16)
	for {
		if n, err := bv.zr.Read(buf[:]); err != nil && (err != io.EOF || n == 0) {
			if err == io.EOF {
				return nil, errors.New("invalid header")
			}
			return nil, err
		}
		bv.h.Write(buf[:])
		if buf[0] == delim {
			return value, nil
		}

		value = append(value, buf[0])
	}
}

// Header Lookup header
func (bv *BlobViewer) Header() error {
	raw, err := bv.readUntil(' ')
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "object type is: \x1b[34m%s\x1b[0m\n", raw)
	if raw, err = bv.readUntil(0); err != nil {
		return err
	}
	size, err := strconv.ParseInt(string(raw), 10, 64)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "object size: \x1b[34m%d\x1b[0m\n", size)
	return nil
}

//Lookup blob details
func (bv *BlobViewer) Lookup() error {
	zr, err := zlib.NewReader(bv.fd)
	if err != nil {
		return err
	}
	bv.zr = zr
	bv.Header()
	mw := io.MultiWriter(bv.w, bv.h)
	_, _ = io.Copy(mw, bv.zr)
	return nil
}

// Close close blob viewer
func (bv *BlobViewer) Close() error {
	if bv.zr != nil {
		bv.zr.Close()
	}
	fmt.Fprintf(os.Stderr, "Hash: %s\n", hex.EncodeToString(bv.h.Sum(nil)))
	if bv.fd != nil {
		return bv.fd.Close()
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s blob-id\n", os.Args[0])
		os.Exit(1)
	}
	bv, err := NewBlobViewer(os.Args[1], os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open: %s error %v\n", os.Args[1], err)
		os.Exit(1)
	}
	defer bv.Close()
	if err := bv.Lookup(); err != nil {
		fmt.Fprintf(os.Stderr, "lookup: %s error %v\n", os.Args[1], err)
		os.Exit(1)
	}
}
