package tx

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protowire"

	"github.com/georgecane/opencoin/pkg/types"
)

// PayloadEnvelope wraps a payload with optional sender pubkey.
type PayloadEnvelope struct {
	Payload      Payload
	SenderPubKey []byte
}

// EncodePayload deterministically encodes a transaction payload using protobuf wire format.
func EncodePayload(p Payload, senderPubKey []byte) ([]byte, error) {
	if p == nil {
		return nil, fmt.Errorf("payload is nil")
	}

	var out []byte
	switch v := p.(type) {
	case Transfer:
		var err error
		out, err = encodeTransfer(v)
		if err != nil {
			return nil, err
		}
	case *Transfer:
		var err error
		out, err = encodeTransfer(*v)
		if err != nil {
			return nil, err
		}
	case StakeDelegate:
		var err error
		out, err = encodeStakeDelegate(v)
		if err != nil {
			return nil, err
		}
	case *StakeDelegate:
		var err error
		out, err = encodeStakeDelegate(*v)
		if err != nil {
			return nil, err
		}
	case StakeUndelegate:
		var err error
		out, err = encodeStakeUndelegate(v)
		if err != nil {
			return nil, err
		}
	case *StakeUndelegate:
		var err error
		out, err = encodeStakeUndelegate(*v)
		if err != nil {
			return nil, err
		}
	case ContractDeploy:
		var err error
		out, err = encodeContractDeploy(v)
		if err != nil {
			return nil, err
		}
	case *ContractDeploy:
		var err error
		out, err = encodeContractDeploy(*v)
		if err != nil {
			return nil, err
		}
	case ContractCall:
		var err error
		out, err = encodeContractCall(v)
		if err != nil {
			return nil, err
		}
	case *ContractCall:
		var err error
		out, err = encodeContractCall(*v)
		if err != nil {
			return nil, err
		}
	case GovernanceProposal:
		var err error
		out, err = encodeGovernanceProposal(v)
		if err != nil {
			return nil, err
		}
	case *GovernanceProposal:
		var err error
		out, err = encodeGovernanceProposal(*v)
		if err != nil {
			return nil, err
		}
	case GovernanceVote:
		var err error
		out, err = encodeGovernanceVote(v)
		if err != nil {
			return nil, err
		}
	case *GovernanceVote:
		var err error
		out, err = encodeGovernanceVote(*v)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown payload type %T", p)
	}
	if len(senderPubKey) > 0 {
		out = protowire.AppendTag(out, 8, protowire.BytesType)
		out = protowire.AppendBytes(out, senderPubKey)
	}
	return out, nil
}

