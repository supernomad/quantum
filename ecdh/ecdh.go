package ecdh

import (
	"crypto/rand"
	"golang.org/x/crypto/curve25519"
)

const (
	keyLength = 32
)

func GenerateECKeyPair() ([]byte, []byte, error) {
	var pub, priv [keyLength]byte

	_, err := rand.Read(priv[:])
	if err != nil {
		return nil, nil, err
	}

	curve25519.ScalarBaseMult(&pub, &priv)

	return pub[:], priv[:], nil
}

func GenerateSharedSecret(pubkey, privkey []byte) []byte {
	var secret, pub, priv [keyLength]byte

	copy(pub[:], pubkey)
	copy(priv[:], privkey)

	curve25519.ScalarMult(&secret, &priv, &pub)
	return secret[:]
}
