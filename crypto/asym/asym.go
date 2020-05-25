package asym

import (
	crypto2 "crypto"
	"fmt"

	"github.com/meshplus/bitxhub-kit/crypto"
	"github.com/meshplus/bitxhub-kit/crypto/asym/ecdsa"
	"github.com/meshplus/bitxhub-kit/types"
)

// AlgorithmOption represent ECDSA curve

type AlgorithmOption int

const (
	// RSA is an enum for the supported RSA key type
	RSA AlgorithmOption = iota
	// ECDSASecp256r1 is an enum for the supported ECDSA key type
	ECDSASecp256r1
	// Ed25519 is an enum for the supported Ed25519 key type
	Ed25519
)

func GenerateKey(opt AlgorithmOption) (crypto.PrivateKey, error) {
	switch opt {
	case RSA:
		return nil, fmt.Errorf("don`t support rsa algorithm currently")
	case ECDSASecp256r1:
		return ecdsa.GenerateKey(ecdsa.Secp256r1)
	case Ed25519:
		return nil, fmt.Errorf("don`t support ed25519 algorithm currently")
	default:
		return nil, fmt.Errorf("wront algorithm type")
	}
}

func Verify(opt AlgorithmOption, sig, digest []byte, from types.Address) (bool, error) {
	switch opt {
	case RSA:
		return false, fmt.Errorf("don`t support rsa algorithm currently")
	case ECDSASecp256r1:
		if len(sig) != 130 {
			return false, fmt.Errorf("signature length is not correct")
		}
		pubBytes := sig[65:]
		pubkey, err := ecdsa.UnmarshalPublicKey(pubBytes, ecdsa.Secp256r1)
		if err != nil {
			return false, err
		}
		return pubkey.Verify(digest, sig)
	case Ed25519:
		return false, fmt.Errorf("don`t support ed25519 algorithm currently")
	default:
		return false, fmt.Errorf("wront algorithm type")
	}
}

func Convert2StdPrivateKey(privKey crypto.PrivateKey) crypto2.PrivateKey {
	switch p := privKey.(type) {
	case *ecdsa.PrivateKey:
		return p.K
	default:
		panic("don't support this algorithm")
	}
}
