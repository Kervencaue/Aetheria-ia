package wallet

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aetheria/blockchain/pkg/crypto"
)

// Wallet represents a user's wallet
type Wallet struct {
	Address    string `json:"address"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

// NewWallet creates a new wallet
func NewWallet() (*Wallet, error) {
	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	return &Wallet{
		Address:    keyPair.Address(),
		PublicKey:  crypto.PublicKeyToHex(keyPair.PublicKey),
		PrivateKey: crypto.PrivateKeyToHex(keyPair.PrivateKey),
	}, nil
}

// SaveToFile saves wallet to a JSON file
func (w *Wallet) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal wallet: %w", err)
	}

	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write wallet file: %w", err)
	}

	return nil
}

// LoadFromFile loads wallet from a JSON file
func LoadFromFile(filename string) (*Wallet, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read wallet file: %w", err)
	}

	var wallet Wallet
	if err := json.Unmarshal(data, &wallet); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wallet: %w", err)
	}

	return &wallet, nil
}

// GetKeyPair returns the key pair for this wallet
func (w *Wallet) GetKeyPair() (*crypto.KeyPair, error) {
	publicKey, err := crypto.PublicKeyFromHex(w.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}

	privateKey, err := crypto.PrivateKeyFromHex(w.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	return &crypto.KeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}, nil
}

// WalletInfo represents public wallet information
type WalletInfo struct {
	Address   string `json:"address"`
	PublicKey string `json:"public_key"`
}

// GetInfo returns public wallet information
func (w *Wallet) GetInfo() *WalletInfo {
	return &WalletInfo{
		Address:   w.Address,
		PublicKey: w.PublicKey,
	}
}
