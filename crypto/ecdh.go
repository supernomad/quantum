// Copyright (c) 2016-2018 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package crypto

import (
	"crypto/rand"

	"golang.org/x/crypto/curve25519"
)

// GenerateECKeyPair - Generates a new eliptical curve key-pair using curve25519 as the underlying cryptographic function.
func GenerateECKeyPair() ([]byte, []byte) {
	var pub, priv [keyLength]byte

	rand.Read(priv[:])
	curve25519.ScalarBaseMult(&pub, &priv)

	return pub[:], priv[:]
}

// GenerateSharedSecret - Generates a shared secret based on the supplied public/private curve25519 eliptical curve keys.
func GenerateSharedSecret(pubkey, privkey []byte) []byte {
	var secret, pub, priv [keyLength]byte

	copy(pub[:], pubkey)
	copy(priv[:], privkey)

	curve25519.ScalarMult(&secret, &priv, &pub)
	return secret[:]
}
