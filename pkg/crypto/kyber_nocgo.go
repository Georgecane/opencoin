//go:build !cgo

package crypto

import "errors"

// Fallback implementations when cgo is disabled. These return explicit errors.

func GenerateKyber768KeypairC() (pk, sk []byte, err error) {
	return nil, nil, errors.New("kyber: CGO disabled; no kyber implementation available")
}

func EncapsulateKyber768C(pk []byte) (ct, ss []byte, err error) {
	return nil, nil, errors.New("kyber: CGO disabled; no kyber implementation available")
}

func DecapsulateKyber768C(sk, ct []byte) (ss []byte, err error) {
	return nil, errors.New("kyber: CGO disabled; no kyber implementation available")
}
