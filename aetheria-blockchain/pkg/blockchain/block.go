package blockchain

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/aetheria/blockchain/pkg/crypto"
)

// Block represents a block in the blockchain
type Block struct {
	Index        uint64         `json:"index"`
	Timestamp    int64          `json:"timestamp"`
	Transactions []*Transaction `json:"transactions"`
	PrevHash     string         `json:"prev_hash"`
	Hash         string         `json:"hash"`
	Validator    string         `json:"validator"`
	Signature    string         `json:"signature"`
}

// NewBlock creates a new block
func NewBlock(index uint64, transactions []*Transaction, prevHash, validator string) *Block {
	block := &Block{
		Index:        index,
		Timestamp:    time.Now().Unix(),
		Transactions: transactions,
		PrevHash:     prevHash,
		Validator:    validator,
	}
	block.Hash = block.calculateHash()
	return block
}

// calculateHash calculates the hash of the block
func (b *Block) calculateHash() string {
	data := fmt.Sprintf("%d%d%s%s", b.Index, b.Timestamp, b.PrevHash, b.Validator)
	
	// Include all transaction hashes
	for _, tx := range b.Transactions {
		data += tx.ID
	}
	
	return crypto.HashString([]byte(data))
}

// Sign signs the block with validator's private key
func (b *Block) Sign(privateKey []byte) error {
	data := []byte(b.Hash)
	signature := crypto.Sign(privateKey, data)
	b.Signature = crypto.SignatureToHex(signature)
	return nil
}

// Verify verifies the block's integrity and signature
func (b *Block) Verify(publicKey []byte) error {
	// Verify hash
	expectedHash := b.calculateHash()
	if b.Hash != expectedHash {
		return fmt.Errorf("invalid block hash")
	}

	// Verify signature
	if b.Signature == "" {
		return fmt.Errorf("block not signed")
	}

	signature, err := crypto.SignatureFromHex(b.Signature)
	if err != nil {
		return fmt.Errorf("invalid signature: %w", err)
	}

	data := []byte(b.Hash)
	if !crypto.Verify(publicKey, data, signature) {
		return fmt.Errorf("invalid block signature")
	}

	// Verify all transactions
	for _, tx := range b.Transactions {
		if !tx.IsCoinbase() {
			if err := tx.Verify(); err != nil {
				return fmt.Errorf("invalid transaction %s: %w", tx.ID, err)
			}
		}
	}

	return nil
}

// Serialize serializes block to bytes
func (b *Block) Serialize() ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(b); err != nil {
		return nil, fmt.Errorf("failed to serialize block: %w", err)
	}
	return buffer.Bytes(), nil
}

// DeserializeBlock deserializes bytes to block
func DeserializeBlock(data []byte) (*Block, error) {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&block); err != nil {
		return nil, fmt.Errorf("failed to deserialize block: %w", err)
	}
	return &block, nil
}

// GetTransactionByID finds a transaction by ID
func (b *Block) GetTransactionByID(txID string) *Transaction {
	for _, tx := range b.Transactions {
		if tx.ID == txID {
			return tx
		}
	}
	return nil
}

// TotalFees calculates total transaction fees in the block
func (b *Block) TotalFees() uint64 {
	var total uint64
	for _, tx := range b.Transactions {
		if !tx.IsCoinbase() {
			total += tx.Fee
		}
	}
	return total
}

// HashBytes returns the hash as bytes
func (b *Block) HashBytes() []byte {
	hash, _ := hex.DecodeString(b.Hash)
	return hash
}
