package blockchain

import (
	"bytes"
	"crypto/ed25519"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/aetheria/blockchain/pkg/crypto"
)

// Transaction represents a transfer of Aetheria tokens
type Transaction struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Amount    uint64    `json:"amount"`
	Fee       uint64    `json:"fee"`
	Timestamp int64     `json:"timestamp"`
	Signature string    `json:"signature"`
	PublicKey string    `json:"public_key"`
}

// NewTransaction creates a new transaction
func NewTransaction(from, to string, amount, fee uint64) *Transaction {
	tx := &Transaction{
		From:      from,
		To:        to,
		Amount:    amount,
		Fee:       fee,
		Timestamp: time.Now().Unix(),
	}
	tx.ID = tx.calculateID()
	return tx
}

// calculateID generates transaction ID from its data
func (tx *Transaction) calculateID() string {
	data := fmt.Sprintf("%s%s%d%d%d", tx.From, tx.To, tx.Amount, tx.Fee, tx.Timestamp)
	return crypto.HashString([]byte(data))
}

// Sign signs the transaction with private key
func (tx *Transaction) Sign(privateKey ed25519.PrivateKey) error {
	publicKey := privateKey.Public().(ed25519.PublicKey)
	tx.PublicKey = crypto.PublicKeyToHex(publicKey)
	
	data := tx.dataToSign()
	signature := crypto.Sign(privateKey, data)
	tx.Signature = crypto.SignatureToHex(signature)
	
	return nil
}

// Verify verifies the transaction signature
func (tx *Transaction) Verify() error {
	if tx.Signature == "" {
		return fmt.Errorf("transaction not signed")
	}

	publicKey, err := crypto.PublicKeyFromHex(tx.PublicKey)
	if err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}

	// Verify that From address matches public key
	expectedFrom := crypto.PublicKeyToAddress(publicKey)
	if tx.From != expectedFrom {
		return fmt.Errorf("from address does not match public key")
	}

	signature, err := crypto.SignatureFromHex(tx.Signature)
	if err != nil {
		return fmt.Errorf("invalid signature: %w", err)
	}

	data := tx.dataToSign()
	if !crypto.Verify(publicKey, data, signature) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

// dataToSign returns the data to be signed
func (tx *Transaction) dataToSign() []byte {
	data := fmt.Sprintf("%s%s%s%d%d%d", tx.ID, tx.From, tx.To, tx.Amount, tx.Fee, tx.Timestamp)
	return []byte(data)
}

// Serialize serializes transaction to bytes
func (tx *Transaction) Serialize() ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(tx); err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}
	return buffer.Bytes(), nil
}

// DeserializeTransaction deserializes bytes to transaction
func DeserializeTransaction(data []byte) (*Transaction, error) {
	var tx Transaction
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&tx); err != nil {
		return nil, fmt.Errorf("failed to deserialize transaction: %w", err)
	}
	return &tx, nil
}

// Hash returns the hash of the transaction
func (tx *Transaction) Hash() []byte {
	data, _ := tx.Serialize()
	return crypto.Hash(data)
}

// HashString returns hex-encoded hash
func (tx *Transaction) HashString() string {
	return hex.EncodeToString(tx.Hash())
}

// IsCoinbase checks if transaction is a coinbase (mining reward)
func (tx *Transaction) IsCoinbase() bool {
	return tx.From == ""
}
