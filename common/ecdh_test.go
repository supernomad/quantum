package common

import (
	"testing"
)

func TestGenerateECKeyPair(t *testing.T) {
	pub, priv := GenerateECKeyPair()
	if len(pub) != keyLength {
		t.Fatalf("GenerateECKeyPair did not return the right length for the public key, got: %d, expected: %d", len(pub), keyLength)
	}
	if len(priv) != keyLength {
		t.Fatalf("GenerateECKeyPair did not return the right length for the private key, got: %d, expected: %d", len(priv), keyLength)
	}
	if testEq(pub, priv) {
		t.Fatalf("GenerateECKeyPair returned identical pub/priv keys this can't possibly happen: got: %v, and %v", pub, priv)
	}
}
