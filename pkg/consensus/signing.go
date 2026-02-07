package consensus

import (
	"github.com/georgecane/opencoin/pkg/encoding"
	"github.com/georgecane/opencoin/pkg/types"
)

// ProposalSignBytes returns deterministic signing bytes for a proposal.
func ProposalSignBytes(p *types.Proposal) ([]byte, error) {
	if p == nil {
		return nil, nil
	}
	cp := *p
	cp.ProposerSig = nil
	if cp.Block != nil {
		blockCopy := *cp.Block
		blockCopy.ValidatorSigs = nil
		cp.Block = &blockCopy
	}
	return encoding.MarshalProposal(&cp)
}

// PrecommitSignBytes returns deterministic signing bytes for a precommit vote.
func PrecommitSignBytes(v *types.PrecommitVote) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	cp := *v
	cp.Signature = nil
	return encoding.MarshalPrecommitVote(&cp)
}

// ViewChangeSignBytes returns deterministic signing bytes for a view change.
func ViewChangeSignBytes(v *types.ViewChange) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	cp := *v
	cp.Signature = nil
	return encoding.MarshalViewChange(&cp)
}
