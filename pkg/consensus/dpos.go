package consensus

import (
	"fmt"
	"sync"

	"github.com/georgecane/opencoin/pkg/types"
)

// DPoSEngine manages Delegated Proof of Stake
type DPoSEngine struct {
	mu            sync.RWMutex
	validators    *types.ValidatorSet
	delegations   map[string][]*types.DelegationRecord
	minStake      uint64
	maxValidators uint32
}

// NewDPoSEngine creates a new DPoS engine
func NewDPoSEngine(minStake uint64, maxValidators uint32) *DPoSEngine {
	return &DPoSEngine{
		validators:    &types.ValidatorSet{Validators: make(map[string]*types.Validator)},
		delegations:   make(map[string][]*types.DelegationRecord),
		minStake:      minStake,
		maxValidators: maxValidators,
	}
}

// RegisterValidator registers a new validator
func (d *DPoSEngine) RegisterValidator(address string, publicKey []byte, stake uint64) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if stake < d.minStake {
		return fmt.Errorf("stake below minimum: %d < %d", stake, d.minStake)
	}

	if _, exists := d.validators.Validators[address]; exists {
		return fmt.Errorf("validator already registered: %s", address)
	}

	if uint32(len(d.validators.Validators)) >= d.maxValidators {
		return fmt.Errorf("max validators reached: %d", d.maxValidators)
	}

	validator := &types.Validator{
		Address:     address,
		PublicKey:   publicKey,
		Stake:       stake,
		Power:       stake, // Initial power equals stake
		Delegations: make(map[string]uint64),
		Commission:  1000, // 10% commission
	}

	d.validators.Validators[address] = validator
	d.validators.TotalPower += stake

	return nil
}

// Delegate allows an account to delegate stake to a validator
func (d *DPoSEngine) Delegate(delegator, validator string, amount uint64) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	v, exists := d.validators.Validators[validator]
	if !exists {
		return fmt.Errorf("validator not found: %s", validator)
	}

	if amount == 0 {
		return fmt.Errorf("delegation amount must be positive")
	}

	// Record delegation
	delegation := &types.DelegationRecord{
		Delegator: delegator,
		Validator: validator,
		Amount:    amount,
	}

	d.delegations[delegator] = append(d.delegations[delegator], delegation)

	// Update validator delegations
	v.Delegations[delegator] += amount
	v.Power += amount
	d.validators.TotalPower += amount

	return nil
}

// Undelegate removes a delegation
func (d *DPoSEngine) Undelegate(delegator, validator string, amount uint64) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	v, exists := d.validators.Validators[validator]
	if !exists {
		return fmt.Errorf("validator not found: %s", validator)
	}

	delegated, hasDelegation := v.Delegations[delegator]
	if !hasDelegation || delegated < amount {
		return fmt.Errorf("insufficient delegation: %d < %d", delegated, amount)
	}

	v.Delegations[delegator] -= amount
	v.Power -= amount
	d.validators.TotalPower -= amount

	return nil
}

// GetValidatorSet returns the current validator set
func (d *DPoSEngine) GetValidatorSet() *types.ValidatorSet {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Return a copy to prevent external modification
	validatorsCopy := make(map[string]*types.Validator)
	for k, v := range d.validators.Validators {
		validatorCopy := *v
		validatorsCopy[k] = &validatorCopy
	}

	return &types.ValidatorSet{
		Validators: validatorsCopy,
		TotalPower: d.validators.TotalPower,
	}
}

// GetValidator returns a specific validator
func (d *DPoSEngine) GetValidator(address string) *types.Validator {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if v, ok := d.validators.Validators[address]; ok {
		validatorCopy := *v
		return &validatorCopy
	}
	return nil
}

// DistributeRewards distributes block rewards to validators and delegators
func (d *DPoSEngine) DistributeRewards(blockReward uint64, proposer string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	v, exists := d.validators.Validators[proposer]
	if !exists {
		return fmt.Errorf("validator not found: %s", proposer)
	}

	// Proposer gets the block reward
	v.Stake += blockReward

	// Distribute to delegators based on commission
	commission := (blockReward * uint64(v.Commission)) / 10000
	delegatorReward := blockReward - commission

	// Proportionally distribute to delegators
	for delegator, amount := range v.Delegations {
		if v.Power > 0 {
			share := (delegatorReward * amount) / v.Power
			// TODO: Track delegator rewards in separate structure
			_ = delegator
			_ = share
		}
	}

	return nil
}

// ValidateBlock checks if a block proposer is a valid validator
func (d *DPoSEngine) ValidateBlock(proposer string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	_, exists := d.validators.Validators[proposer]
	return exists
}
