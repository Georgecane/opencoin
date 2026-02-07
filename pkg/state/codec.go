package state

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protowire"

	"github.com/georgecane/opencoin/pkg/types"
)

func marshalAccount(acct *types.Account) ([]byte, error) {
	if acct == nil {
		return nil, fmt.Errorf("account is nil")
	}
	var b []byte
	b = protowire.AppendTag(b, 1, protowire.BytesType)
	b = protowire.AppendBytes(b, []byte(acct.Address))
	b = protowire.AppendTag(b, 2, protowire.VarintType)
	b = protowire.AppendVarint(b, acct.Balance)
	b = protowire.AppendTag(b, 3, protowire.VarintType)
	b = protowire.AppendVarint(b, acct.Nonce)
	b = protowire.AppendTag(b, 4, protowire.VarintType)
	b = protowire.AppendVarint(b, acct.Stake)
	b = protowire.AppendTag(b, 5, protowire.VarintType)
	b = protowire.AppendVarint(b, acct.RC)
	b = protowire.AppendTag(b, 6, protowire.VarintType)
	b = protowire.AppendVarint(b, acct.RCMax)
	b = protowire.AppendTag(b, 7, protowire.VarintType)
	b = protowire.AppendVarint(b, uint64(acct.LastRCEffectiveTime))
	if len(acct.Code) > 0 {
		b = protowire.AppendTag(b, 8, protowire.BytesType)
		b = protowire.AppendBytes(b, acct.Code)
	}
	if len(acct.PubKey) > 0 {
		b = protowire.AppendTag(b, 9, protowire.BytesType)
		b = protowire.AppendBytes(b, acct.PubKey)
	}
	return b, nil
}

func unmarshalAccount(b []byte) (*types.Account, error) {
	acct := &types.Account{}
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return nil, fmt.Errorf("invalid account tag")
		}
		b = b[n:]
		switch num {
		case 1:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("invalid address type")
			}
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid address")
			}
			acct.Address = types.Address(string(v))
			b = b[n:]
		case 2:
			if typ != protowire.VarintType {
				return nil, fmt.Errorf("invalid balance type")
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid balance")
			}
			acct.Balance = v
			b = b[n:]
		case 3:
			if typ != protowire.VarintType {
				return nil, fmt.Errorf("invalid nonce type")
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid nonce")
			}
			acct.Nonce = v
			b = b[n:]
		case 4:
			if typ != protowire.VarintType {
				return nil, fmt.Errorf("invalid stake type")
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid stake")
			}
			acct.Stake = v
			b = b[n:]
		case 5:
			if typ != protowire.VarintType {
				return nil, fmt.Errorf("invalid rc type")
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid rc")
			}
			acct.RC = v
			b = b[n:]
		case 6:
			if typ != protowire.VarintType {
				return nil, fmt.Errorf("invalid rc_max type")
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid rc_max")
			}
			acct.RCMax = v
			b = b[n:]
		case 7:
			if typ != protowire.VarintType {
				return nil, fmt.Errorf("invalid last_rc_effective_time type")
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid last_rc_effective_time")
			}
			acct.LastRCEffectiveTime = int64(v)
			b = b[n:]
		case 8:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("invalid code type")
			}
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid code")
			}
			acct.Code = append(acct.Code[:0], v...)
			b = b[n:]
		case 9:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("invalid pub_key type")
			}
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid pub_key")
			}
			acct.PubKey = append(acct.PubKey[:0], v...)
			b = b[n:]
		default:
			n := protowire.ConsumeFieldValue(num, typ, b)
			if n < 0 {
				return nil, fmt.Errorf("invalid account field")
			}
			b = b[n:]
		}
	}
	return acct, nil
}
