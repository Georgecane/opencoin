package genesis

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/georgecane/opencoin/pkg/crypto"
	"github.com/georgecane/opencoin/pkg/rc"
	"github.com/georgecane/opencoin/pkg/types"
)

// Genesis defines initial chain configuration.
type Genesis struct {
	ChainID     string           `json:"chain_id"`
	GenesisTime time.Time        `json:"genesis_time"`
	RCParams    rc.Params        `json:"rc_params"`
	MinStake    uint64           `json:"min_validator_stake"`
	Validators  []GenesisValidator `json:"validators"`
	Accounts    []GenesisAccount `json:"accounts"`
}

type GenesisValidator struct {
	OperatorAddress types.Address `json:"operator_address"`
	ConsensusPubKey []byte        `json:"consensus_pubkey"`
	Stake           uint64        `json:"stake"`
	Commission      uint16        `json:"commission"`
}

type GenesisAccount struct {
	Address types.Address `json:"address"`
	Balance uint64        `json:"balance"`
	Stake   uint64        `json:"stake"`
}

// DefaultGenesis returns a default genesis config.
func DefaultGenesis() *Genesis {
	return &Genesis{
		ChainID:     "opencoin-1",
		GenesisTime: time.Now().UTC(),
		RCParams: rc.Params{
			Alpha:      1000,
			Beta:       1,
			CSize:      1,
			CCompute:   1,
			CStorage:   50,
			MaxSkewSec: 30,
			WindowN:    11,
		},
		MinStake: 1_000_000,
	}
}

// Validate validates genesis parameters.
func (g *Genesis) Validate() error {
	if g.ChainID == "" {
		return fmt.Errorf("chain_id required")
	}
	if err := g.RCParams.ValidateGenesis(); err != nil {
		return err
	}
	if g.MinStake < 1_000_000 || g.MinStake > 1_000_000_000_000_000_000 {
		return fmt.Errorf("min_validator_stake out of bounds")
	}
	for _, v := range g.Validators {
		if v.Stake < g.MinStake {
			return fmt.Errorf("validator stake below minimum")
		}
		if v.OperatorAddress == "" {
			return fmt.Errorf("validator missing operator address")
		}
		if _, err := crypto.DecodeAddress(string(v.OperatorAddress)); err != nil {
			return fmt.Errorf("invalid operator address")
		}
		if len(v.ConsensusPubKey) == 0 {
			return fmt.Errorf("validator missing consensus pubkey")
		}
	}
	return nil
}

// Save writes genesis to a file.
func (g *Genesis) Save(path string) error {
	if err := g.Validate(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// Load reads genesis from a file.
func Load(path string) (*Genesis, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var g Genesis
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, err
	}
	if err := g.Validate(); err != nil {
		return nil, err
	}
	return &g, nil
}
