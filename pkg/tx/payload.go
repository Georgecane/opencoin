package tx

import "github.com/georgecane/opencoin/pkg/types"

// PayloadType identifies the payload variant.
type PayloadType uint8

const (
	PayloadTransfer PayloadType = iota + 1
	PayloadStakeDelegate
	PayloadStakeUndelegate
	PayloadContractDeploy
	PayloadContractCall
	PayloadGovernanceProposal
	PayloadGovernanceVote
)

// Payload is implemented by all transaction payload variants.
type Payload interface {
	PayloadType() PayloadType
}

type Transfer struct {
	To     types.Address
	Amount uint64
}

func (Transfer) PayloadType() PayloadType { return PayloadTransfer }

type StakeDelegate struct {
	Validator types.Address
	Amount    uint64
}

func (StakeDelegate) PayloadType() PayloadType { return PayloadStakeDelegate }

type StakeUndelegate struct {
	Validator types.Address
	Amount    uint64
}

func (StakeUndelegate) PayloadType() PayloadType { return PayloadStakeUndelegate }

type ContractDeploy struct {
	WASMCode []byte
	Salt     []byte
}

func (ContractDeploy) PayloadType() PayloadType { return PayloadContractDeploy }

type ContractCall struct {
	Address types.Address
	Method  string
	Args    [][]byte
}

func (ContractCall) PayloadType() PayloadType { return PayloadContractCall }

type GovernanceProposal struct {
	Title       string
	Description string
	ParamKey    string
	ParamValue  string
}

func (GovernanceProposal) PayloadType() PayloadType { return PayloadGovernanceProposal }

type GovernanceVote struct {
	ProposalID uint64
	Option     types.VoteOption
}

func (GovernanceVote) PayloadType() PayloadType { return PayloadGovernanceVote }
