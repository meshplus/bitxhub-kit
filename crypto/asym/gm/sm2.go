package gm

import (
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/meshplus/bitxhub-kit/crypto"
	"github.com/meshplus/bitxhub-kit/crypto/asym/ecdsa"
	"github.com/meshplus/bitxhub-kit/types"
	gm "github.com/ultramesh/crypto-gm"
)

// SM2PrivateKey gm private key.
type SM2PrivateKey struct {
	curve crypto.KeyType
	K     *gm.SM2PrivateKey
}

// PublicKey gm public key.
type PublicKey struct {
	K *gm.SM2PublicKey
}

func (priv *SM2PrivateKey) Bytes() ([]byte, error) {
	return priv.K.Bytes()
}

func (priv *SM2PrivateKey) Type() crypto.KeyType {
	return priv.curve
}

func (priv *SM2PrivateKey) Sign(digest []byte) ([]byte, error) {
	sig, err := priv.K.Sign(nil, digest, rand.Reader)
	if err != nil {
		return nil, err
	}

	pubKeyBytes, err := priv.PublicKey().Bytes()
	if err != nil {
		return nil, err
	}

	return append(pubKeyBytes, sig...), nil
}

func (priv *SM2PrivateKey) PublicKey() crypto.PublicKey {
	pubKey := priv.K.Public().(*gm.SM2PublicKey)
	priv.K.SetPublicKey(pubKey)
	return &PublicKey{K: &priv.K.PublicKey}
}

func (pub *PublicKey) Bytes() ([]byte, error) {
	return pub.K.Bytes()
}

func (pub *PublicKey) Type() crypto.KeyType {
	switch pub.K.Curve {
	case gm.GetSm2Curve():
		return crypto.SM2
	}

	return -1
}

func (pub *PublicKey) Address() (*types.Address, error) {
	x := new(big.Int).SetBytes(pub.K.X[:])
	y := new(big.Int).SetBytes(pub.K.Y[:])
	data := elliptic.Marshal(pub.K.Curve, x, y)
	ret := ecdsa.Keccak256(data[1:])

	return types.NewAddress(ret[12:]), nil
}

func (pub *PublicKey) Verify(digest []byte, sig []byte) (bool, error) {
	return pub.K.Verify(nil, sig, digest)
}

func UnmarshalPublicKey(data []byte, opt crypto.KeyType) (crypto.PublicKey, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty public key data")
	}

	switch opt {
	case crypto.SM2:
		var pubKey gm.SM2PublicKey
		if err := pubKey.FromBytes(data, int(opt)); err != nil {
			return nil, err
		}

		return &PublicKey{K: &pubKey}, nil
	}

	return nil, fmt.Errorf("not supported crypto type %d", opt)
}

func UnmarshalPrivateKey(data []byte, opt crypto.KeyType) (*SM2PrivateKey, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty private key data")
	}

	switch opt {
	case crypto.SM2:
		var privKey gm.SM2PrivateKey
		if err := privKey.FromBytes(data, int(opt)); err != nil {
			return nil, err
		}

		return &SM2PrivateKey{K: &privKey, curve: crypto.SM2}, nil
	}

	return nil, fmt.Errorf("not supported crypto type %d", opt)
}
