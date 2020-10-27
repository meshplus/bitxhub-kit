package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	mt "github.com/cbergoon/merkletree"
	"golang.org/x/crypto/sha3"
)

// Lengths of hashes and addresses in bytes.
const (
	HashLength    = 32
	AddressLength = 20
)

type Hash struct {
	RawHash [HashLength]byte
	Hash    string
}

type Address struct {
	RawAddress [AddressLength]byte
	Address    string
}

// CalculateHash hashes the values of a TestContent
func (h *Hash) CalculateHash() ([]byte, error) {
	return h.RawHash[:], nil
}

// Equals tests for equality of two Contents
func (h *Hash) Equals(other mt.Content) (bool, error) {
	tOther, ok := other.(*Hash)
	if !ok {
		return false, errors.New("parameter should be type TransactionHash")
	}
	return bytes.Equal(h.RawHash[:], tOther.RawHash[:]), nil
}

func (h *Hash) SetBytes(b []byte) {
	if len(b) > HashLength {
		b = b[len(b)-HashLength:]
	}

	copy(h.RawHash[HashLength-len(b):], b)
	h.Hash = ""
}

func (h *Hash) SetString(s string) {
	h.Hash = s
}

func (h *Hash) Bytes() []byte {
	return h.RawHash[:]
}

func (h *Hash) String() string {
	if h.Hash == "" {
		// if hash field is empty, initialize it for only once
		h.Hash = "0x" + string(toCheckSum(h.RawHash[:]))
	}
	return h.Hash
}

func (h *Hash) Reset() { *h = Hash{} }

func (h *Hash) ProtoMessage() {}

func (h *Hash) MarshalTo(data []byte) (int, error) {
	data = data[:h.Size()]
	copy(data, h.RawHash[:])

	return h.Size(), nil
}

func (h Hash) Size() int {
	return HashLength
}

func (h *Hash) Unmarshal(data []byte) error {
	h.SetBytes(data)

	return nil
}

// Serialize given address to JSON
func (h *Hash) MarshalJSON() ([]byte, error) {
	rs := []byte(fmt.Sprintf(`"%s"`, h.String()))

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

	copy(h.RawHash[:], ret)

	return nil
}

// SetBytes sets the address to the value of b.
// If b is larger than len(a) it will panic.
func (a *Address) SetBytes(b []byte) {
	if len(b) > AddressLength {
		b = b[len(b)-AddressLength:]
	}
	copy(a.RawAddress[AddressLength-len(b):], b)
	a.Address = ""
}

func (a *Address) SetString(s string) {
	a.Address = s
}

func (a *Address) Bytes() []byte {
	return a.RawAddress[:]
}

// String returns an EIP55-compliant hex string representation of the address.
func (a *Address) String() string {
	if a.Address == "" {
		// if address field is empty, initialize it for only once
		a.Address = "0x" + string(toCheckSum(a.RawAddress[:]))
	}
	return a.Address
}

func (a *Address) Reset() { *a = Address{} }

func (a *Address) ProtoMessage() {}

func (a *Address) Size() int {
	return AddressLength
}

func (a *Address) MarshalTo(data []byte) (int, error) {
	data = data[:a.Size()]
	copy(data, a.RawAddress[:])

	return a.Size(), nil
}

func (a *Address) Unmarshal(data []byte) error {
	a.SetBytes(data)

	return nil
}

// Serialize given address to JSON
func (a *Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
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

	n, err := hex.Decode(a.RawAddress[:], data)
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

	a.Set(NewAddress(bytes))

	return nil
}

// Sets a to other
func (a *Address) Set(other *Address) {
	for i, v := range other.RawAddress {
		a.RawAddress[i] = v
	}
	a.Address = ""
}

// BytesToAddress returns Address with value b.
// If b is larger than len(h), b will be cropped address the left.
func NewAddress(b []byte) *Address {
	a := &Address{}
	a.SetBytes(b)
	return a
}

func NewAddressByStr(s string) *Address {
	hashBytes, _ := HexDecodeString(s)
	if len(hashBytes) != AddressLength {
		return nil
	}
	var rawAddr [AddressLength]byte
	copy(rawAddr[:], hashBytes)
	return &Address{RawAddress: rawAddr}
}

func NewHash(b []byte) *Hash {
	a := &Hash{}
	a.SetBytes(b)
	return a
}

func NewHashByStr(s string) *Hash {
	hashBytes, _ := HexDecodeString(s)
	if len(hashBytes) != HashLength {
		return nil
	}
	var rawHash [HashLength]byte
	copy(rawHash[:], hashBytes)
	return &Hash{RawHash: rawHash}
}

// HexDecodeString return rawBytes of a hex hash represent
func HexDecodeString(s string) ([]byte, error) {
	if len(s) > 2 && s[:2] == "0x" {
		s = s[2:]
	}
	return hex.DecodeString(s)
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
