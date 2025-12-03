package crypto

import "fmt"

// KeyPairKyber holds a Kyber public/private key pair (Kyber-768 by default)
type KeyPairKyber struct {
	PublicKey  []byte
	PrivateKey []byte
}

// GenerateKeyPairKyber768 creates a Kyber-768 key pair.
func GenerateKeyPairKyber768() (*KeyPairKyber, error) {
	pk, sk, err := GenerateKyber768KeypairC()
	if err != nil {
		return nil, fmt.Errorf("GenerateKeyPairKyber768: %w", err)
	}
	return &KeyPairKyber{PublicKey: pk, PrivateKey: sk}, nil
}

// EncapsulateKyber768 creates a ciphertext and a shared secret for the given public key.
func EncapsulateKyber768(pk []byte) (ct, ss []byte, err error) {
	return EncapsulateKyber768C(pk)
}

// DecapsulateKyber768 recovers the shared secret from the ciphertext using the private key.
func DecapsulateKyber768(sk, ct []byte) (ss []byte, err error) {
	return DecapsulateKyber768C(sk, ct)
}
