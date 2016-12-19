// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE
package common

import (
	"crypto/rand"

	"golang.org/x/crypto/curve25519"
)

const (
	keyLength = 32
)

// GenerateECKeyPair - Generates a new eliptical curve key-pair
func GenerateECKeyPair() ([]byte, []byte) {
	var pub, priv [keyLength]byte

	rand.Read(priv[:])
	curve25519.ScalarBaseMult(&pub, &priv)

	return pub[:], priv[:]
}

// GenerateSharedSecret - Generates a shared secret based on the supplied public/private keys
func GenerateSharedSecret(pubkey, privkey []byte) []byte {
	var secret, pub, priv [keyLength]byte

	copy(pub[:], pubkey)
	copy(priv[:], privkey)

	curve25519.ScalarMult(&secret, &priv, &pub)
	return secret[:]
}
