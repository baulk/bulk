package netutils

import (
	"crypto/sha256"
	"crypto/sha512"
	"hash"
	"strings"

	"github.com/zeebo/blake3"
	"golang.org/x/crypto/sha3"
)

// HashComparator hashCompare
type HashComparator struct {
	H hash.Hash
	S string
}

// NewHashComparator todo
func NewHashComparator(hsx string) *HashComparator {
	if hsx == "" {
		return nil
	}
	hsx = strings.ToLower(hsx)
	if strings.HasPrefix(hsx, "blake3:") {
		return &HashComparator{H: blake3.New(), S: strings.TrimPrefix(hsx, "blake3:")}
	}
	if strings.HasPrefix(hsx, "sha256:") {
		return &HashComparator{H: sha256.New(), S: strings.TrimPrefix(hsx, "sha256:")}
	}
	if strings.HasPrefix(hsx, "sha512:") {
		return &HashComparator{H: sha512.New(), S: strings.TrimPrefix(hsx, "sha256:")}
	}
	if strings.HasPrefix(hsx, "sha3-512:") {
		return &HashComparator{H: sha3.New512(), S: strings.TrimPrefix(hsx, "sha3-512:")}
	}
	if strings.HasPrefix(hsx, "sha3-256:") {
		return &HashComparator{H: sha3.New256(), S: strings.TrimPrefix(hsx, "sha3-512:")}
	}
	return &HashComparator{H: sha256.New(), S: strings.TrimPrefix(hsx, "sha256:")}
}
