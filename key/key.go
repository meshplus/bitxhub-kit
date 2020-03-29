package key

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/meshplus/bitxhub-kit/crypto"
	"github.com/meshplus/bitxhub-kit/crypto/asym/ecdsa"
	"github.com/meshplus/bitxhub-kit/crypto/sym"
	"github.com/meshplus/bitxhub-kit/types"
)

const padding = "abcdefghijklmnopqrstuvwxyz"

// Key represents a private key
type Key struct {
	Address    types.Address `json:"address"`
	PrivateKey string        `json:"private_key"`
	Encrypted  bool          `json:"encrypted"`
}

// New create key using ecdsa (secp256r1)
// if password is empty, encrypted is false
func New(password string) (*Key, error) {
	key, err := ecdsa.GenerateKey(ecdsa.Secp256r1)
	if err != nil {
		return nil, fmt.Errorf("generate key: %s", err)
	}

	return NewWithPrivateKey(key, password)
}

func NewWithPrivateKey(privateKey crypto.PrivateKey, password string) (*Key, error) {
	bytes, err := privateKey.Bytes()
	if err != nil {
		return nil, fmt.Errorf("marshal key: %s", err)
	}

	var cipher string
	var encrypted bool

	if password != "" {
		des, err := sym.GenerateKey(sym.ThirdDES, []byte(password+padding))
		if err != nil {
			return nil, err
		}

		data, err := des.Encrypt(bytes)
		if err != nil {
			return nil, fmt.Errorf("encrypt private key: %s", err)
		}

		cipher = hex.EncodeToString(data)
		encrypted = true
	} else {
		cipher = hex.EncodeToString(bytes)
	}

	address, err := privateKey.PublicKey().Address()
	if err != nil {
		return nil, err
	}

	return &Key{
		Address:    address,
		PrivateKey: cipher,
		Encrypted:  encrypted,
	}, nil
}

func LoadKey(path string) (*Key, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	key := &Key{}
	if err := key.Unmarshal(data); err != nil {
		return nil, err
	}

	return key, nil
}

func (key *Key) GetPrivateKey(password string) (crypto.PrivateKey, error) {
	bytes, err := hex.DecodeString(key.PrivateKey)
	if err != nil {
		return nil, err
	}

	if key.Encrypted {
		des, err := sym.GenerateKey(sym.ThirdDES, []byte(password+padding))
		if err != nil {
			return nil, err
		}

		bytes, err = des.Decrypt(bytes)
		if err != nil {
			return nil, err
		}
	}

	return ecdsa.UnmarshalPrivateKey(bytes, ecdsa.Secp256r1)
}

func (key *Key) Pretty() (string, error) {
	ret, err := json.MarshalIndent(key, "", "	")
	if err != nil {
		return "", err
	}

	return string(ret), nil
}

func (key *Key) Marshal() ([]byte, error) {
	return json.Marshal(key)
}

func (key *Key) Unmarshal(data []byte) error {
	return json.Unmarshal(data, key)
}
