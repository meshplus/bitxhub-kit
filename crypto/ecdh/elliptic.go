package ecdh

import (
	"crypto/elliptic"
	"fmt"
	"math/big"

	"github.com/meshplus/bitxhub-kit/crypto"
)

type ellipticECDH struct {
	curve elliptic.Curve
}

func NewEllipticECDH(c elliptic.Curve) (KeyExchange, error) {
	if c == nil {
		return nil, fmt.Errorf("invalid curve")
	}

	return ellipticECDH{curve: c}, nil
}

func (e ellipticECDH) Check(peerPubkey []byte) error {
	if len(peerPubkey) == 0 {
		return fmt.Errorf("empty public key byte")
	}
	if len(peerPubkey) != 65 {
		return fmt.Errorf("public key data length is not 65")
	}
	x, y := getXYFromPub(peerPubkey)
	if !e.curve.IsOnCurve(x, y) {
		return fmt.Errorf("peer's public key is not on curve")
	}

	return nil
}

func (e ellipticECDH) ComputeSecret(privkey crypto.PrivateKey, peerPubkey []byte) ([]byte, error) {
	err := e.Check(peerPubkey)
	if err != nil {
		return nil, err
	}
	x, y := getXYFromPub(peerPubkey)
	privBytes, err := privkey.Bytes()
	if err != nil {
		return nil, err
	}
	sX, _ := e.curve.ScalarMult(x, y, privBytes)
	secret := sX.Bytes()

	return secret, nil
}

func getXYFromPub(pub []byte) (X, Y *big.Int) {
	x := big.NewInt(0)
	y := big.NewInt(0)
	x.SetBytes(pub[1:33])
	y.SetBytes(pub[33:])
	return x, y
}
