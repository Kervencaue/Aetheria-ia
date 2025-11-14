package consensus

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/aetheria/blockchain/pkg/blockchain"
)

// PoS implements Proof of Stake consensus
type PoS struct {
	ValidatorSet *ValidatorSet
	MinStake     uint64
	BlockTime    time.Duration
}

// NewPoS creates a new PoS consensus engine
func NewPoS(minStake uint64, blockTime time.Duration) *PoS {
	return &PoS{
		ValidatorSet: NewValidatorSet(),
		MinStake:     minStake,
		BlockTime:    blockTime,
	}
}

// SelectValidator selects a validator based on stake weight
// Uses weighted random selection where probability is proportional to stake
func (pos *PoS) SelectValidator(prevBlockHash string, timestamp int64) (*Validator, error) {
	validators := pos.ValidatorSet.GetValidators()
	if len(validators) == 0 {
		return nil, fmt.Errorf("no validators available")
	}

	// Filter validators with minimum stake
	eligibleValidators := make([]*Validator, 0)
	for _, v := range validators {
		if v.CanValidate(pos.MinStake) {
			eligibleValidators = append(eligibleValidators, v)
		}
	}

	if len(eligibleValidators) == 0 {
		return nil, fmt.Errorf("no eligible validators")
	}

	// Calculate total stake
	var totalStake uint64
	for _, v := range eligibleValidators {
		totalStake += v.Stake
	}

	// Generate deterministic random number based on previous block hash and timestamp
	seed := pos.generateSeed(prevBlockHash, timestamp)
	
	// Select validator using weighted random selection
	target := new(big.Int).Mod(seed, big.NewInt(int64(totalStake)))
	
	var cumulative uint64
	for _, v := range eligibleValidators {
		cumulative += v.Stake
		if target.Cmp(big.NewInt(int64(cumulative))) < 0 {
			return v, nil
		}
	}

	// Fallback to last validator (should not happen)
	return eligibleValidators[len(eligibleValidators)-1], nil
}

// generateSeed generates a deterministic seed for validator selection
func (pos *PoS) generateSeed(prevBlockHash string, timestamp int64) *big.Int {
	data := []byte(prevBlockHash)
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, uint64(timestamp))
	data = append(data, timeBytes...)
	
	hash := sha256.Sum256(data)
	seed := new(big.Int).SetBytes(hash[:])
	return seed
}

// ValidateBlock validates a block according to PoS rules
func (pos *PoS) ValidateBlock(block *blockchain.Block, prevBlock *blockchain.Block) error {
	// Check if validator exists and has minimum stake
	validator, err := pos.ValidatorSet.GetValidator(block.Validator)
	if err != nil {
		return fmt.Errorf("validator not found: %w", err)
	}

	if !validator.CanValidate(pos.MinStake) {
		return fmt.Errorf("validator does not have minimum stake")
	}

	// Verify block signature
	if err := block.Verify(validator.PublicKey); err != nil {
		return fmt.Errorf("invalid block signature: %w", err)
	}

	// Check block time (should not be too far in the future)
	now := time.Now().Unix()
	if block.Timestamp > now+int64(pos.BlockTime.Seconds()) {
		return fmt.Errorf("block timestamp too far in the future")
	}

	// Check that block timestamp is after previous block
	if prevBlock != nil && block.Timestamp <= prevBlock.Timestamp {
		return fmt.Errorf("block timestamp must be after previous block")
	}

	return nil
}

// CalculateReward calculates the block reward for a validator
func (pos *PoS) CalculateReward(block *blockchain.Block) uint64 {
	// Base reward
	reward := blockchain.BlockReward
	
	// Add transaction fees
	reward += block.TotalFees()
	
	return reward
}

// RegisterValidator registers a new validator
func (pos *PoS) RegisterValidator(validator *Validator) error {
	if validator.Stake < pos.MinStake {
		return fmt.Errorf("stake %d is below minimum %d", validator.Stake, pos.MinStake)
	}
	return pos.ValidatorSet.AddValidator(validator)
}

// UnregisterValidator removes a validator
func (pos *PoS) UnregisterValidator(address string) error {
	return pos.ValidatorSet.RemoveValidator(address)
}

// UpdateValidatorStake updates a validator's stake
func (pos *PoS) UpdateValidatorStake(address string, stake uint64) error {
	return pos.ValidatorSet.UpdateStake(address, stake)
}

// GetNextBlockTime returns the time when the next block should be created
func (pos *PoS) GetNextBlockTime(lastBlockTime int64) time.Time {
	return time.Unix(lastBlockTime, 0).Add(pos.BlockTime)
}

// ShouldCreateBlock checks if it's time to create a new block
func (pos *PoS) ShouldCreateBlock(lastBlockTime int64) bool {
	nextBlockTime := pos.GetNextBlockTime(lastBlockTime)
	return time.Now().After(nextBlockTime)
}

// SelectValidatorSimple selects a random validator (for testing/simple scenarios)
func (pos *PoS) SelectValidatorSimple() (*Validator, error) {
	validators := pos.ValidatorSet.GetValidators()
	if len(validators) == 0 {
		return nil, fmt.Errorf("no validators available")
	}

	// Filter eligible validators
	eligibleValidators := make([]*Validator, 0)
	for _, v := range validators {
		if v.CanValidate(pos.MinStake) {
			eligibleValidators = append(eligibleValidators, v)
		}
	}

	if len(eligibleValidators) == 0 {
		return nil, fmt.Errorf("no eligible validators")
	}

	// Random selection
	rand.Seed(time.Now().UnixNano())
	return eligibleValidators[rand.Intn(len(eligibleValidators))], nil
}
