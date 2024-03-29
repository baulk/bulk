package archive

import (
	"errors"
	"io"
	"os"

	"github.com/baulk/bulk/modules/archive/foundation"
	"github.com/baulk/bulk/modules/archive/rar"
	"github.com/baulk/bulk/modules/archive/s7z"
	"github.com/baulk/bulk/modules/archive/tar"
	"github.com/baulk/bulk/modules/archive/zip"
)

func readMagic(fd *os.File) ([]byte, error) {
	buf := make([]byte, 0, 512)
	l, err := fd.Read(buf)
	if err != nil {
		return nil, err
	}
	if _, err := fd.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	return buf[0:l], nil
}

// NewExtractor todo
func NewExtractor(file string, es *foundation.ExtractOptions) (Extractor, error) {
	if es == nil {
		es = &foundation.ExtractOptions{}
	}

	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	mb, err := readMagic(fd)
	if err != nil {
		return nil, err
	}

	if zip.Matched(mb) {
		e, err := zip.NewExtractor(fd, es)
		if err != nil {
			return nil, err
		}
		return e, nil
	}
	if s7z.Matched(mb) {
		e, err := s7z.NewExtractor(fd, es)
		if err != nil {
			return nil, err
		}
		return e, nil
	}
	if rar.Matched(mb) {
		e, err := rar.NewExtractor(fd, es)
		if err != nil {
			return nil, err
		}
		return e, nil
	}
	if tar.Matched(mb) {
		e, _ := tar.NewExtractor(fd, es)
		return e, nil
	}
	if al := tar.MatchExtension(file); al != foundation.None {
		e, err := tar.NewBrewingExtractor(fd, es, al)
		if err != nil {
			return nil, err
		}
		return e, nil
	}
	return nil, errors.New("unsupport format")
}