// DecodePayload decodes a TransactionPayload into a typed payload.
func DecodePayload(payload []byte) (*PayloadEnvelope, error) {
	if len(payload) == 0 {
		return nil, fmt.Errorf("empty payload")
	}
	env := &PayloadEnvelope{}
	var fieldNum protowire.Number
	var typ protowire.Type
	var n int
	for len(payload) > 0 {
		fieldNum, typ, n = protowire.ConsumeTag(payload)
		if n < 0 {
			return nil, fmt.Errorf("invalid payload tag")
		}
		payload = payload[n:]

		switch fieldNum {
		case 1:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("unexpected transfer wire type %v", typ)
			}
			var b []byte
			b, n = protowire.ConsumeBytes(payload)
			if n < 0 {
				return nil, fmt.Errorf("invalid transfer bytes")
			}
			payload = payload[n:]
			if env.Payload != nil {
				return nil, fmt.Errorf("duplicate payload")
			}
			p, err := decodeTransfer(b)
			if err != nil {
				return nil, err
			}
			env.Payload = p
		case 2:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("unexpected stake delegate wire type %v", typ)
			}
			var b []byte
			b, n = protowire.ConsumeBytes(payload)
			if n < 0 {
				return nil, fmt.Errorf("invalid stake delegate bytes")
			}
			payload = payload[n:]
			if env.Payload != nil {
				return nil, fmt.Errorf("duplicate payload")
			}
			p, err := decodeStakeDelegate(b)
			if err != nil {
				return nil, err
			}
			env.Payload = p
		case 3:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("unexpected stake undelegate wire type %v", typ)
			}
			var b []byte
			b, n = protowire.ConsumeBytes(payload)
			if n < 0 {
				return nil, fmt.Errorf("invalid stake undelegate bytes")
			}
			payload = payload[n:]
			if env.Payload != nil {
				return nil, fmt.Errorf("duplicate payload")
			}
			p, err := decodeStakeUndelegate(b)
			if err != nil {
				return nil, err
			}
			env.Payload = p
		case 4:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("unexpected contract deploy wire type %v", typ)
			}
			var b []byte
			b, n = protowire.ConsumeBytes(payload)
			if n < 0 {
				return nil, fmt.Errorf("invalid contract deploy bytes")
			}
			payload = payload[n:]
			if env.Payload != nil {
				return nil, fmt.Errorf("duplicate payload")
			}
			p, err := decodeContractDeploy(b)
			if err != nil {
				return nil, err
			}
			env.Payload = p
		case 5:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("unexpected contract call wire type %v", typ)
			}
			var b []byte
			b, n = protowire.ConsumeBytes(payload)
			if n < 0 {
				return nil, fmt.Errorf("invalid contract call bytes")
			}
			payload = payload[n:]
			if env.Payload != nil {
				return nil, fmt.Errorf("duplicate payload")
			}
			p, err := decodeContractCall(b)
			if err != nil {
				return nil, err
			}
			env.Payload = p
		case 6:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("unexpected governance proposal wire type %v", typ)
			}
			var b []byte
			b, n = protowire.ConsumeBytes(payload)
			if n < 0 {
				return nil, fmt.Errorf("invalid governance proposal bytes")
			}
			payload = payload[n:]
			if env.Payload != nil {
				return nil, fmt.Errorf("duplicate payload")
			}
			p, err := decodeGovernanceProposal(b)
			if err != nil {
				return nil, err
			}
			env.Payload = p
		case 7:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("unexpected governance vote wire type %v", typ)
			}
			var b []byte
			b, n = protowire.ConsumeBytes(payload)
			if n < 0 {
				return nil, fmt.Errorf("invalid governance vote bytes")
			}
			payload = payload[n:]
			if env.Payload != nil {
				return nil, fmt.Errorf("duplicate payload")
			}
			p, err := decodeGovernanceVote(b)
			if err != nil {
				return nil, err
			}
			env.Payload = p
		case 8:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("unexpected sender_pubkey wire type %v", typ)
			}
			if len(env.SenderPubKey) > 0 {
				return nil, fmt.Errorf("duplicate sender_pubkey")
			}
			var b []byte
			b, n = protowire.ConsumeBytes(payload)
			if n < 0 {
				return nil, fmt.Errorf("invalid sender_pubkey bytes")
			}
			env.SenderPubKey = append(env.SenderPubKey[:0], b...)
			payload = payload[n:]
		default:
			n = protowire.ConsumeFieldValue(fieldNum, typ, payload)
			if n < 0 {
				return nil, fmt.Errorf("invalid payload field %d", fieldNum)
			}
			payload = payload[n:]
		}
	}
	if env.Payload == nil {
		return nil, fmt.Errorf("payload not found")
	}
	return env, nil
}

func encodeTransfer(t Transfer) ([]byte, error) {
	var inner []byte
	inner = protowire.AppendTag(inner, 1, protowire.BytesType)
	inner = protowire.AppendBytes(inner, []byte(t.To))
	inner = protowire.AppendTag(inner, 2, protowire.VarintType)
	inner = protowire.AppendVarint(inner, t.Amount)

	var out []byte
	out = protowire.AppendTag(out, 1, protowire.BytesType)
	out = protowire.AppendBytes(out, inner)
	return out, nil
}

func encodeStakeDelegate(t StakeDelegate) ([]byte, error) {
	var inner []byte
	inner = protowire.AppendTag(inner, 1, protowire.BytesType)
	inner = protowire.AppendBytes(inner, []byte(t.Validator))
	inner = protowire.AppendTag(inner, 2, protowire.VarintType)
	inner = protowire.AppendVarint(inner, t.Amount)

	var out []byte
	out = protowire.AppendTag(out, 2, protowire.BytesType)
	out = protowire.AppendBytes(out, inner)
	return out, nil
}

