package hashing

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
)

// Hash has methods to hash strings
type Hash struct{}

// Make will give you a Hash struct to hash strings
func Make() Hash { return Hash{} }

// SHA1 will give you a SHA1 hash of a string
func (Hash) SHA1(src string) (string, error) {
	sha := sha1.New()
	_, error := io.WriteString(sha, src)
	hash := hex.EncodeToString(sha.Sum(nil))
	return hash, error
}
