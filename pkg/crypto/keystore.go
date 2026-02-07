package crypto

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
)

type keyFile struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

// SaveKeyPair saves a Dilithium keypair to disk.
func SaveKeyPair(path string, kp *KeyPair) error {
	if kp == nil {
		return fmt.Errorf("keypair is nil")
	}
	data := keyFile{
		PublicKey:  base64.StdEncoding.EncodeToString(kp.PublicKey),
		PrivateKey: base64.StdEncoding.EncodeToString(kp.PrivateKey),
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o600)
}

// LoadKeyPair loads a Dilithium keypair from disk.
func LoadKeyPair(path string) (*KeyPair, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var data keyFile
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	pub, err := base64.StdEncoding.DecodeString(data.PublicKey)
	if err != nil {
		return nil, err
	}
	priv, err := base64.StdEncoding.DecodeString(data.PrivateKey)
	if err != nil {
		return nil, err
	}
	return &KeyPair{PublicKey: pub, PrivateKey: priv}, nil
}

// SaveEd25519 saves an Ed25519 keypair to disk.
func SaveEd25519(path string, kp *Ed25519KeyPair) error {
	if kp == nil {
		return fmt.Errorf("keypair is nil")
	}
	data := keyFile{
		PublicKey:  base64.StdEncoding.EncodeToString(kp.PublicKey),
		PrivateKey: base64.StdEncoding.EncodeToString(kp.PrivateKey),
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o600)
}

// LoadEd25519 loads an Ed25519 keypair from disk.
func LoadEd25519(path string) (*Ed25519KeyPair, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var data keyFile
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	pub, err := base64.StdEncoding.DecodeString(data.PublicKey)
	if err != nil {
		return nil, err
	}
	priv, err := base64.StdEncoding.DecodeString(data.PrivateKey)
	if err != nil {
		return nil, err
	}
	return &Ed25519KeyPair{PublicKey: pub, PrivateKey: priv}, nil
}
