package common

import (
	"testing"
)

func TestEcdh(t *testing.T) {
	pub, priv := GenerateECKeyPair()
	if len(pub) != keyLength {
		t.Fatalf("GenerateECKeyPair did not return the right length for the public key,\nactual: %d, expected: %d", len(pub), keyLength)
	}
	if len(priv) != keyLength {
		t.Fatalf("GenerateECKeyPair did not return the right length for the private key,\nactual: %d, expected: %d", len(priv), keyLength)
	}
	if testEq(pub, priv) {
		t.Fatalf("GenerateECKeyPair returned identical pub/priv keys this can't possibly happen:\npub: %v, priv: %v", pub, priv)
	}
	secret := GenerateSharedSecret(pub, priv)
	if len(secret) != keyLength {
		t.Fatalf("GenerateECKeyPair did not return the right length for the shared secret,\nactual: %d, expected: %d", len(secret), keyLength)
	}
	if testEq(secret, pub) || testEq(secret, priv) {
		t.Fatalf("GenerateECKeyPair returned identical secret and pub/priv keys this can't possibly happen:\npub: %v, priv: %v, secret: %v", pub, priv, secret)
	}
}
