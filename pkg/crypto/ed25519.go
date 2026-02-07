package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
)

// Ed25519KeyPair holds an Ed25519 keypair.
type Ed25519KeyPair struct {
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

// GenerateEd25519 generates a new Ed25519 keypair.
func GenerateEd25519() (*Ed25519KeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("ed25519 generate: %w", err)
	}
	return &Ed25519KeyPair{PublicKey: pub, PrivateKey: priv}, nil
}

// SignEd25519 signs a message using the private key.
func SignEd25519(priv ed25519.PrivateKey, msg []byte) ([]byte, error) {
	if len(priv) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid ed25519 private key")
	}
	return ed25519.Sign(priv, msg), nil
}

// VerifyEd25519 verifies a signature using the public key.
func VerifyEd25519(pub ed25519.PublicKey, msg, sig []byte) bool {
	if len(pub) != ed25519.PublicKeySize {
		return false
	}
	return ed25519.Verify(pub, msg, sig)
}
