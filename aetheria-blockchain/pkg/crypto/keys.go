package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// KeyPair represents a public/private key pair
type KeyPair struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
}

// GenerateKeyPair creates a new ED25519 key pair
func GenerateKeyPair() (*KeyPair, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	return &KeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}

// PublicKeyToAddress converts public key to address (hex-encoded hash)
func PublicKeyToAddress(publicKey ed25519.PublicKey) string {
	hash := Hash(publicKey)
	return hex.EncodeToString(hash[:20]) // Use first 20 bytes
}

// Address returns the address for this key pair
func (kp *KeyPair) Address() string {
	return PublicKeyToAddress(kp.PublicKey)
}

// PublicKeyFromHex decodes a hex-encoded public key
func PublicKeyFromHex(hexKey string) (ed25519.PublicKey, error) {
	bytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("invalid hex public key: %w", err)
	}
	if len(bytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: %d", len(bytes))
	}
	return ed25519.PublicKey(bytes), nil
}

// PrivateKeyFromHex decodes a hex-encoded private key
func PrivateKeyFromHex(hexKey string) (ed25519.PrivateKey, error) {
	bytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("invalid hex private key: %w", err)
	}
	if len(bytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: %d", len(bytes))
	}
	return ed25519.PrivateKey(bytes), nil
}

// PublicKeyToHex encodes public key to hex string
func PublicKeyToHex(publicKey ed25519.PublicKey) string {
	return hex.EncodeToString(publicKey)
}

// PrivateKeyToHex encodes private key to hex string
func PrivateKeyToHex(privateKey ed25519.PrivateKey) string {
	return hex.EncodeToString(privateKey)
}
