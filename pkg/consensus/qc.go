package consensus

import (
	"fmt"

	"github.com/georgecane/opencoin/pkg/crypto"
	"github.com/georgecane/opencoin/pkg/types"
)

// VerifyQC verifies a QC using validator index + bitmap.
func VerifyQC(qc *types.QuorumCertificate, set *types.ValidatorSet, verifier crypto.Verifier) error {
	if qc == nil || set == nil {
		return fmt.Errorf("nil qc or validator set")
	}
	if len(set.Validators) == 0 {
		return fmt.Errorf("empty validator set")
	}
	if len(qc.SigBitmap) != (len(set.Validators)+7)/8 {
		return fmt.Errorf("invalid bitmap length")
	}
	for i, v := range set.Validators {
		byteIdx := i / 8
		bitIdx := uint(i % 8)
		signed := (qc.SigBitmap[byteIdx] & (1 << bitIdx)) != 0
		if !signed {
			continue
		}
		if i >= len(qc.Signatures) || len(qc.Signatures[i]) == 0 {
			return fmt.Errorf("missing signature for validator index %d", i)
		}
		vote := &types.PrecommitVote{
			BlockHash: qc.BlockHash,
			Height:    qc.Height,
			Round:     qc.Round,
			Validator: v.Address,
		}
		msg, err := PrecommitSignBytes(vote)
		if err != nil {
			return err
		}
		if !verifier.Verify(msg, qc.Signatures[i], v.PublicKey) {
			return fmt.Errorf("invalid signature for validator index %d", i)
		}
	}
	return nil
}
