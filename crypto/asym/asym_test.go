package asym

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestECDSAr1(t *testing.T) {
	digest := sha256.Sum256([]byte("hyperchain"))

	priv, err := GenerateKey(ECDSASecp256r1)
	require.Nil(t, err)

	sig, err := priv.Sign(digest[:])
	require.Nil(t, err)

	from, err := priv.PublicKey().Address()
	require.Nil(t, err)

	b, err := Verify(ECDSASecp256r1, sig, digest[:], from)
	require.Nil(t, err)
	require.Equal(t, true, b)
}
