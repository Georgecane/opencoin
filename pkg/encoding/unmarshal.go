package encoding

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protowire"

	"github.com/georgecane/opencoin/pkg/types"
)

// UnmarshalTransaction decodes a Transaction from protobuf wire format.
func UnmarshalTransaction(b []byte) (*types.Transaction, error) {
	if len(b) == 0 {
		return nil, fmt.Errorf("empty transaction")
	}
	var tx types.Transaction
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return nil, fmt.Errorf("invalid transaction tag")
		}
		b = b[n:]
		switch num {
		case 1:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("invalid from type")
			}
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid from")
			}
			tx.From = types.Address(string(v))
			b = b[n:]
		case 2:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("invalid to type")
			}
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid to")
			}
			tx.To = types.Address(string(v))
			b = b[n:]
		case 3:
			if typ != protowire.VarintType {
				return nil, fmt.Errorf("invalid nonce type")
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid nonce")
			}
			tx.Nonce = v
			b = b[n:]
		case 4:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("invalid payload type")
			}
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid payload")
			}
			tx.Payload = append(tx.Payload[:0], v...)
			b = b[n:]
		case 5:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("invalid signature type")
			}
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid signature")
			}
			tx.Signature = append(tx.Signature[:0], v...)
			b = b[n:]
		default:
			n = protowire.ConsumeFieldValue(num, typ, b)
			if n < 0 {
				return nil, fmt.Errorf("invalid transaction field %d", num)
			}
			b = b[n:]
		}
	}
	return &tx, nil
}

// UnmarshalBlock decodes a Block from protobuf wire format.
func UnmarshalBlock(b []byte) (*types.Block, error) {
	if len(b) == 0 {
		return nil, fmt.Errorf("empty block")
	}
	var block types.Block
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return nil, fmt.Errorf("invalid block tag")
		}
		b = b[n:]
		switch num {
		case 1:
			if typ != protowire.VarintType {
				return nil, fmt.Errorf("invalid height type")
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid height")
			}
			block.Height = v
			b = b[n:]
		case 2:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("invalid prev_hash type")
			}
			v, n := protowire.ConsumeBytes(b)
			if n < 0 || len(v) != len(block.PrevHash) {
				return nil, fmt.Errorf("invalid prev_hash")
			}
			copy(block.PrevHash[:], v)
			b = b[n:]
		case 3:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("invalid state_root type")
			}
			v, n := protowire.ConsumeBytes(b)
			if n < 0 || len(v) != len(block.StateRoot) {
				return nil, fmt.Errorf("invalid state_root")
			}
			copy(block.StateRoot[:], v)
			b = b[n:]
		case 4:
			if typ != protowire.VarintType {
				return nil, fmt.Errorf("invalid timestamp type")
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid timestamp")
			}
			block.Timestamp = int64(v)
			b = b[n:]
		case 5:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("invalid proposer type")
			}
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid proposer")
			}
			block.Proposer = types.Address(string(v))
			b = b[n:]
		case 6:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("invalid tx type")
			}
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid tx bytes")
			}
			tx, err := UnmarshalTransaction(v)
			if err != nil {
				return nil, err
			}
			block.Transactions = append(block.Transactions, tx)
			b = b[n:]
		case 7:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("invalid validator sig type")
			}
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid validator sig")
			}
			block.ValidatorSigs = append(block.ValidatorSigs, append([]byte(nil), v...))
			b = b[n:]
		default:
			n = protowire.ConsumeFieldValue(num, typ, b)
			if n < 0 {
				return nil, fmt.Errorf("invalid block field %d", num)
			}
			b = b[n:]
		}
	}
	return &block, nil
}
