package base

import (
	"bytes"
	"errors"
	"io"
	"strings"
)

// StrCat cat strings:
// You should know that StrCat gradually builds advantages
// only when the number of parameters is> 2.
func StrCat(sv ...string) string {
	var sb strings.Builder
	var size int
	for _, s := range sv {
		size += len(s)
	}
	sb.Grow(size)
	for _, s := range sv {
		_, _ = sb.WriteString(s)
	}
	return sb.String()
}

// WriteFile write file
func WriteFile(out io.Writer, sv ...string) error {
	if len(sv) == 1 {
		_, err := out.Write([]byte(sv[0]))
		return err
	}
	var sb bytes.Buffer
	var size int
	for _, s := range sv {
		size += len(s)
	}
	sb.Grow(size)
	for _, s := range sv {
		_, _ = sb.WriteString(s)
	}
	if _, err := out.Write(sb.Bytes()); err != nil {
		return err
	}
	return nil
}

// ByteCat cat strings:
// You should know that StrCat gradually builds advantages
// only when the number of parameters is> 2.
func ByteCat(sv ...[]byte) string {
	var sb bytes.Buffer
	var size int
	for _, s := range sv {
		size += len(s)
	}
	sb.Grow(size)
	for _, s := range sv {
		_, _ = sb.Write(s)
	}
	return sb.String()
}

// ErrorCat todo
func ErrorCat(sv ...string) error {
	return errors.New(StrCat(sv...))
}
