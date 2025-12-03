package main

import (
	"fmt"
	"os"

	"github.com/georgecane/opencoin/pkg/crypto"
)

func main() {
	kp, err := crypto.GenerateKeyPair()
	if err != nil {
		fmt.Println("Key generation failed:", err)
		os.Exit(1)
	}

	fmt.Printf("Generated keypair: pk=%d sk=%d\n", len(kp.PublicKey), len(kp.PrivateKey))

	msg := []byte("Hello, OpenCoin!")
	sig, err := crypto.SignC(kp.PrivateKey, msg)
	if err != nil {
		fmt.Println("Sign failed:", err)
		os.Exit(1)
	}

	ok := crypto.VerifyC(kp.PublicKey, msg, sig)
	fmt.Println("Verify result:", ok)
}
