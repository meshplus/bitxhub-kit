package gm

import (
	"fmt"

	"github.com/meshplus/bitxhub-kit/crypto"
	gm "github.com/ultramesh/crypto-gm"
)

func New(opt crypto.KeyType) (crypto.PrivateKey, error) {
	switch opt {
	case crypto.SM2:
		pri, err := gm.GenerateSM2Key()
		if err != nil {
			return nil, err
		}

		return &SM2PrivateKey{K: pri, curve: crypto.SM2}, nil
	}
	return nil, fmt.Errorf("wrong curve option")
}
