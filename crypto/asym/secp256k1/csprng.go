package secp256k1

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"io"
)

const aesIV = "HYPERCHAINAES IV"

func csprng(entropylen int) ([]byte, error) {

	entropy := make([]byte, entropylen)
	_, err := io.ReadFull(rand.Reader, entropy)
	if err != nil {
		return nil, err
	}
	// Initialize an SHA-512 hash context; digest ...
	md := sha512.New()
	_, _ = md.Write(entropy) // the entropy,
	key := md.Sum(nil)[:32]  // and compute ChopMD-256(SHA-512),

	// Create an AES-CTR instance to use as a CSPRNG.
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create a CSPRNG that xors a stream of zeros with
	// the output of the AES-CTR instance.
	csprng := cipher.StreamReader{
		R: zeroReader,
		S: cipher.NewCTR(block, []byte(aesIV)),
	}

	randbuf := make([]byte, entropylen)
	_, err = io.ReadFull(csprng.R, randbuf)
	if err != nil {
		return nil, err
	}
	return randbuf[:], nil

}

type zr struct {
	io.Reader
}

// Read replaces the contents of dst with zeros.
func (z *zr) Read(dst []byte) (n int, err error) {
	for i := range dst {
		dst[i] = 0
	}
	return len(dst), nil
}

var zeroReader = &zr{}
