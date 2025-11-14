package crypto

import (
	"crypto/sha256"
	"encoding/hex"
)

// Hash calculates SHA-256 hash of data
func Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// HashString returns hex-encoded hash
func HashString(data []byte) string {
	return hex.EncodeToString(Hash(data))
}

// DoubleHash performs double SHA-256 hashing
func DoubleHash(data []byte) []byte {
	first := sha256.Sum256(data)
	second := sha256.Sum256(first[:])
	return second[:]
}
