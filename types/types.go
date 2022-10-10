package types

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"math/big"
	"sync"

	mt "github.com/cbergoon/merkletree"
	"github.com/golang/protobuf/jsonpb"
	"github.com/meshplus/bitxhub-kit/hexutil"
	"golang.org/x/crypto/sha3"
)

// hasherPool holds LegacyKeccak256 hashers for rlpHash.
var (
	hasherPool = sync.Pool{
		New: func() interface{} { return sha3.NewLegacyKeccak256() },
	}
	ErrSyntax        = fmt.Errorf("invalid hex string")
	ErrMissingPrefix = fmt.Errorf("hex string without 0x prefix")
	ErrOddLength     = fmt.Errorf("hex string of odd length")
)

type KeccakState interface {
	hash.Hash
	Read([]byte) (int, error)
}

// Lengths of hashes and addresses in bytes.
const (
	HashLength      = 32
	AddressLength   = 20
	BloomByteLength = 256
	badNibble       = ^uint64(0)
)

type Hash struct {
	RawHash [HashLength]byte `json:"raw_hash"`
	Hash    string           `json:"hash"`
}

type Address struct {
	RawAddress [AddressLength]byte `json:"raw_address"`
	Address    string              `json:"address"`
}

type Bloom [BloomByteLength]byte

func (h *Hash) MarshalJSONPB(m *jsonpb.Marshaler) ([]byte, error) {
	return h.MarshalJSON()
}

func (h *Hash) UnmarshalJSONPB(m *jsonpb.Unmarshaler, data []byte) error {
	return h.UnmarshalJSON(data)
}

func (a *Address) MarshalJSONPB(m *jsonpb.Marshaler) ([]byte, error) {
	return a.MarshalJSON()
}

func (a *Address) UnmarshalJSONPB(m *jsonpb.Unmarshaler, data []byte) error {
	return a.UnmarshalJSON(data)
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
	a.Hash = a.String()
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

func (b *Bloom) ProtoMessage() {}

func (b *Bloom) Size() int {
	return BloomByteLength
}

func (b *Bloom) MarshalTo(data []byte) (int, error) {
	data = data[:b.Size()]
	copy(data, b[:])

	return b.Size(), nil
}

func (b *Bloom) Unmarshal(data []byte) error {
	if len(data) > BloomByteLength {
		data = data[len(data)-BloomByteLength:]
	}

	copy(b[:], data)

	return nil
}

// Add adds d to the filter. Future calls of Test(d) will return true.
func (b *Bloom) Add(d []byte) {
	b.add(d, make([]byte, 6))
}

// add is internal version of Add, which takes a scratch buffer for reuse (needs to be at least 6 bytes)
func (b *Bloom) add(d []byte, buf []byte) {
	i1, v1, i2, v2, i3, v3 := bloomValues(d, buf)
	b[i1] |= v1
	b[i2] |= v2
	b[i3] |= v3
}

// MarshalText encodes b as a hex string with 0x prefix.
func (b Bloom) MarshalText() ([]byte, error) {
	return hexutil.Bytes(b[:]).MarshalText()
}

// UnmarshalText b as a hex string with 0x prefix.
func (b *Bloom) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("Bloom", input, b[:])
}

// Test checks if the given topic is present in the bloom filter
func (b Bloom) Test(topic []byte) bool {
	i1, v1, i2, v2, i3, v3 := bloomValues(topic, make([]byte, 6))
	return v1 == v1&b[i1] &&
		v2 == v2&b[i2] &&
		v3 == v3&b[i3]
}

// bloomValues returns the bytes (index-value pairs) to set for the given data
func bloomValues(data []byte, hashbuf []byte) (uint, byte, uint, byte, uint, byte) {
	sha := hasherPool.Get().(KeccakState)
	sha.Reset()
	sha.Write(data)
	sha.Read(hashbuf)
	hasherPool.Put(sha)
	// The actual bits to flip
	v1 := byte(1 << (hashbuf[1] & 0x7))
	v2 := byte(1 << (hashbuf[3] & 0x7))
	v3 := byte(1 << (hashbuf[5] & 0x7))
	// The indices for the bytes to OR in
	i1 := BloomByteLength - uint((binary.BigEndian.Uint16(hashbuf)&0x7ff)>>3) - 1
	i2 := BloomByteLength - uint((binary.BigEndian.Uint16(hashbuf[2:])&0x7ff)>>3) - 1
	i3 := BloomByteLength - uint((binary.BigEndian.Uint16(hashbuf[4:])&0x7ff)>>3) - 1

	return i1, v1, i2, v2, i3, v3
}

// OrBloom executes an Or operation on the bloom
func (b *Bloom) OrBloom(bl *Bloom) {
	bin := new(big.Int).SetBytes(b[:])
	input := new(big.Int).SetBytes(bl[:])
	bin.Or(bin, input)
	b.SetBytes(bin.Bytes())
}

// SetBytes sets the content of b to the given bytes.
// It panics if d is not of suitable size.
func (b *Bloom) SetBytes(d []byte) {
	if len(b) < len(d) {
		panic(fmt.Sprintf("bloom bytes too big %d %d", len(b), len(d)))
	}
	copy(b[BloomByteLength-len(d):], d)
}
