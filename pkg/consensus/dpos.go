package consensus

import (
	"fmt"
	"sort"
	"sync"

	"github.com/georgecane/opencoin/pkg/types"
)

// DPoS manages validator registration, delegation, and slashing.
type DPoS struct {
	mu            sync.RWMutex
	validators    map[types.Address]*types.Validator
	minStake      uint64
	maxValidators uint32
}

// NewDPoS creates a new DPoS manager.
func NewDPoS(minStake uint64, maxValidators uint32) *DPoS {
	return &DPoS{
		validators:    make(map[types.Address]*types.Validator),
		minStake:      minStake,
		maxValidators: maxValidators,
	}
}

// RegisterValidator registers a new validator.
func (d *DPoS) RegisterValidator(address types.Address, publicKey []byte, stake uint64, commission uint16) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if stake < d.minStake {
		return fmt.Errorf("stake below minimum: %d < %d", stake, d.minStake)
	}
	if _, exists := d.validators[address]; exists {
		return fmt.Errorf("validator already registered: %s", address)
	}
	if uint32(len(d.validators)) >= d.maxValidators {
		return fmt.Errorf("max validators reached: %d", d.maxValidators)
	}

	d.validators[address] = &types.Validator{
		Address:     address,
		PublicKey:   publicKey,
		Power:       stake,
		Stake:       stake,
		Delegations: make(map[types.Address]uint64),
		Commission:  commission,
	}
	return nil
}

// Delegate delegates stake to a validator.
func (d *DPoS) Delegate(delegator types.Address, validator types.Address, amount uint64) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	v, ok := d.validators[validator]
	if !ok {
		return fmt.Errorf("validator not found: %s", validator)
	}
	if amount == 0 {
		return fmt.Errorf("delegation amount must be positive")
	}
	v.Delegations[delegator] += amount
	v.Power += amount
	return nil
}

// Undelegate removes a delegation.
func (d *DPoS) Undelegate(delegator types.Address, validator types.Address, amount uint64) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	v, ok := d.validators[validator]
	if !ok {
		return fmt.Errorf("validator not found: %s", validator)
	}
	delegated := v.Delegations[delegator]
	if delegated < amount {
		return fmt.Errorf("insufficient delegation")
	}
	v.Delegations[delegator] -= amount
	v.Power -= amount
	return nil
}

// SlashDoubleSign slashes and jails a validator for double-signing.
func (d *DPoS) SlashDoubleSign(validator types.Address, slashBps uint64, jailEpochs uint64, currentEpoch uint64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	v, ok := d.validators[validator]
	if !ok {
		return fmt.Errorf("validator not found: %s", validator)
	}
	slashAmount := (v.Stake * slashBps) / 10000
	if v.Stake >= slashAmount {
		v.Stake -= slashAmount
	}
	v.Power = v.Stake
	v.JailedUntilEpoch = currentEpoch + jailEpochs
	return nil
}

// SlashOffline slashes and jails a validator for downtime.
func (d *DPoS) SlashOffline(validator types.Address, slashBps uint64, jailEpochs uint64, currentEpoch uint64) error {
	return d.SlashDoubleSign(validator, slashBps, jailEpochs, currentEpoch)
}

// ValidatorSet returns the current validator set ordered by power, then address.
func (d *DPoS) ValidatorSet() *types.ValidatorSet {
	d.mu.RLock()
	defer d.mu.RUnlock()

	validators := make([]*types.Validator, 0, len(d.validators))
	var totalPower uint64
	for _, v := range d.validators {
		copyV := *v
		validators = append(validators, &copyV)
		totalPower += v.Power
	}
	sort.Slice(validators, func(i, j int) bool {
		if validators[i].Power == validators[j].Power {
			return validators[i].Address < validators[j].Address
		}
		return validators[i].Power > validators[j].Power
	})

	index := make(map[types.Address]uint32)
	for i, v := range validators {
		v.Index = uint32(i)
		index[v.Address] = uint32(i)
	}

	return &types.ValidatorSet{
		Validators:  validators,
		TotalPower:  totalPower,
		IndexByAddr: index,
	}
}

// GetValidator returns a validator by address.
func (d *DPoS) GetValidator(address types.Address) *types.Validator {
	d.mu.RLock()
	defer d.mu.RUnlock()
	v, ok := d.validators[address]
	if !ok {
		return nil
	}
	copyV := *v
	return &copyV
}
