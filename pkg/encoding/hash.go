package encoding

import (
	"crypto/sha256"
	"fmt"

	"github.com/georgecane/opencoin/pkg/types"
)

// HashBytes computes SHA-256 over input data.
func HashBytes(data []byte) types.Hash {
	sum := sha256.Sum256(data)
	return types.Hash(sum)
}

// HashTransaction computes the canonical transaction hash.
func HashTransaction(tx *types.Transaction) (types.Hash, error) {
	b, err := MarshalTransaction(tx)
	if err != nil {
		return types.Hash{}, err
	}
	return HashBytes(b), nil
}

// HashBlock computes the canonical block hash.
func HashBlock(block *types.Block) (types.Hash, error) {
	b, err := MarshalBlock(block)
	if err != nil {
		return types.Hash{}, err
	}
	return HashBytes(b), nil
}

// HashStateNode computes the canonical state node hash.
func HashStateNode(node *types.StateNode) (types.Hash, error) {
	b, err := MarshalStateNode(node)
	if err != nil {
		return types.Hash{}, err
	}
	return HashBytes(b), nil
}

// HashQuorumCertificate computes the canonical QC hash.
func HashQuorumCertificate(qc *types.QuorumCertificate) (types.Hash, error) {
	b, err := MarshalQuorumCertificate(qc)
	if err != nil {
		return types.Hash{}, err
	}
	return HashBytes(b), nil
}

// HashProposal computes the canonical proposal hash.
func HashProposal(p *types.Proposal) (types.Hash, error) {
	b, err := MarshalProposal(p)
	if err != nil {
		return types.Hash{}, err
	}
	return HashBytes(b), nil
}

// HashPrecommitVote computes the canonical vote hash.
func HashPrecommitVote(v *types.PrecommitVote) (types.Hash, error) {
	b, err := MarshalPrecommitVote(v)
	if err != nil {
		return types.Hash{}, err
	}
	return HashBytes(b), nil
}

// HashViewChange computes the canonical view-change hash.
func HashViewChange(vc *types.ViewChange) (types.Hash, error) {
	b, err := MarshalViewChange(vc)
	if err != nil {
		return types.Hash{}, err
	}
	return HashBytes(b), nil
}

// HashBytesOrErr wraps HashBytes to satisfy interfaces needing error.
func HashBytesOrErr(data []byte, err error) (types.Hash, error) {
	if err != nil {
		return types.Hash{}, fmt.Errorf("hash bytes: %w", err)
	}
	return HashBytes(data), nil
}
