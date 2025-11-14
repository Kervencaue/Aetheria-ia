package blockchain

import (
	"fmt"
	"sync"
)

// State represents the global state of the blockchain
type State struct {
	Balances map[string]uint64 `json:"balances"` // address -> balance
	Stakes   map[string]uint64 `json:"stakes"`   // address -> staked amount
	mu       sync.RWMutex
}

// NewState creates a new state
func NewState() *State {
	return &State{
		Balances: make(map[string]uint64),
		Stakes:   make(map[string]uint64),
	}
}

// GetBalance returns the balance of an address
func (s *State) GetBalance(address string) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Balances[address]
}

// GetStake returns the staked amount of an address
func (s *State) GetStake(address string) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Stakes[address]
}

// SetBalance sets the balance of an address
func (s *State) SetBalance(address string, amount uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Balances[address] = amount
}

// AddBalance adds to the balance of an address
func (s *State) AddBalance(address string, amount uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Balances[address] += amount
}

// SubBalance subtracts from the balance of an address
func (s *State) SubBalance(address string, amount uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.Balances[address] < amount {
		return fmt.Errorf("insufficient balance")
	}
	s.Balances[address] -= amount
	return nil
}

// AddStake adds to the staked amount of an address
func (s *State) AddStake(address string, amount uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.Balances[address] < amount {
		return fmt.Errorf("insufficient balance to stake")
	}
	
	s.Balances[address] -= amount
	s.Stakes[address] += amount
	return nil
}

// RemoveStake removes from the staked amount of an address
func (s *State) RemoveStake(address string, amount uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.Stakes[address] < amount {
		return fmt.Errorf("insufficient stake")
	}
	
	s.Stakes[address] -= amount
	s.Balances[address] += amount
	return nil
}

// ApplyTransaction applies a transaction to the state
func (s *State) ApplyTransaction(tx *Transaction) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Coinbase transaction (mining reward)
	if tx.IsCoinbase() {
		s.Balances[tx.To] += tx.Amount
		return nil
	}

	// Check balance
	totalRequired := tx.Amount + tx.Fee
	if s.Balances[tx.From] < totalRequired {
		return fmt.Errorf("insufficient balance: has %d, needs %d", s.Balances[tx.From], totalRequired)
	}

	// Apply transaction
	s.Balances[tx.From] -= totalRequired
	s.Balances[tx.To] += tx.Amount

	return nil
}

// ApplyBlock applies all transactions in a block to the state
func (s *State) ApplyBlock(block *Block) error {
	for _, tx := range block.Transactions {
		if err := s.ApplyTransaction(tx); err != nil {
			return fmt.Errorf("failed to apply transaction %s: %w", tx.ID, err)
		}
	}
	
	// Add fees to validator
	fees := block.TotalFees()
	if fees > 0 {
		s.AddBalance(block.Validator, fees)
	}
	
	return nil
}

// Clone creates a copy of the state
func (s *State) Clone() *State {
	s.mu.RLock()
	defer s.mu.RUnlock()

	newState := NewState()
	for addr, balance := range s.Balances {
		newState.Balances[addr] = balance
	}
	for addr, stake := range s.Stakes {
		newState.Stakes[addr] = stake
	}
	return newState
}

// TotalStaked returns the total amount staked in the network
func (s *State) TotalStaked() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var total uint64
	for _, stake := range s.Stakes {
		total += stake
	}
	return total
}

// GetValidators returns all addresses with stake
func (s *State) GetValidators() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	validators := make([]string, 0, len(s.Stakes))
	for addr, stake := range s.Stakes {
		if stake > 0 {
			validators = append(validators, addr)
		}
	}
	return validators
}
