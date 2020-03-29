package secp256k1

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSign(t *testing.T) {
	key := "3e83c9cd9a39bf96d1f77a978e1fb32be0ad1732eee157011e162e9749b2e90a"
	pub := "0454f73fc55299a40b20638f10c2d704dcb605e7e61b56ea11b4ad9528e533ab8501089fc3a87512ad4f6a6631086ab2734c56811268a1b7816d684d709c6becb8"
	keyBytes, _ := hex.DecodeString(key)
	pubBytes, _ := hex.DecodeString(pub)
	h := make([]byte, 32)
	_, _ = rand.Read(h)
	s, e := Sign(h, keyBytes)
	if e != nil {
		t.Error(e)
	}
	tmp, e := RecoverPubkey(h, s)
	if e != nil {
		t.Error(e)
	}
	for i := range tmp {
		if tmp[i] != pubBytes[i] {
			t.Error()
		}
	}

}

func TestSm2P256Curve_Params(t *testing.T) {
	para := S256().Params()
	assert.NotNil(t, para)
	target, _ := new(big.Int).SetString("79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798", 16)
	assert.True(t, target.Cmp(para.Gx) == 0)
}

func TestSm2P256Curve_Add(t *testing.T) {
	para := S256().Params()
	e := S256()
	a, b := e.ScalarBaseMult(big.NewInt(9).Bytes())
	c, d := e.Double(e.Double(e.Double(para.Gx, para.Gy)))
	c, d = e.Add(c, d, para.Gx, para.Gy)
	assert.True(t, a.Cmp(c) == 0)
	assert.True(t, b.Cmp(d) == 0)
}

func TestSm2P256Curve_IsOnCurve(t *testing.T) {
	e := S256()
	a, b := e.ScalarBaseMult(big.NewInt(5201314).Bytes())
	assert.True(t, e.IsOnCurve(a, b))
}

func TestSm2P256Curve_ScalarMult(t *testing.T) {
	e := S256()
	para := S256().Params()
	a, b := e.ScalarBaseMult(big.NewInt(5201314).Bytes())
	c, d := e.ScalarMult(para.Gx, para.Gy, big.NewInt(5201314).Bytes())
	assert.True(t, e.IsOnCurve(a, b))
	assert.True(t, a.Cmp(c) == 0)
	assert.True(t, b.Cmp(d) == 0)
}
