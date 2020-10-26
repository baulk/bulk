package netutils

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"

	"github.com/zeebo/blake3"
	"golang.org/x/crypto/sha3"
)

// Verifier hash Verifier
type Verifier struct {
	H hash.Hash
	S string
}

// NewSumizer new
func NewSumizer(algorithm string) *Verifier {
	algorithm = strings.ToLower(algorithm)
	switch algorithm {
	case "blake3":
		return &Verifier{H: blake3.New()}
	case "sha256":
		return &Verifier{H: sha256.New()}
	case "sha512":
		return &Verifier{H: sha512.New()}
	case "sha3-256":
		return &Verifier{H: sha3.New256()}
	case "sha3-512":
		return &Verifier{H: sha3.New512()}
	default:
	}
	return &Verifier{H: sha256.New()}
}

// NewVerifier todo
func NewVerifier(eu *DownloadEntity) *Verifier {
	if eu.HashValue != "" {
		var algorithm, checksum string
		if i := strings.IndexByte(eu.HashValue, ':'); i != -1 {
			algorithm = eu.HashValue[0:i]
			checksum = eu.HashValue[i+1:]
		} else {
			checksum = eu.HashValue
		}
		verifier := NewSumizer(algorithm)
		verifier.S = checksum
		return verifier
	}
	if eu.Algorithm == "" {
		return nil
	}
	return NewSumizer(eu.Algorithm)
}

// IsMatch hash is matched
func (hc *Verifier) IsMatch(filename string) error {
	checksum := hex.EncodeToString(hc.H.Sum(nil))
	if hc.S == "" {
		fmt.Fprintf(os.Stderr, "\x1b[34m%s  %s\x1b[0m\n", checksum, filename)
		return nil
	}
	if checksum != hc.S {
		return fmt.Errorf("The calculated hash value %s is different from %s", checksum, hc.S)
	}
	return nil
}

// FileHashEqual file exists
func FileHashEqual(fullname string, eu *DownloadEntity) bool {
	if eu.HashValue == "" {
		return false
	}
	if _, err := os.Stat(fullname); err != nil {
		return false
	}
	hc := NewVerifier(eu)
	if hc == nil {
		return false
	}
	fd, err := os.Open(fullname)
	if err != nil {
		return false
	}
	if _, err := io.Copy(hc.H, fd); err != nil && err != io.EOF {
		return false
	}
	if hsx2 := hex.EncodeToString(hc.H.Sum(nil)); hsx2 != hc.S {
		return false
	}
	return true
}

// IsHashNotMatch is hash not match
func IsHashNotMatch(err error) bool {
	return strings.HasPrefix(err.Error(), "The calculated hash value ")
}
