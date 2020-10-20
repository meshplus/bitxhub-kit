package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	mt "github.com/cbergoon/merkletree"
	"github.com/meshplus/bitxhub-kit/hexutil"
	"golang.org/x/crypto/sha3"
)

// Lengths of hashes and addresses in bytes.
const (
	HashLength    = 32
	AddressLength = 20
)

type Hash struct {
	rawHash [HashLength]byte
	hash    string
}

type Address struct {
	rawAddress [AddressLength]byte
	address    string
}

// CalculateHash hashes the values of a TestContent
func (h *Hash) CalculateHash() ([]byte, error) {
	return h.rawHash[:], nil
}

// Equals tests for equality of two Contents
func (h *Hash) Equals(other mt.Content) (bool, error) {
	tOther, ok := other.(*Hash)
	if !ok {
		return false, errors.New("parameter should be type TransactionHash")
	}
	return bytes.Equal(h.rawHash[:], tOther.rawHash[:]), nil
}

func (h *Hash) SetBytes(b []byte) {
	if len(b) > HashLength {
		b = b[len(b)-HashLength:]
	}

	copy(h.rawHash[HashLength-len(b):], b)
	h.hash = ""
}

func (h Hash) Bytes() []byte {
	return h.rawHash[:]
}

func (h *Hash) Hex() string {
	if h.hash == "" {
		// if hash field is empty, initialize it for only once
		h.hash = "0x" + string(toCheckSum(h.rawHash[:]))
	}
	return h.hash
}

func (h *Hash) MarshalTo(data []byte) (int, error) {
	data = data[:h.Size()]
	copy(data, h.rawHash[:])

	return h.Size(), nil
}

func (h *Hash) String() string {
	return h.Hex()
}

func (h Hash) ShortString() string {
	s := hex.EncodeToString(h.rawHash[:])
	return s[:6] + "..." + s[len(s)-6:]
}

func (h Hash) Size() int {
	return HashLength
}

func (h *Hash) Unmarshal(data []byte) error {
	h.SetBytes(data)

	return nil
}

// Serialize given address to JSON
func (h Hash) MarshalJSON() ([]byte, error) {
	rs := []byte(fmt.Sprintf(`"%s"`, h.Hex()))

	return rs, nil
}

func (h *Hash) UnmarshalJSON(data []byte) error {
	if len(data) > 2 && data[0] == '"' && data[len(data)-1] == '"' {
		data = data[1 : len(data)-1]
	}

	if len(data) > 2 && data[0] == '0' && data[1] == 'x' {
		data = data[2:]
	}

	if len(data) != 2*HashLength {
		return fmt.Errorf("invalid hash length, expected %d got %d bytes", 2*HashLength, len(data))
	}

	ret, err := hex.DecodeString(string(data))
	if err != nil {
		return err
	}

	copy(h.rawHash[:], ret)

	return nil
}

// SetBytes sets the address to the value of b.
// If b is larger than len(a) it will panic.
func (a *Address) SetBytes(b []byte) {
	if len(b) > AddressLength {
		b = b[len(b)-AddressLength:]
	}
	copy(a.rawAddress[AddressLength-len(b):], b)
	a.address = ""
}

func (a Address) Bytes() []byte {
	return a.rawAddress[:]
}

// Hex returns an EIP55-compliant hex string representation of the address.
func (a *Address) Hex() string {
	if a.address == "" {
		// if address field is empty, initialize it for only once
		a.address = "0x" + string(toCheckSum(a.rawAddress[:]))
	}
	return a.address
}

func (a *Address) String() string {
	return a.Hex()
}

func (a Address) ShortString() string {
	s := hex.EncodeToString(a.rawAddress[:])
	return s[:6] + "..." + s[len(s)-6:]
}

func (a Address) Size() int {
	return AddressLength
}

func (a *Address) MarshalTo(data []byte) (int, error) {
	data = data[:a.Size()]
	copy(data, a.rawAddress[:])

	return a.Size(), nil
}

func (a *Address) Unmarshal(data []byte) error {
	a.SetBytes(data)

	return nil
}

// Serialize given address to JSON
func (a Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Hex())
}

// UnmarshalJSON parses a hash in hex syntax.
func (a *Address) UnmarshalJSON(data []byte) error {
	if len(data) > 2 && data[0] == '"' && data[len(data)-1] == '"' {
		data = data[1 : len(data)-1]
	}

	if len(data) > 2 && data[0] == '0' && data[1] == 'x' {
		data = data[2:]
	}

	if len(data) != 2*AddressLength {
		return fmt.Errorf("invalid address length, expected %d got %d bytes", 2*AddressLength, len(data))
	}

	n, err := hex.Decode(a.rawAddress[:], data)
	if err != nil {
		return err
	}

	if n != AddressLength {
		return fmt.Errorf("invalid address")
	}

	bytes, err := hex.DecodeString(string(data))
	if err != nil {
		return err
	}

	a.Set(Bytes2Address(bytes))

	return nil
}

// Sets a to other
func (a *Address) Set(other *Address) {
	for i, v := range other.rawAddress {
		a.rawAddress[i] = v
	}
	a.address = ""
}

// BytesToAddress returns Address with value b.
// If b is larger than len(h), b will be cropped address the left.
func Bytes2Address(b []byte) *Address {
	a := &Address{}
	a.SetBytes(b)
	return a
}

func String2Address(s string) *Address {
	d := hexutil.Decode(s)
	return Bytes2Address(d)
}

func Bytes2Hash(b []byte) *Hash {
	a := &Hash{}
	a.SetBytes(b)
	return a
}

func String2Hash(s string) *Hash {
	d := hexutil.Decode(s)
	return Bytes2Hash(d)
}

func toCheckSum(a []byte) []byte {
	unchecksummed := hex.EncodeToString(a[:])
	sha := sha3.NewLegacyKeccak256()
	sha.Write([]byte(unchecksummed))
	hash := sha.Sum(nil)

	result := []byte(unchecksummed)
	for i := 0; i < len(result); i++ {
		hashByte := hash[i/2]
		if i%2 == 0 {
			hashByte = hashByte >> 4
		} else {
			hashByte &= 0xf
		}
		if result[i] > '9' && hashByte > 7 {
			result[i] -= 32
		}
	}
	return result
}

func IsValidAddressByte(data []byte) bool {
	if len(data) > 2 && data[0] == '"' && data[len(data)-1] == '"' {
		data = data[1 : len(data)-1]
	}

	if len(data) > 2 && data[0] == '0' && data[1] == 'x' {
		data = data[2:]
	}

	// address data length check
	if len(data) != 2*AddressLength {
		return false
	}

	var a [AddressLength]byte

	// address hex format check
	n, err := hex.Decode(a[:], data)
	if err != nil {
		return false
	}

	// address decoded data length check
	if n != AddressLength {
		return false
	}
	return true
}