func encodeStakeUndelegate(t StakeUndelegate) ([]byte, error) {
	var inner []byte
	inner = protowire.AppendTag(inner, 1, protowire.BytesType)
	inner = protowire.AppendBytes(inner, []byte(t.Validator))
	inner = protowire.AppendTag(inner, 2, protowire.VarintType)
	inner = protowire.AppendVarint(inner, t.Amount)

	var out []byte
	out = protowire.AppendTag(out, 3, protowire.BytesType)
	out = protowire.AppendBytes(out, inner)
	return out, nil
}

func encodeContractDeploy(t ContractDeploy) ([]byte, error) {
	var inner []byte
	inner = protowire.AppendTag(inner, 1, protowire.BytesType)
	inner = protowire.AppendBytes(inner, t.WASMCode)
	inner = protowire.AppendTag(inner, 2, protowire.BytesType)
	inner = protowire.AppendBytes(inner, t.Salt)

	var out []byte
	out = protowire.AppendTag(out, 4, protowire.BytesType)
	out = protowire.AppendBytes(out, inner)
	return out, nil
}

func encodeContractCall(t ContractCall) ([]byte, error) {
	var inner []byte
	inner = protowire.AppendTag(inner, 1, protowire.BytesType)
	inner = protowire.AppendBytes(inner, []byte(t.Address))
	inner = protowire.AppendTag(inner, 2, protowire.BytesType)
	inner = protowire.AppendBytes(inner, []byte(t.Method))
	for _, arg := range t.Args {
		inner = protowire.AppendTag(inner, 3, protowire.BytesType)
		inner = protowire.AppendBytes(inner, arg)
	}

	var out []byte
	out = protowire.AppendTag(out, 5, protowire.BytesType)
	out = protowire.AppendBytes(out, inner)
	return out, nil
}

func encodeGovernanceProposal(t GovernanceProposal) ([]byte, error) {
	var inner []byte
	inner = protowire.AppendTag(inner, 1, protowire.BytesType)
	inner = protowire.AppendBytes(inner, []byte(t.Title))
	inner = protowire.AppendTag(inner, 2, protowire.BytesType)
	inner = protowire.AppendBytes(inner, []byte(t.Description))
	inner = protowire.AppendTag(inner, 3, protowire.BytesType)
	inner = protowire.AppendBytes(inner, []byte(t.ParamKey))
	inner = protowire.AppendTag(inner, 4, protowire.BytesType)
	inner = protowire.AppendBytes(inner, []byte(t.ParamValue))

	var out []byte
	out = protowire.AppendTag(out, 6, protowire.BytesType)
	out = protowire.AppendBytes(out, inner)
	return out, nil
}

func encodeGovernanceVote(t GovernanceVote) ([]byte, error) {
	var inner []byte
	inner = protowire.AppendTag(inner, 1, protowire.VarintType)
	inner = protowire.AppendVarint(inner, t.ProposalID)
	inner = protowire.AppendTag(inner, 2, protowire.VarintType)
	inner = protowire.AppendVarint(inner, uint64(t.Option))

	var out []byte
	out = protowire.AppendTag(out, 7, protowire.BytesType)
	out = protowire.AppendBytes(out, inner)
	return out, nil
}

func decodeTransfer(b []byte) (Payload, error) {
	var out Transfer
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return nil, fmt.Errorf("invalid transfer tag")
		}
		b = b[n:]
		switch num {
		case 1:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("invalid transfer to type")
			}
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid transfer to")
			}
			out.To = types.Address(string(v))
			b = b[n:]
		case 2:
			if typ != protowire.VarintType {
				return nil, fmt.Errorf("invalid transfer amount type")
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid transfer amount")
			}
			out.Amount = v
			b = b[n:]
		default:
			n := protowire.ConsumeFieldValue(num, typ, b)
			if n < 0 {
				return nil, fmt.Errorf("invalid transfer field")
			}
			b = b[n:]
		}
	}
	return out, nil
}

func decodeStakeDelegate(b []byte) (Payload, error) {
	var out StakeDelegate
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return nil, fmt.Errorf("invalid stake delegate tag")
		}
		b = b[n:]
		switch num {
		case 1:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("invalid validator type")
			}
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid validator")
			}
			out.Validator = types.Address(string(v))
			b = b[n:]
		case 2:
			if typ != protowire.VarintType {
				return nil, fmt.Errorf("invalid amount type")
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid amount")
			}
			out.Amount = v
			b = b[n:]
		default:
			n := protowire.ConsumeFieldValue(num, typ, b)
			if n < 0 {
				return nil, fmt.Errorf("invalid delegate field")
			}
			b = b[n:]
		}
	}
	return out, nil
}

