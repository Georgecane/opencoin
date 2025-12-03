//go:build !cgo

package crypto

import (
	"errors"
)

var errKyberCGODisabled = errors.New("kyber: CGO not enabled; build with CGO to use native Kyber implementation")

func GenerateKyber768KeypairC() (pk, sk []byte, err error) {
	return nil, nil, errKyberCGODisabled
}

func EncapsulateKyber768C(pk []byte) (ct, ss []byte, err error) {
	return nil, nil, errKyberCGODisabled
}

func DecapsulateKyber768C(sk, ct []byte) (ss []byte, err error) {
	return nil, errKyberCGODisabled
}
