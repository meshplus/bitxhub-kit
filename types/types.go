package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/meshplus/bitxhub-kit/hexutil"
)

// Lengths of hashes and addresses in bytes.
const (
	HashLength    = 32
	AddressLength = 20
)

type Hash [HashLength]byte
type Address [AddressLength]byte

func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-HashLength:]
	}

	copy(h[HashLength-len(b):], b)
}

func (h Hash) Bytes() []byte {
	return h[:]
}

func (h Hash) Hex() string {
	return "0x" + hex.EncodeToString(h[:])
}

func (h *Hash) MarshalTo(data []byte) (int, error) {
	data = data[:h.Size()]
	copy(data, h[:])

	return h.Size(), nil
}

func (h Hash) String() string {
	return h.Hex()
}

func (h Hash) ShortString() string {
	s := hex.EncodeToString(h[:])
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

	copy(h[:], ret)

	return nil
}

// SetBytes sets the address to the value of b.
// If b is larger than len(a) it will panic.
func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AddressLength:]
	}
	copy(a[AddressLength-len(b):], b)
}

func (a Address) Bytes() []byte {
	return a[:]
}

// Hex returns an EIP55-compliant hex string representation of the address.
func (a Address) Hex() string { return "0x" + hex.EncodeToString(a[:]) }

func (a Address) String() string {
	return a.Hex()
}

func (a Address) ShortString() string {
	s := hex.EncodeToString(a[:])
	return s[:6] + "..." + s[len(s)-6:]
}

func (a Address) Size() int {
	return AddressLength
}

func (a *Address) MarshalTo(data []byte) (int, error) {
	data = data[:a.Size()]
	copy(data, a[:])

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

	n, err := hex.Decode(a[:], data)
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
func (a *Address) Set(other Address) {
	for i, v := range other {
		a[i] = v
	}
}

// BytesToAddress returns Address with value b.
// If b is larger than len(h), b will be cropped from the left.
func Bytes2Address(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}

func String2Address(s string) Address {
	d := hexutil.Decode(s)
	return Bytes2Address(d)
}

func Bytes2Hash(b []byte) Hash {
	var a Hash
	a.SetBytes(b)
	return a
}

func String2Hash(s string) Hash {
	d := hexutil.Decode(s)
	return Bytes2Hash(d)
}
