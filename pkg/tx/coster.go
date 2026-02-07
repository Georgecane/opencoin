package tx

import (
	"fmt"

	"github.com/georgecane/opencoin/pkg/contracts"
	"github.com/georgecane/opencoin/pkg/encoding"
	"github.com/georgecane/opencoin/pkg/rc"
	"github.com/georgecane/opencoin/pkg/types"
)

// Coster computes RC cost for a transaction.
type Coster struct {
	Params   rc.Params
	Contracts *contracts.ContractEngine
}

// Cost computes RC cost based on size, instructions, and state writes.
func (c *Coster) Cost(txn *types.Transaction) (uint64, error) {
	if txn == nil {
		return 0, fmt.Errorf("tx is nil")
	}
	payload, err := DecodePayload(txn.Payload)
	if err != nil {
		return 0, err
	}
	sizeBytes, err := encoding.MarshalTransaction(txn)
	if err != nil {
		return 0, err
	}
	var instructions uint64
	var writes uint64

	switch p := payload.(type) {
	case Transfer:
		writes = 2
	case StakeDelegate, StakeUndelegate:
		writes = 1
	case ContractDeploy:
		if c.Contracts != nil {
			instructions = c.Contracts.EstimateInstructions(p.WASMCode)
		}
		writes = 1
	case ContractCall:
		if c.Contracts != nil {
			instructions = c.Contracts.EstimateContractCall(string(p.Address))
			writes = c.Contracts.EstimateStateWrites(string(p.Address))
		} else {
			writes = 1
		}
	case GovernanceProposal, GovernanceVote:
		writes = 1
	default:
		return 0, fmt.Errorf("unsupported payload type")
	}
	return c.Params.Cost(uint64(len(sizeBytes)), instructions, writes), nil
}
