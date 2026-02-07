package encoding

import (
	"encoding/binary"
	"fmt"

	"google.golang.org/protobuf/encoding/protowire"

	"github.com/georgecane/opencoin/pkg/types"
)

// MarshalTransaction deterministically encodes a Transaction in protobuf wire format.
func MarshalTransaction(tx *types.Transaction) ([]byte, error) {
	if tx == nil {
		return nil, fmt.Errorf("transaction is nil")
	}
	var b []byte
	b = protowire.AppendTag(b, 1, protowire.BytesType)
	b = protowire.AppendBytes(b, []byte(tx.From))
	b = protowire.AppendTag(b, 2, protowire.BytesType)
	b = protowire.AppendBytes(b, []byte(tx.To))
	b = protowire.AppendTag(b, 3, protowire.VarintType)
	b = protowire.AppendVarint(b, tx.Nonce)
	b = protowire.AppendTag(b, 4, protowire.BytesType)
	b = protowire.AppendBytes(b, tx.Payload)
	b = protowire.AppendTag(b, 5, protowire.BytesType)
	b = protowire.AppendBytes(b, tx.Signature)
	return b, nil
}

// MarshalBlock deterministically encodes a Block in protobuf wire format.
func MarshalBlock(block *types.Block) ([]byte, error) {
	if block == nil {
		return nil, fmt.Errorf("block is nil")
	}
	var b []byte
	b = protowire.AppendTag(b, 1, protowire.VarintType)
	b = protowire.AppendVarint(b, block.Height)
	b = protowire.AppendTag(b, 2, protowire.BytesType)
	b = protowire.AppendBytes(b, block.PrevHash[:])
	b = protowire.AppendTag(b, 3, protowire.BytesType)
	b = protowire.AppendBytes(b, block.StateRoot[:])
	b = protowire.AppendTag(b, 4, protowire.VarintType)
	b = protowire.AppendVarint(b, uint64(block.Timestamp))
	b = protowire.AppendTag(b, 5, protowire.BytesType)
	b = protowire.AppendBytes(b, []byte(block.Proposer))
	for _, tx := range block.Transactions {
		txBytes, err := MarshalTransaction(tx)
		if err != nil {
			return nil, err
		}
		b = protowire.AppendTag(b, 6, protowire.BytesType)
		b = protowire.AppendBytes(b, txBytes)
	}
	for _, sig := range block.ValidatorSigs {
		b = protowire.AppendTag(b, 7, protowire.BytesType)
		b = protowire.AppendBytes(b, sig)
	}
	return b, nil
}

// MarshalStateNode deterministically encodes a StateNode.
func MarshalStateNode(node *types.StateNode) ([]byte, error) {
	if node == nil {
		return nil, fmt.Errorf("state node is nil")
	}
	var b []byte
	b = protowire.AppendTag(b, 1, protowire.BytesType)
	b = protowire.AppendBytes(b, node.RootHash[:])
	for _, parent := range node.Parents {
		b = protowire.AppendTag(b, 2, protowire.BytesType)
		b = protowire.AppendBytes(b, parent[:])
	}
	b = protowire.AppendTag(b, 3, protowire.VarintType)
	b = protowire.AppendVarint(b, node.Height)
	return b, nil
}

// MarshalProposal deterministically encodes a Proposal.
func MarshalProposal(p *types.Proposal) ([]byte, error) {
	if p == nil {
		return nil, fmt.Errorf("proposal is nil")
	}
	var b []byte
	blockBytes, err := MarshalBlock(p.Block)
	if err != nil {
		return nil, err
	}
	b = protowire.AppendTag(b, 1, protowire.BytesType)
	b = protowire.AppendBytes(b, blockBytes)
	b = protowire.AppendTag(b, 2, protowire.VarintType)
	b = protowire.AppendVarint(b, p.Round)
	b = protowire.AppendTag(b, 3, protowire.BytesType)
	b = protowire.AppendBytes(b, p.ProposerSig)
	return b, nil
}

// MarshalPrecommitVote deterministically encodes a PrecommitVote.
func MarshalPrecommitVote(v *types.PrecommitVote) ([]byte, error) {
	if v == nil {
		return nil, fmt.Errorf("precommit vote is nil")
	}
	var b []byte
	b = protowire.AppendTag(b, 1, protowire.BytesType)
	b = protowire.AppendBytes(b, v.BlockHash[:])
	b = protowire.AppendTag(b, 2, protowire.VarintType)
	b = protowire.AppendVarint(b, v.Height)
	b = protowire.AppendTag(b, 3, protowire.VarintType)
	b = protowire.AppendVarint(b, v.Round)
	b = protowire.AppendTag(b, 4, protowire.BytesType)
	b = protowire.AppendBytes(b, []byte(v.Validator))
	b = protowire.AppendTag(b, 5, protowire.BytesType)
	b = protowire.AppendBytes(b, v.Signature)
	return b, nil
}

// MarshalQuorumCertificate deterministically encodes a QuorumCertificate.
func MarshalQuorumCertificate(qc *types.QuorumCertificate) ([]byte, error) {
	if qc == nil {
		return nil, fmt.Errorf("quorum certificate is nil")
	}
	var b []byte
	b = protowire.AppendTag(b, 1, protowire.BytesType)
	b = protowire.AppendBytes(b, qc.BlockHash[:])
	b = protowire.AppendTag(b, 2, protowire.VarintType)
	b = protowire.AppendVarint(b, qc.Height)
	b = protowire.AppendTag(b, 3, protowire.VarintType)
	b = protowire.AppendVarint(b, qc.Round)
	b = protowire.AppendTag(b, 4, protowire.BytesType)
	b = protowire.AppendBytes(b, qc.SigBitmap)
	if len(qc.AggregatedSig) > 0 {
		b = protowire.AppendTag(b, 5, protowire.BytesType)
		b = protowire.AppendBytes(b, qc.AggregatedSig)
	}
	for _, sig := range qc.Signatures {
		b = protowire.AppendTag(b, 6, protowire.BytesType)
		b = protowire.AppendBytes(b, sig)
	}
	return b, nil
}

// MarshalViewChange deterministically encodes a ViewChange.
func MarshalViewChange(vc *types.ViewChange) ([]byte, error) {
	if vc == nil {
		return nil, fmt.Errorf("view change is nil")
	}
	var b []byte
	b = protowire.AppendTag(b, 1, protowire.VarintType)
	b = protowire.AppendVarint(b, vc.Height)
	b = protowire.AppendTag(b, 2, protowire.VarintType)
	b = protowire.AppendVarint(b, vc.Round)
	b = protowire.AppendTag(b, 3, protowire.BytesType)
	b = protowire.AppendBytes(b, []byte(vc.Validator))
	b = protowire.AppendTag(b, 4, protowire.BytesType)
	b = protowire.AppendBytes(b, vc.Signature)
	return b, nil
}

// MarshalUint64 deterministic encode uint64 as big-endian fixed64.
func MarshalUint64(v uint64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], v)
	return buf[:]
}
