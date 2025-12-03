//go:build cgo

package crypto

/*
#cgo CFLAGS: -I${SRCDIR}/../../dilithium/ref -O2
#include "api.h"

// Include reference implementation C sources so cgo compiles them into the
// Go package. This keeps the build self-contained for development.
#include "../../dilithium/ref/sign.c"
#include "../../dilithium/ref/packing.c"
#include "../../dilithium/ref/poly.c"
#include "../../dilithium/ref/polyvec.c"
#include "../../dilithium/ref/ntt.c"
#include "../../dilithium/ref/randombytes.c"
#include "../../dilithium/ref/reduce.c"
#include "../../dilithium/ref/rounding.c"
#include "../../dilithium/ref/fips202.c"
#include "../../dilithium/ref/symmetric-shake.c"

#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// GenerateKeyPairC generates a Dilithium2 key pair using the C reference implementation
func GenerateKeyPairC() (*KeyPair, error) {
	// Sizes taken from dilithium/ref/api.h for Dilithium2
	const pkLenConst = 1312
	const skLenConst = 2560

	pk := make([]byte, pkLenConst)
	sk := make([]byte, skLenConst)

	ret := C.pqcrystals_dilithium2_ref_keypair((*C.uint8_t)(unsafe.Pointer(&pk[0])), (*C.uint8_t)(unsafe.Pointer(&sk[0])))
	if ret != 0 {
		return nil, fmt.Errorf("dilithium keypair generation failed: %d", int(ret))
	}

	return &KeyPair{PublicKey: pk, PrivateKey: sk}, nil
}

// SignC signs a message using the provided Dilithium private key via the C implementation
func SignC(privateKey, message []byte) ([]byte, error) {
	if len(privateKey) == 0 {
		return nil, fmt.Errorf("private key is empty")
	}

	// Max signature size from dilithium/ref/api.h
	const sigMaxConst = 2420
	sig := make([]byte, sigMaxConst)
	var siglen C.size_t

	var mptr *C.uint8_t
	if len(message) > 0 {
		mptr = (*C.uint8_t)(unsafe.Pointer(&message[0]))
	} else {
		mptr = nil
	}

	skptr := (*C.uint8_t)(unsafe.Pointer(&privateKey[0]))

	ret := C.pqcrystals_dilithium2_ref_signature((*C.uint8_t)(unsafe.Pointer(&sig[0])), &siglen,
		mptr, C.size_t(len(message)), nil, 0, skptr)
	if ret != 0 {
		return nil, fmt.Errorf("dilithium sign failed: %d", int(ret))
	}

	return sig[:int(siglen)], nil
}

// VerifyC verifies a Dilithium signature using the C implementation
func VerifyC(publicKey, message, signature []byte) bool {
	if len(publicKey) == 0 || len(signature) == 0 {
		return false
	}

	var mptr *C.uint8_t
	if len(message) > 0 {
		mptr = (*C.uint8_t)(unsafe.Pointer(&message[0]))
	} else {
		mptr = nil
	}

	ret := C.pqcrystals_dilithium2_ref_verify((*C.uint8_t)(unsafe.Pointer(&signature[0])), C.size_t(len(signature)),
		mptr, C.size_t(len(message)), nil, 0, (*C.uint8_t)(unsafe.Pointer(&publicKey[0])))

	return ret == 0
}
