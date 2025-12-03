package types

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// Block represents a block in the DAG
type Block struct {
	Hash      string
	Height    uint64
	Timestamp time.Time
	Proposer  string   // validator address
	Parents   []string // parent block hashes (for DAG)
	Txs       []*Transaction
	StateRoot string // merkle root of DAG state
	Signature []byte
}

// Transaction represents a transaction in the blockchain
type Transaction struct {
	ID        string
	From      string
	To        string
	Amount    uint64
	Contract  *ContractCall `json:"contract,omitempty"`
	Nonce     uint64
	Timestamp time.Time
	Signature []byte
}

// ContractCall represents a smart contract invocation
type ContractCall struct {
	Address string   // contract address
	Method  string   // method name
	Args    [][]byte // arguments
	Gas     uint64
}

// DAGVertex represents a node in the DAG
type DAGVertex struct {
	BlockHash string
	Block     *Block
	Parents   []*DAGVertex // parent vertices
	Children  []*DAGVertex // child vertices
	Order     uint64       // topological order
}

// ValidatorSet represents the active validator set
type ValidatorSet struct {
	Validators map[string]*Validator
	TotalPower uint64
}

// Validator represents a validator in DPoS
type Validator struct {
	Address     string
	PublicKey   []byte
	Power       uint64            // voting power from stake
	Stake       uint64            // amount staked by validator
	Delegations map[string]uint64 // delegator -> amount
	Commission  uint16            // commission rate in basis points
}

// DelegationRecord tracks delegations
type DelegationRecord struct {
	Delegator string
	Validator string
	Amount    uint64
	Timestamp time.Time
}

// AccountState represents account balance and nonce
type AccountState struct {
	Address string
	Balance uint64
	Nonce   uint64
	Code    []byte // WASM contract code if applicable
}

// ComputeHash returns the SHA256 hash of the block
func (b *Block) ComputeHash() string {
	data := b.Timestamp.String() + b.Proposer + b.StateRoot
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// ComputeTxHash returns the SHA256 hash of a transaction
func (t *Transaction) ComputeHash() string {
	data := t.From + t.To + string(rune(t.Amount)) + t.Timestamp.String()
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