func decodeStakeUndelegate(b []byte) (Payload, error) {
	var out StakeUndelegate
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return nil, fmt.Errorf("invalid stake undelegate tag")
		}
		b = b[n:]
		switch num {
		case 1:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("invalid validator type")
			}
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid validator")
			}
			out.Validator = types.Address(string(v))
			b = b[n:]
		case 2:
			if typ != protowire.VarintType {
				return nil, fmt.Errorf("invalid amount type")
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid amount")
			}
			out.Amount = v
			b = b[n:]
		default:
			n := protowire.ConsumeFieldValue(num, typ, b)
			if n < 0 {
				return nil, fmt.Errorf("invalid undelegate field")
			}
			b = b[n:]
		}
	}
	return out, nil
}

func decodeContractDeploy(b []byte) (Payload, error) {
	var out ContractDeploy
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return nil, fmt.Errorf("invalid contract deploy tag")
		}
		b = b[n:]
		switch num {
		case 1:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("invalid wasm code type")
			}
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid wasm code")
			}
			out.WASMCode = append(out.WASMCode[:0], v...)
			b = b[n:]
		case 2:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("invalid salt type")
			}
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid salt")
			}
			out.Salt = append(out.Salt[:0], v...)
			b = b[n:]
		default:
			n := protowire.ConsumeFieldValue(num, typ, b)
			if n < 0 {
				return nil, fmt.Errorf("invalid deploy field")
			}
			b = b[n:]
		}
	}
	return out, nil
}

func decodeContractCall(b []byte) (Payload, error) {
	var out ContractCall
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return nil, fmt.Errorf("invalid contract call tag")
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
			out.Address = types.Address(string(v))
			b = b[n:]
		case 2:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("invalid method type")
			}
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid method")
			}
			out.Method = string(v)
			b = b[n:]
		case 3:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("invalid args type")
			}
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid arg")
			}
			out.Args = append(out.Args, append([]byte(nil), v...))
			b = b[n:]
		default:
			n := protowire.ConsumeFieldValue(num, typ, b)
			if n < 0 {
				return nil, fmt.Errorf("invalid call field")
			}
			b = b[n:]
		}
	}
	return out, nil
}

func decodeGovernanceProposal(b []byte) (Payload, error) {
	var out GovernanceProposal
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return nil, fmt.Errorf("invalid governance proposal tag")
		}
		b = b[n:]
		switch num {
		case 1:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("invalid title type")
			}
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid title")
			}
			out.Title = string(v)
			b = b[n:]
		case 2:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("invalid description type")
			}
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid description")
			}
			out.Description = string(v)
			b = b[n:]
		case 3:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("invalid param key type")
			}
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid param key")
			}
			out.ParamKey = string(v)
			b = b[n:]
		case 4:
			if typ != protowire.BytesType {
				return nil, fmt.Errorf("invalid param value type")
			}
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid param value")
			}
			out.ParamValue = string(v)
			b = b[n:]
		default:
			n := protowire.ConsumeFieldValue(num, typ, b)
			if n < 0 {
				return nil, fmt.Errorf("invalid governance field")
			}
			b = b[n:]
		}
	}
	return out, nil
}

func decodeGovernanceVote(b []byte) (Payload, error) {
	var out GovernanceVote
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return nil, fmt.Errorf("invalid governance vote tag")
		}
		b = b[n:]
		switch num {
		case 1:
			if typ != protowire.VarintType {
				return nil, fmt.Errorf("invalid proposal id type")
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid proposal id")
			}
			out.ProposalID = v
			b = b[n:]
		case 2:
			if typ != protowire.VarintType {
				return nil, fmt.Errorf("invalid vote option type")
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return nil, fmt.Errorf("invalid vote option")
			}
			out.Option = types.VoteOption(v)
			b = b[n:]
		default:
			n := protowire.ConsumeFieldValue(num, typ, b)
			if n < 0 {
				return nil, fmt.Errorf("invalid vote field")
			}
			b = b[n:]
		}
	}
	return out, nil
}
