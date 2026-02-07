package state

import (
	"fmt"

	"github.com/cockroachdb/pebble"

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

// PreviewBlock computes the expected state root for a block without mutating persistent state.
func (s *State) PreviewBlock(block *types.Block, engine *contracts.ContractEngine) (types.Hash, error) {
	if block == nil {
		return types.Hash{}, fmt.Errorf("block is nil")
	}
	lastTimestamps, err := s.store.GetLastTimestamps()
	if err != nil {
		return types.Hash{}, err
	}
	effectiveTime := rc.EffectiveTime(block.Timestamp, lastTimestamps, s.rcParams.MaxSkewSec)

	batch := s.store.NewIndexedBatch()
	defer batch.Close()

	get := func(addr types.Address) (*types.Account, error) {
		acct, err := getAccountFromReader(batch, addr)
		if err != nil {
			return nil, err
		}
		if acct == nil {
			return &types.Account{Address: addr}, nil
		}
		return acct, nil
	}
	set := func(acct *types.Account) error {
		return setAccountWithWriter(batch, acct, nil)
	}

	for _, tx := range block.Transactions {
		if err := s.applyTransactionWithKV(tx, engine, effectiveTime, get, set, true); err != nil {
			return types.Hash{}, err
		}
	}

	return ComputeStateRootFromReader(batch)
}

// ApplyBlock applies a block to state, updating RC and computing a new state root.
func (s *State) ApplyBlock(block *types.Block, engine *contracts.ContractEngine) (types.Hash, error) {
	if block == nil {
		return types.Hash{}, fmt.Errorf("block is nil")
	}
	previewRoot, err := s.PreviewBlock(block, engine)
	if err != nil {
		return types.Hash{}, err
	}
	if block.StateRoot != previewRoot {
		return types.Hash{}, fmt.Errorf("state root mismatch")
	}
	lastTimestamps, err := s.store.GetLastTimestamps()
	if err != nil {
		return types.Hash{}, err
	}
	effectiveTime := rc.EffectiveTime(block.Timestamp, lastTimestamps, s.rcParams.MaxSkewSec)

	batch := s.store.NewIndexedBatch()
	defer batch.Close()

	get := func(addr types.Address) (*types.Account, error) {
		acct, err := getAccountFromReader(batch, addr)
		if err != nil {
			return nil, err
		}
		if acct == nil {
			return &types.Account{Address: addr}, nil
		}
		return acct, nil
	}
	set := func(acct *types.Account) error {
		return setAccountWithWriter(batch, acct, nil)
	}
	for _, tx := range block.Transactions {
		if err := s.applyTransactionWithKV(tx, engine, effectiveTime, get, set, false); err != nil {
			return types.Hash{}, err
		}
	}

	// Update timestamp window with raw block timestamp.
	lastTimestamps = append(lastTimestamps, block.Timestamp)
	if len(lastTimestamps) > s.rcParams.WindowN {
		lastTimestamps = lastTimestamps[len(lastTimestamps)-s.rcParams.WindowN:]
	}
	if err := setLastTimestampsWithWriter(batch, lastTimestamps); err != nil {
		return types.Hash{}, err
	}
	root, err := ComputeStateRootFromReader(batch)
	if err != nil {
		return types.Hash{}, err
	}
	if root != block.StateRoot {
		return types.Hash{}, fmt.Errorf("state root mismatch after apply")
	}
	hash, err := encoding.HashBlock(block)
	if err != nil {
		return types.Hash{}, err
	}
	if err := setBlockWithWriter(batch, block, hash); err != nil {
		return types.Hash{}, err
	}
	if err := batch.Commit(pebble.Sync); err != nil {
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
	get := func(addr types.Address) (*types.Account, error) {
		return s.GetAccount(addr)
	}
	set := func(acct *types.Account) error {
		return s.store.SetAccount(acct)
	}
	return s.applyTransactionWithKV(txn, engine, effectiveTime, get, set, false)
}

func (s *State) applyTransactionWithKV(txn *types.Transaction, engine *contracts.ContractEngine, effectiveTime int64, get func(types.Address) (*types.Account, error), set func(*types.Account) error, preview bool) error {
	if txn == nil {
		return fmt.Errorf("transaction is nil")
	}
	if txn.From == "" || txn.To == "" {
		return fmt.Errorf("invalid sender or recipient")
	}

	sender, err := get(txn.From)
	if err != nil {
		return err
	}
	if sender == nil {
		sender = &types.Account{Address: txn.From}
	}
	payloadEnv, err := tx.DecodePayload(txn.Payload)
	if err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}
	pubKey, register, err := tx.ResolveSenderPubKey(txn, sender.PubKey, payloadEnv.SenderPubKey)
	if err != nil {
		return err
	}
	if err := tx.VerifySignature(txn, pubKey); err != nil {
		return err
	}
	if register {
		sender.PubKey = pubKey
	}

	// RC regeneration for sender.
	sender.RC, sender.LastRCEffectiveTime = s.rcParams.Regen(sender.RC, sender.Stake, sender.LastRCEffectiveTime, effectiveTime)
	sender.RCMax = s.rcParams.RCMax(sender.Stake)

	if sender.Nonce != txn.Nonce {
		return fmt.Errorf("invalid nonce: expected %d, got %d", sender.Nonce, txn.Nonce)
	}

	sizeBytes, err := encoding.MarshalTransaction(txn)
	if err != nil {
		return err
	}

	var instructions uint64
	var stateWrites uint64

	switch p := payloadEnv.Payload.(type) {
	case tx.Transfer:
		if sender.Balance < p.Amount {
			return fmt.Errorf("insufficient balance")
		}
		sender.Balance -= p.Amount
		receiver, err := get(p.To)
		if err != nil {
			return err
		}
		if receiver == nil {
			receiver = &types.Account{Address: p.To}
		}
		receiver.Balance += p.Amount
		if err := set(receiver); err != nil {
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
		if engine == nil {
			return fmt.Errorf("contract engine not configured")
		}
		addr := txn.To
		if addr == "" {
			return fmt.Errorf("contract deploy missing target address")
		}
		if !preview {
			if err := engine.DeployContract(string(addr), p.WASMCode, string(addr)); err != nil {
				return err
			}
		} else if err := contracts.ValidateWasmCode(p.WASMCode); err != nil {
			return err
		}
		contractAcct, err := get(addr)
		if err != nil {
			return err
		}
		if contractAcct == nil {
			contractAcct = &types.Account{Address: addr}
		}
		contractAcct.Code = append(contractAcct.Code[:0], p.WASMCode...)
		if err := set(contractAcct); err != nil {
			return err
		}
		stateWrites = 1
	case tx.ContractCall:
		if engine == nil {
			return fmt.Errorf("contract engine not configured")
		}
		if preview {
			instructions = engine.EstimateContractCall(string(p.Address))
			stateWrites = engine.EstimateStateWrites(string(p.Address))
		} else {
			result, err := engine.ExecuteContractWithResult(&contracts.ExecutionContext{
				Caller:       string(txn.From),
				ContractAddr: string(p.Address),
			})
			if err != nil {
				return err
			}
			instructions = result.Instructions
			stateWrites = result.StateWrites
		}
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

	return set(sender)
}
