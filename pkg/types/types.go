package types

import "fmt"

// Address is a bech32-encoded account identifier.
type Address string

// Hash is a 32-byte hash used for blocks, state roots, and consensus messages.
type Hash [32]byte

func (h Hash) String() string {
	return fmt.Sprintf("%x", h[:])
}

// Transaction is the canonical transaction format.
type Transaction struct {
	From      Address
	To        Address
	Nonce     uint64
	Payload   []byte
	Signature []byte // Ed25519
}

// Block is the canonical block format.
type Block struct {
	Height        uint64
	PrevHash      Hash
	StateRoot     Hash
	Timestamp     int64
	Proposer      Address
	Transactions  []*Transaction
	ValidatorSigs [][]byte // ordered by validator-set index, empty slice means missing signature
}

// StateNode represents a DAG node for state versioning.
type StateNode struct {
	RootHash Hash
	Parents  []Hash
	Height   uint64
}

// Proposal is a HotStuff-style proposal message.
type Proposal struct {
	Block       *Block
	Round       uint64
	ProposerSig []byte
}

// PrecommitVote is a HotStuff-style vote message.
type PrecommitVote struct {
	BlockHash Hash
	Height    uint64
	Round     uint64
	Validator Address
	Signature []byte
}

// QuorumCertificate represents a QC for a round.
// SigBitmap indicates which validators signed by index.
// AggregatedSig is optional; when empty, Signatures should be used.
type QuorumCertificate struct {
	BlockHash     Hash
	Height        uint64
	Round         uint64
	SigBitmap     []byte
	AggregatedSig []byte
	Signatures    [][]byte // ordered by validator-set index, empty slice means missing signature
}

// ViewChange represents a view change message.
type ViewChange struct {
	Height    uint64
	Round     uint64
	Validator Address
	Signature []byte
}

// Validator represents a validator in DPoS.
type Validator struct {
	OperatorAddress Address
	ConsensusPubKey []byte
	Power       uint64
	Stake       uint64
	Delegations map[Address]uint64
	Commission  uint16
	Index       uint32

	// Slashing/jailing state.
	JailedUntilEpoch uint64
}

// ValidatorSet is an ordered validator set.
type ValidatorSet struct {
	Validators []*Validator // index is position in slice
	TotalPower uint64
	IndexByAddr map[Address]uint32
}

// Account represents on-chain account state.
type Account struct {
	Address             Address
	Balance             uint64
	Nonce               uint64
	Stake               uint64
	RC                  uint64
	RCMax               uint64
	LastRCEffectiveTime int64
	Code                []byte
	PubKey              []byte
}

// Contract represents a deployed contract.
type Contract struct {
	Address  Address
	Owner    Address
	Code     []byte
	Balance  uint64
	Deployed int64
}

// Governance proposal and vote.
type GovernanceProposal struct {
	ID          uint64
	Title       string
	Description string
	ParamKey    string
	ParamValue  string
	Submitter   Address
}

type VoteOption uint8

const (
	VoteOptionUnspecified VoteOption = iota
	VoteOptionYes
	VoteOptionNo
	VoteOptionAbstain
	VoteOptionVeto
)

type GovernanceVote struct {
	ProposalID uint64
	Voter      Address
	Option     VoteOption
}
