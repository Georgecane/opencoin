package crypto

import (
	"crypto/sha256"
	"fmt"
)

// KeyPair represents a Dilithium key pair
type KeyPair struct {
	PublicKey  []byte
	PrivateKey []byte
}

// Signer interface for signing operations
type Signer interface {
	Sign(message []byte) ([]byte, error)
	PublicKey() []byte
}

// Verifier interface for signature verification
type Verifier interface {
	Verify(message []byte, signature []byte, publicKey []byte) bool
}

// DilithiumSigner implements the Signer interface using Dilithium
type DilithiumSigner struct {
	publicKey  []byte
	privateKey []byte
}

// NewDilithiumSigner creates a new Dilithium signer
func NewDilithiumSigner(publicKey, privateKey []byte) *DilithiumSigner {
	return &DilithiumSigner{
		publicKey:  publicKey,
		privateKey: privateKey,
	}
}

// Sign signs a message with the private key
func (ds *DilithiumSigner) Sign(message []byte) ([]byte, error) {
	if len(ds.privateKey) == 0 {
		return nil, fmt.Errorf("private key not set")
	}

	// Use CGO wrapper if available
	sig, err := SignC(ds.privateKey, message)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

// PublicKey returns the public key
func (ds *DilithiumSigner) PublicKey() []byte {
	return ds.publicKey
}

// DilithiumVerifier implements the Verifier interface
type DilithiumVerifier struct {
}

// NewDilithiumVerifier creates a new Dilithium verifier
func NewDilithiumVerifier() *DilithiumVerifier {
	return &DilithiumVerifier{}
}

// Verify verifies a signature
func (dv *DilithiumVerifier) Verify(message []byte, signature []byte, publicKey []byte) bool {
	if len(publicKey) == 0 || len(signature) == 0 {
		return false
	}

	return VerifyC(publicKey, message, signature)
}

// GenerateKeyPair generates a new Dilithium key pair
// Note: This will need CGO binding to the actual C implementation
func GenerateKeyPair() (*KeyPair, error) {
	return GenerateKeyPairC()
}

// Hash represents a SHA-256 hash
func Hash(data []byte) []byte {
	h := sha256.Sum256(data)
	out := make([]byte, sha256.Size)
	copy(out, h[:])
	return out
}
