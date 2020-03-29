package key

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	key, err := New("key")
	require.Nil(t, err)

	data, err := key.Marshal()
	require.Nil(t, err)

	k := &Key{}
	err = k.Unmarshal(data)
	require.Nil(t, err)

	privateKey, err := key.GetPrivateKey("key")
	require.Nil(t, err)

	address, err := privateKey.PublicKey().Address()
	require.Nil(t, err)
	require.EqualValues(t, address, key.Address)
}

func TestLoadKey(t *testing.T) {
	key, err := LoadKey("./testdata/key.json")
	require.Nil(t, err)
	require.EqualValues(t, "0x909b162095adc0cce4df54424965d57e30f58d10", key.Address.Hex())
}
