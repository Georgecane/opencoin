package state

import (
	"fmt"

	"github.com/georgecane/opencoin/pkg/contracts"
	"github.com/georgecane/opencoin/pkg/encoding"
	"github.com/georgecane/opencoin/pkg/rc"
	"github.com/georgecane/opencoin/pkg/tx"
	"github.com/georgecane/opencoin/pkg/types"
)

// State coordinates persistent state, RC accounting, and DAG roots.
type State struct {
	store    *Store
	dag      *DAG
	rcParams rc.Params
}

// NewState creates a new State manager.
func NewState(store *Store, dag *DAG, rcParams rc.Params) *State {
	return &State{store: store, dag: dag, rcParams: rcParams}
}

// Store returns the underlying store.
func (s *State) Store() *Store { return s.store }

// RCParams returns the RC parameters.
func (s *State) RCParams() rc.Params { return s.rcParams }

// GetAccount returns an account or a zero-value account.
func (s *State) GetAccount(addr types.Address) (*types.Account, error) {
	acct, err := s.store.GetAccount(addr)
	if err != nil {
		return nil, err
	}
	if acct == nil {
		return &types.Account{Address: addr}, nil
	}
	return acct, nil
}

// ApplyBlock applies a block to state, updating RC and computing a new state root.
func (s *State) ApplyBlock(block *types.Block, engine *contracts.ContractEngine) (types.Hash, error) {
	if block == nil {
		return types.Hash{}, fmt.Errorf("block is nil")
	}
	lastTimestamps, err := s.store.GetLastTimestamps()
	if err != nil {
		return types.Hash{}, err
	}
	effectiveTime := rc.EffectiveTime(block.Timestamp, lastTimestamps, s.rcParams.MaxSkewSec)

	for _, tx := range block.Transactions {
		if err := s.applyTransaction(tx, engine, effectiveTime); err != nil {
			return types.Hash{}, err
		}
	}

	// Update timestamp window with raw block timestamp.
	lastTimestamps = append(lastTimestamps, block.Timestamp)
	if len(lastTimestamps) > s.rcParams.WindowN {
		lastTimestamps = lastTimestamps[len(lastTimestamps)-s.rcParams.WindowN:]
	}
	if err := s.store.SetLastTimestamps(lastTimestamps); err != nil {
		return types.Hash{}, err
	}

	root, err := ComputeStateRoot(s.store)
	if err != nil {
		return types.Hash{}, err
	}
	node := &types.StateNode{
		RootHash: root,
		Parents:  []types.Hash{block.PrevHash},
		Height:   block.Height,
	}
	_ = s.dag.AddNode(node)
	return root, nil
}

func (s *State) applyTransaction(txn *types.Transaction, engine *contracts.ContractEngine, effectiveTime int64) error {
	if txn == nil {
		return fmt.Errorf("transaction is nil")
	}
	if txn.From == "" || txn.To == "" {
		return fmt.Errorf("invalid sender or recipient")
	}

	sender, err := s.GetAccount(txn.From)
	if err != nil {
		return err
	}
	// RC regeneration for sender.
	sender.RC, sender.LastRCEffectiveTime = s.rcParams.Regen(sender.RC, sender.Stake, sender.LastRCEffectiveTime, effectiveTime)
	sender.RCMax = s.rcParams.RCMax(sender.Stake)

	if sender.Nonce != txn.Nonce {
		return fmt.Errorf("invalid nonce: expected %d, got %d", sender.Nonce, txn.Nonce)
	}

	payload, err := tx.DecodePayload(txn.Payload)
	if err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}

	sizeBytes, err := encoding.MarshalTransaction(txn)
	if err != nil {
		return err
	}

	var instructions uint64
	var stateWrites uint64

	switch p := payload.(type) {
	case tx.Transfer:
		if sender.Balance < p.Amount {
			return fmt.Errorf("insufficient balance")
		}
		sender.Balance -= p.Amount
		receiver, err := s.GetAccount(p.To)
		if err != nil {
			return err
		}
		receiver.Balance += p.Amount
		if err := s.store.SetAccount(receiver); err != nil {
			return err
		}
		stateWrites = 2
	case tx.StakeDelegate:
		if sender.Balance < p.Amount {
			return fmt.Errorf("insufficient balance")
		}
		sender.Balance -= p.Amount
		sender.Stake += p.Amount
		stateWrites = 1
	case tx.StakeUndelegate:
		if sender.Stake < p.Amount {
			return fmt.Errorf("insufficient stake")
		}
		sender.Stake -= p.Amount
		sender.Balance += p.Amount
		stateWrites = 1
	case tx.ContractDeploy:
		addr := txn.To
		if addr == "" {
			return fmt.Errorf("contract deploy missing target address")
		}
		if err := engine.DeployContract(string(addr), p.WASMCode, string(addr)); err != nil {
			return err
		}
		contractAcct, err := s.GetAccount(addr)
		if err != nil {
			return err
		}
		contractAcct.Code = append(contractAcct.Code[:0], p.WASMCode...)
		if err := s.store.SetAccount(contractAcct); err != nil {
			return err
		}
		stateWrites = 1
	case tx.ContractCall:
		result, err := engine.ExecuteContractWithResult(&contracts.ExecutionContext{
			Caller:       string(txn.From),
			ContractAddr: string(p.Address),
		})
		if err != nil {
			return err
		}
		instructions = result.Instructions
		stateWrites = result.StateWrites
	case tx.GovernanceProposal:
		// Governance state handled in governance module; minimal placeholder.
		stateWrites = 1
	case tx.GovernanceVote:
		stateWrites = 1
	default:
		return fmt.Errorf("unsupported payload type")
	}

	cost := s.rcParams.Cost(uint64(len(sizeBytes)), instructions, stateWrites)
	if sender.RC < cost {
		return fmt.Errorf("insufficient rc")
	}
	sender.RC -= cost
	sender.Nonce++
	sender.RCMax = s.rcParams.RCMax(sender.Stake)

	return s.store.SetAccount(sender)
}
