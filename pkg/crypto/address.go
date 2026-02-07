package crypto

import (
	"crypto/sha256"
	"fmt"

	"github.com/btcsuite/btcd/btcutil/bech32"
)

const (
	AddressHRP      = "ocn"
	AddressHashSize = 20
)

// AddressFromPubKey derives a bech32 address from an Ed25519 public key.
func AddressFromPubKey(pub []byte) (string, error) {
	if len(pub) == 0 {
		return "", fmt.Errorf("empty public key")
	}
	sum := sha256.Sum256(pub)
	addrBytes := sum[:AddressHashSize]
	conv, err := bech32.ConvertBits(addrBytes, 8, 5, true)
	if err != nil {
		return "", fmt.Errorf("bech32 convert: %w", err)
	}
	addr, err := bech32.Encode(AddressHRP, conv)
	if err != nil {
		return "", fmt.Errorf("bech32 encode: %w", err)
	}
	return addr, nil
}

// DecodeAddress decodes a bech32 address and returns the 20-byte hash.
func DecodeAddress(addr string) ([]byte, error) {
	hrp, data, err := bech32.Decode(addr)
	if err != nil {
		return nil, fmt.Errorf("bech32 decode: %w", err)
	}
	if hrp != AddressHRP {
		return nil, fmt.Errorf("invalid address hrp: %s", hrp)
	}
	out, err := bech32.ConvertBits(data, 5, 8, false)
	if err != nil {
		return nil, fmt.Errorf("bech32 convert: %w", err)
	}
	if len(out) != AddressHashSize {
		return nil, fmt.Errorf("invalid address length: %d", len(out))
	}
	return out, nil
}
