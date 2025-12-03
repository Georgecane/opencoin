//go:build !cgo
// +build !cgo

package crypto

import "errors"

// Fallback implementations when CGO is not enabled.
// These return errors so callers can detect the absence of the native implementation.

// GenerateKeyPairC returns an error when CGO is disabled
func GenerateKeyPairC() (*KeyPair, error) {
	return nil, errors.New("Dilithium CGO support not enabled: build with CGO to use native Dilithium implementation")
}

// SignC returns an error when CGO is disabled
func SignC(privateKey, message []byte) ([]byte, error) {
	return nil, errors.New("Dilithium CGO support not enabled: build with CGO to use native Dilithium implementation")
}

// VerifyC returns false when CGO is disabled
func VerifyC(publicKey, message, signature []byte) bool {
	return false
}
