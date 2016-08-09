package common

import (
	"testing"
)

func TestNewMapping(t *testing.T) {
	privateIP := "0.0.0.0"
	machineID := "dad2ac7b7a62a1420fd1b873495073afc6850fc1b024ac4656ce5d61a3aae1fa"
	publicaddress := "1.1.1.1:80"
	publicKey := make([]byte, 32)

	actual := NewMapping(privateIP, publicaddress, machineID, publicKey)
	if actual.Address != publicaddress || actual.MachineID != machineID || actual.PrivateIP != privateIP || len(actual.PublicKey) != 32 {
		t.Fatalf("NewMapping did not return the right value, got: %v", actual)
	}
}

func TestParseMapping(t *testing.T) {
	privateIP := "0.0.0.0"
	machineID := "dad2ac7b7a62a1420fd1b873495073afc6850fc1b024ac4656ce5d61a3aae1fa"
	publicaddress := "1.1.1.1:80"
	publicKey := make([]byte, 32)

	expected := NewMapping(privateIP, publicaddress, machineID, publicKey)
	actual, err := ParseMapping(expected.Bytes(), make([]byte, 32))
	if err != nil {
		t.Fatalf("Error occured during test: %s", err)
	}
	if actual.Address != expected.Address || actual.MachineID != expected.MachineID || actual.PrivateIP != expected.PrivateIP || len(actual.PublicKey) != len(expected.PublicKey) {
		t.Fatalf("ParseMapping did not return the right value, got: %v, expected: %v", actual, expected)
	}
}
