package tx

import (
	"github.com/georgecane/opencoin/pkg/encoding"
	"github.com/georgecane/opencoin/pkg/types"
)

// SigningBytes returns deterministic bytes for transaction signing.
func SigningBytes(tx *types.Transaction) ([]byte, error) {
	if tx == nil {
		return nil, nil
	}
	copyTx := *tx
	copyTx.Signature = nil
	return encoding.MarshalTransaction(&copyTx)
}
