package tx

import (
	"bytes"
	"crypto/ed25519"
	"fmt"

	"github.com/georgecane/opencoin/pkg/crypto"
	"github.com/georgecane/opencoin/pkg/types"
)

// ResolveSenderPubKey validates sender pubkey rules and returns the pubkey to verify against.
// register is true when the pubkey should be stored in account state (first spend).
func ResolveSenderPubKey(txn *types.Transaction, stored []byte, payloadPubKey []byte) (pubKey []byte, register bool, err error) {
	if txn == nil {
		return nil, false, fmt.Errorf("tx is nil")
	}
	if len(payloadPubKey) > 0 && len(payloadPubKey) != ed25519.PublicKeySize {
		return nil, false, fmt.Errorf("invalid sender_pubkey length")
	}
	if len(stored) > 0 && len(stored) != ed25519.PublicKeySize {
		return nil, false, fmt.Errorf("invalid stored pubkey length")
	}

	if len(stored) == 0 {
		if len(payloadPubKey) == 0 {
			return nil, false, fmt.Errorf("sender_pubkey required for first spend")
		}
		if err := ensureAddressMatches(txn.From, payloadPubKey); err != nil {
			return nil, false, err
		}
		return append([]byte(nil), payloadPubKey...), true, nil
	}

	if len(payloadPubKey) > 0 && !bytes.Equal(payloadPubKey, stored) {
		return nil, false, fmt.Errorf("sender_pubkey does not match registered key")
	}
	if err := ensureAddressMatches(txn.From, stored); err != nil {
		return nil, false, err
	}
	return append([]byte(nil), stored...), false, nil
}

// VerifySignature verifies the Ed25519 signature for a transaction.
func VerifySignature(txn *types.Transaction, pubKey []byte) error {
	if txn == nil {
		return fmt.Errorf("tx is nil")
	}
	if len(pubKey) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid pubkey length")
	}
	if len(txn.Signature) != ed25519.SignatureSize {
		return fmt.Errorf("invalid signature length")
	}
	signBytes, err := SigningBytes(txn)
	if err != nil {
		return err
	}
	if !crypto.VerifyEd25519(ed25519.PublicKey(pubKey), signBytes, txn.Signature) {
		return fmt.Errorf("invalid signature")
	}
	return nil
}

func ensureAddressMatches(addr types.Address, pubKey []byte) error {
	derived, err := crypto.AddressFromPubKey(pubKey)
	if err != nil {
		return fmt.Errorf("derive address: %w", err)
	}
	if types.Address(derived) != addr {
		return fmt.Errorf("sender address mismatch")
	}
	return nil
}
