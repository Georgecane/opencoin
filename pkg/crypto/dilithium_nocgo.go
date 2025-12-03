//go:build !cgo
// +build !cgo

package crypto

import (
	"errors"
	"fmt"
)

// Fallback implementations when CGO is not enabled.
// These return explicit errors so callers can detect the absence of the native implementation.

var errDilithiumCGODisabled = errors.New("dilithium: CGO not enabled; build with CGO to use native Dilithium implementation")

// GenerateKeyPairC returns an error when CGO is disabled
func GenerateKeyPairC() (*KeyPair, error) {
	return nil, errDilithiumCGODisabled
}

// SignC returns an error when CGO is disabled
func SignC(privateKey, message []byte) ([]byte, error) {
	if !CGOEnabled {
		return nil, errDilithiumCGODisabled
	}
	return nil, fmt.Errorf("dilithium: unexpected code path in non-cgo build")
}

// VerifyC returns false when CGO is disabled
func VerifyC(publicKey, message, signature []byte) bool {
	return false
}
