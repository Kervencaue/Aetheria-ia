package crypto

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
)

// Sign signs data with private key
func Sign(privateKey ed25519.PrivateKey, data []byte) []byte {
	return ed25519.Sign(privateKey, data)
}

// Verify verifies signature with public key
func Verify(publicKey ed25519.PublicKey, data []byte, signature []byte) bool {
	return ed25519.Verify(publicKey, data, signature)
}

// SignatureToHex encodes signature to hex string
func SignatureToHex(signature []byte) string {
	return hex.EncodeToString(signature)
}

// SignatureFromHex decodes hex-encoded signature
func SignatureFromHex(hexSig string) ([]byte, error) {
	signature, err := hex.DecodeString(hexSig)
	if err != nil {
		return nil, fmt.Errorf("invalid hex signature: %w", err)
	}
	if len(signature) != ed25519.SignatureSize {
		return nil, fmt.Errorf("invalid signature size: %d", len(signature))
	}
	return signature, nil
}
