//go:build cgo

package crypto

/*
#cgo CFLAGS: -I${SRCDIR}/../../kyber/ref -O2
#include "api.h"

// Include reference implementation C sources so cgo compiles them into the
// Go package. We avoid duplicating symbols (e.g. `randombytes`) that are
// already compiled elsewhere; the set below provides the KEM implementation
// symbols required.
#include "../../kyber/ref/kem.c"
#include "../../kyber/ref/indcpa.c"
#include "../../kyber/ref/ntt.c"
#include "../../kyber/ref/poly.c"
#include "../../kyber/ref/polyvec.c"
#include "../../kyber/ref/reduce.c"
#include "../../kyber/ref/cbd.c"
#include "../../kyber/ref/verify.c"
#include "../../kyber/ref/fips202.c"
#include "../../kyber/ref/symmetric-shake.c"

#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// GenerateKyber768KeypairC generates a Kyber-768 keypair using the C reference implementation
func GenerateKyber768KeypairC() (pk, sk []byte, err error) {
	// Sizes defined in kyber/ref/api.h for kyber768
	const pkLenConst = 1184
	const skLenConst = 2400
	pkLen := pkLenConst
	skLen := skLenConst

	pk = make([]byte, pkLen)
	sk = make([]byte, skLen)

	ret := C.pqcrystals_kyber768_ref_keypair((*C.uint8_t)(unsafe.Pointer(&pk[0])), (*C.uint8_t)(unsafe.Pointer(&sk[0])))
	if ret != 0 {
		return nil, nil, fmt.Errorf("kyber keypair generation failed: %d", int(ret))
	}

	return pk, sk, nil
}

// EncapsulateKyber768C performs KEM encapsulation for Kyber-768 (returns ciphertext and shared secret)
func EncapsulateKyber768C(pk []byte) (ct, ss []byte, err error) {
	if len(pk) == 0 {
		return nil, nil, fmt.Errorf("public key empty")
	}

	// Sizes from kyber/ref/api.h
	const ctLenConst = 1088
	const ssLenConst = 32
	ctLen := ctLenConst
	ssLen := ssLenConst

	ct = make([]byte, ctLen)
	ss = make([]byte, ssLen)

	ret := C.pqcrystals_kyber768_ref_enc((*C.uint8_t)(unsafe.Pointer(&ct[0])), (*C.uint8_t)(unsafe.Pointer(&ss[0])), (*C.uint8_t)(unsafe.Pointer(&pk[0])))
	if ret != 0 {
		return nil, nil, fmt.Errorf("kyber encapsulation failed: %d", int(ret))
	}

	return ct, ss, nil
}

// DecapsulateKyber768C performs KEM decapsulation for Kyber-768 (returns shared secret)
func DecapsulateKyber768C(sk, ct []byte) (ss []byte, err error) {
	if len(sk) == 0 || len(ct) == 0 {
		return nil, fmt.Errorf("sk or ct empty")
	}

	const ssLenConst = 32
	ssLen := ssLenConst
	ss = make([]byte, ssLen)

	ret := C.pqcrystals_kyber768_ref_dec((*C.uint8_t)(unsafe.Pointer(&ss[0])), (*C.uint8_t)(unsafe.Pointer(&ct[0])), (*C.uint8_t)(unsafe.Pointer(&sk[0])))
	if ret != 0 {
		return nil, fmt.Errorf("kyber decapsulation failed: %d", int(ret))
	}

	return ss, nil
}
