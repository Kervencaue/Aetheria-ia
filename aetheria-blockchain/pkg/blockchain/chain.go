package blockchain

import (
	"fmt"
	"sync"
)

const (
	// BlockReward is the reward for creating a block (in Aetheria tokens)
	BlockReward = 50
	// MinStakeAmount is the minimum amount required to become a validator
	MinStakeAmount = 1000
)

// Blockchain represents the blockchain
type Blockchain struct {
	Blocks            []*Block
	PendingTxs        []*Transaction
	State             *State
	GenesisAddress    string
	mu                sync.RWMutex
	txPool            map[string]*Transaction
}

// NewBlockchain creates a new blockchain with genesis block
func NewBlockchain(genesisAddress string, initialSupply uint64) *Blockchain {
	bc := &Blockchain{
		Blocks:         make([]*Block, 0),
		PendingTxs:     make([]*Transaction, 0),
		State:          NewState(),
		GenesisAddress: genesisAddress,
		txPool:         make(map[string]*Transaction),
	}

	// Create genesis block
	genesis := bc.createGenesisBlock(genesisAddress, initialSupply)
	bc.Blocks = append(bc.Blocks, genesis)
	bc.State.ApplyBlock(genesis)

	return bc
}

// createGenesisBlock creates the first block in the chain
func (bc *Blockchain) createGenesisBlock(address string, initialSupply uint64) *Block {
	// Create coinbase transaction for initial supply
	coinbase := &Transaction{
		From:      "",
		To:        address,
		Amount:    initialSupply,
		Fee:       0,
		Timestamp: 0,
	}
	coinbase.ID = coinbase.calculateID()

	genesis := &Block{
		Index:        0,
		Timestamp:    0,
		Transactions: []*Transaction{coinbase},
		PrevHash:     "0",
		Validator:    "genesis",
	}
	genesis.Hash = genesis.calculateHash()
	genesis.Signature = "genesis"

	return genesis
}

// GetLatestBlock returns the last block in the chain
func (bc *Blockchain) GetLatestBlock() *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	
	if len(bc.Blocks) == 0 {
		return nil
	}
	return bc.Blocks[len(bc.Blocks)-1]
}

// AddBlock adds a new block to the chain
func (bc *Blockchain) AddBlock(block *Block) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Validate block
	if err := bc.validateBlock(block); err != nil {
		return fmt.Errorf("invalid block: %w", err)
	}

	// Apply block to state
	tempState := bc.State.Clone()
	if err := tempState.ApplyBlock(block); err != nil {
		return fmt.Errorf("failed to apply block: %w", err)
	}

	// Add block to chain
	bc.Blocks = append(bc.Blocks, block)
	bc.State = tempState

	// Remove transactions from pool
	for _, tx := range block.Transactions {
		delete(bc.txPool, tx.ID)
	}

	// Remove from pending
	bc.PendingTxs = make([]*Transaction, 0)

	return nil
}

// validateBlock validates a block before adding it to the chain
func (bc *Blockchain) validateBlock(block *Block) error {
	latest := bc.GetLatestBlock()
	
	// Check index
	if block.Index != latest.Index+1 {
		return fmt.Errorf("invalid block index: expected %d, got %d", latest.Index+1, block.Index)
	}

	// Check previous hash
	if block.PrevHash != latest.Hash {
		return fmt.Errorf("invalid previous hash")
	}

	// Check hash
	expectedHash := block.calculateHash()
	if block.Hash != expectedHash {
		return fmt.Errorf("invalid block hash")
	}

	// Verify all transactions
	for _, tx := range block.Transactions {
		if !tx.IsCoinbase() {
			if err := tx.Verify(); err != nil {
				return fmt.Errorf("invalid transaction %s: %w", tx.ID, err)
			}
		}
	}

	return nil
}

// AddTransaction adds a transaction to the pending pool
func (bc *Blockchain) AddTransaction(tx *Transaction) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Verify transaction
	if err := tx.Verify(); err != nil {
		return fmt.Errorf("invalid transaction: %w", err)
	}

	// Check if transaction already exists
	if _, exists := bc.txPool[tx.ID]; exists {
		return fmt.Errorf("transaction already exists")
	}

	// Check balance
	balance := bc.State.GetBalance(tx.From)
	totalRequired := tx.Amount + tx.Fee
	if balance < totalRequired {
		return fmt.Errorf("insufficient balance: has %d, needs %d", balance, totalRequired)
	}

	// Add to pool
	bc.txPool[tx.ID] = tx
	bc.PendingTxs = append(bc.PendingTxs, tx)

	return nil
}

// CreateBlock creates a new block with pending transactions
func (bc *Blockchain) CreateBlock(validator string) *Block {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	latest := bc.GetLatestBlock()
	
	// Create coinbase transaction for block reward
	coinbase := &Transaction{
		From:      "",
		To:        validator,
		Amount:    BlockReward,
		Fee:       0,
		Timestamp: 0,
	}
	coinbase.ID = coinbase.calculateID()

	// Add pending transactions
	transactions := []*Transaction{coinbase}
	transactions = append(transactions, bc.PendingTxs...)

	// Create block
	block := NewBlock(latest.Index+1, transactions, latest.Hash, validator)
	
	return block
}

// GetBlock returns a block by index
func (bc *Blockchain) GetBlock(index uint64) *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if index >= uint64(len(bc.Blocks)) {
		return nil
	}
	return bc.Blocks[index]
}

// GetBlockByHash returns a block by hash
func (bc *Blockchain) GetBlockByHash(hash string) *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	for _, block := range bc.Blocks {
		if block.Hash == hash {
			return block
		}
	}
	return nil
}

// GetTransaction returns a transaction by ID
func (bc *Blockchain) GetTransaction(txID string) *Transaction {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	// Check in blocks
	for _, block := range bc.Blocks {
		if tx := block.GetTransactionByID(txID); tx != nil {
			return tx
		}
	}

	// Check in pool
	if tx, exists := bc.txPool[txID]; exists {
		return tx
	}

	return nil
}

// Height returns the current blockchain height
func (bc *Blockchain) Height() uint64 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return uint64(len(bc.Blocks))
}

// IsValid validates the entire blockchain
func (bc *Blockchain) IsValid() bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	for i := 1; i < len(bc.Blocks); i++ {
		currentBlock := bc.Blocks[i]
		prevBlock := bc.Blocks[i-1]

		// Check hash
		if currentBlock.Hash != currentBlock.calculateHash() {
			return false
		}

		// Check previous hash
		if currentBlock.PrevHash != prevBlock.Hash {
			return false
		}

		// Check index
		if currentBlock.Index != prevBlock.Index+1 {
			return false
		}
	}

	return true
}
