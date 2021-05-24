package types

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	hash0         = "0x9f41dd84524bf8a42f8ab58ecfca6e1752d6fd93fe8dc00af4c71963c97db59f"
	formalHash    = "0x9f41DD84524bF8A42F8ab58eCFCA6E1752D6Fd93fE8dc00Af4c71963c97dB59f"
	account       = "0x929545f44692178edb7fa468b44c5351596184ba"
	formalAccount = "0x929545f44692178EDb7FA468B44c5351596184Ba"
)

func TestHash(t *testing.T) {
	hash1 := NewHashByStr(hash0)

	require.Equal(t, formalHash, hash1.String())
	require.Equal(t, formalHash, hash1.String())

	// test address SetBytes and Set
	hash2 := &Hash{}
	hash2.SetBytes(hash1.Bytes())
	require.Equal(t, true, bytes.Equal(hash1.RawHash[:], hash2.Bytes()))
	require.Equal(t, hash1.String(), hash2.String())
	require.Equal(t, hash1.String(), hash2.String())

	// test for address marshal and unmarshal
	encode := make([]byte, HashLength)
	n, err := hash1.MarshalTo(encode)
	require.Nil(t, err)
	require.Equal(t, HashLength, n)

	hash3 := &Hash{}
	err = hash3.Unmarshal(encode)
	require.Nil(t, err)
	require.Equal(t, true, bytes.Equal(hash1.RawHash[:], hash3.Bytes()))
	require.Equal(t, hash1.String(), hash3.String())
	require.Equal(t, hash1.String(), hash3.String())

	// test for address marshalJson and unmarshalJson
	encode2, err := hash1.MarshalJSON()
	require.Nil(t, err)
	require.NotNil(t, encode2)

	hash4 := &Hash{}
	err = hash4.UnmarshalJSON(encode2)
	require.Nil(t, err)
	require.Equal(t, true, bytes.Equal(hash1.RawHash[:], hash4.Bytes()))
	require.Equal(t, hash1.String(), hash4.String())
	require.Equal(t, hash1.String(), hash4.String())

	// test for formal address
	hash5 := NewHashByStr(formalHash)
	require.Equal(t, formalHash, hash5.String())
	require.Equal(t, hash1.String(), hash5.String())

	hash6 := &Hash{}
	hash6.SetString(formalHash)
	require.Equal(t, formalHash, hash6.String())
	require.Equal(t, hash1.String(), hash6.String())
}

func TestAddress(t *testing.T) {
	addr1 := NewAddressByStr(account)

	require.Equal(t, formalAccount, addr1.String())
	require.Equal(t, formalAccount, addr1.String())

	// test address SetBytes and Set
	addr2 := &Address{}
	addr2.SetBytes(addr1.Bytes())
	require.Equal(t, true, bytes.Equal(addr1.RawAddress[:], addr2.Bytes()))
	require.Equal(t, addr1.String(), addr2.String())
	require.Equal(t, addr1.String(), addr2.String())

	// test for address marshal and unmarshal
	encode := make([]byte, AddressLength)
	n, err := addr1.MarshalTo(encode)
	require.Nil(t, err)
	require.Equal(t, AddressLength, n)

	addr3 := &Address{}
	err = addr3.Unmarshal(encode)
	require.Nil(t, err)
	require.Equal(t, true, bytes.Equal(addr1.RawAddress[:], addr3.Bytes()))
	require.Equal(t, addr1.String(), addr3.String())
	require.Equal(t, addr1.String(), addr3.String())

	// test for address marshalJson and unmarshalJson
	encode2, err := addr1.MarshalJSON()
	require.Nil(t, err)
	require.NotNil(t, encode2)

	addr4 := &Address{}
	err = addr4.UnmarshalJSON(encode2)
	require.Nil(t, err)
	require.Equal(t, true, bytes.Equal(addr1.RawAddress[:], addr4.Bytes()))
	require.Equal(t, addr1.String(), addr4.String())
	require.Equal(t, addr1.String(), addr4.String())

	// test for formal address
	addr5 := NewAddressByStr(formalAccount)
	require.Equal(t, formalAccount, addr5.String())
	require.Equal(t, addr1.String(), addr5.String())

	addr6 := &Address{}
	addr6.SetString(formalAccount)
	require.Equal(t, formalAccount, addr6.String())
	require.Equal(t, addr1.String(), addr6.String())
}

func TestIsValidAddressByte(t *testing.T) {
	require.Equal(t, false, IsValidAddressByte([]byte(hash0)))
	require.Equal(t, true, IsValidAddressByte([]byte(account)))
	require.Equal(t, true, IsValidAddressByte([]byte(formalAccount)))
}

func TestBloom_MarshalTo(t *testing.T) {
	bloom0 := Bloom{0, 1, 1}

	data, err := bloom0.MarshalJSON()
	require.Nil(t, err)

	var bloom1 Bloom

	err = bloom1.UnmarshalJSON(data)
	require.Nil(t, err)

	require.Equal(t, bloom0, bloom1)
}
