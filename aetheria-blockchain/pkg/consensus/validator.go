package consensus

import (
	"crypto/ed25519"
	"fmt"

	"github.com/aetheria/blockchain/pkg/crypto"
)

// Validator represents a validator in the PoS consensus
type Validator struct {
	Address    string
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
	Stake      uint64
}

// NewValidator creates a new validator
func NewValidator(address string, publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey, stake uint64) *Validator {
	return &Validator{
		Address:    address,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		Stake:      stake,
	}
}

// ValidatorFromKeyPair creates a validator from a key pair
func ValidatorFromKeyPair(keyPair *crypto.KeyPair, stake uint64) *Validator {
	return &Validator{
		Address:    keyPair.Address(),
		PublicKey:  keyPair.PublicKey,
		PrivateKey: keyPair.PrivateKey,
		Stake:      stake,
	}
}

// ValidatorInfo represents public validator information
type ValidatorInfo struct {
	Address   string `json:"address"`
	PublicKey string `json:"public_key"`
	Stake     uint64 `json:"stake"`
	Weight    float64 `json:"weight"`
}

// GetInfo returns public validator information
func (v *Validator) GetInfo(totalStake uint64) *ValidatorInfo {
	weight := 0.0
	if totalStake > 0 {
		weight = float64(v.Stake) / float64(totalStake)
	}

	return &ValidatorInfo{
		Address:   v.Address,
		PublicKey: crypto.PublicKeyToHex(v.PublicKey),
		Stake:     v.Stake,
		Weight:    weight,
	}
}

// CanValidate checks if validator has minimum stake
func (v *Validator) CanValidate(minStake uint64) bool {
	return v.Stake >= minStake
}

// ValidatorSet manages a set of validators
type ValidatorSet struct {
	Validators map[string]*Validator
}

// NewValidatorSet creates a new validator set
func NewValidatorSet() *ValidatorSet {
	return &ValidatorSet{
		Validators: make(map[string]*Validator),
	}
}

// AddValidator adds a validator to the set
func (vs *ValidatorSet) AddValidator(validator *Validator) error {
	if _, exists := vs.Validators[validator.Address]; exists {
		return fmt.Errorf("validator already exists")
	}
	vs.Validators[validator.Address] = validator
	return nil
}

// RemoveValidator removes a validator from the set
func (vs *ValidatorSet) RemoveValidator(address string) error {
	if _, exists := vs.Validators[address]; !exists {
		return fmt.Errorf("validator not found")
	}
	delete(vs.Validators, address)
	return nil
}

// GetValidator returns a validator by address
func (vs *ValidatorSet) GetValidator(address string) (*Validator, error) {
	validator, exists := vs.Validators[address]
	if !exists {
		return nil, fmt.Errorf("validator not found")
	}
	return validator, nil
}

// UpdateStake updates a validator's stake
func (vs *ValidatorSet) UpdateStake(address string, stake uint64) error {
	validator, exists := vs.Validators[address]
	if !exists {
		return fmt.Errorf("validator not found")
	}
	validator.Stake = stake
	return nil
}

// TotalStake returns the total stake of all validators
func (vs *ValidatorSet) TotalStake() uint64 {
	var total uint64
	for _, validator := range vs.Validators {
		total += validator.Stake
	}
	return total
}

// GetValidators returns all validators
func (vs *ValidatorSet) GetValidators() []*Validator {
	validators := make([]*Validator, 0, len(vs.Validators))
	for _, validator := range vs.Validators {
		validators = append(validators, validator)
	}
	return validators
}

// GetValidatorInfos returns public information for all validators
func (vs *ValidatorSet) GetValidatorInfos() []*ValidatorInfo {
	totalStake := vs.TotalStake()
	infos := make([]*ValidatorInfo, 0, len(vs.Validators))
	for _, validator := range vs.Validators {
		infos = append(infos, validator.GetInfo(totalStake))
	}
	return infos
}

// Size returns the number of validators
func (vs *ValidatorSet) Size() int {
	return len(vs.Validators)
}
