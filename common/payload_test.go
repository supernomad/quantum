package common

import (
	"testing"
)

var testPacket []byte

func init() {
	testPacket = make([]byte, 18)
	// IP (1.1.1.1)
	testPacket[0] = 1
	testPacket[1] = 1
	testPacket[2] = 1
	testPacket[3] = 1

	// Nonce
	testPacket[4] = 2
	testPacket[5] = 2
	testPacket[6] = 2
	testPacket[7] = 2
	testPacket[8] = 2
	testPacket[9] = 2
	testPacket[10] = 2
	testPacket[11] = 2
	testPacket[12] = 2
	testPacket[13] = 2
	testPacket[14] = 2
	testPacket[15] = 2

	// Packet data
	testPacket[16] = 3
	testPacket[17] = 3
}

func TestNewTunPayload(t *testing.T) {
	payload := NewTunPayload(testPacket, 2)
	for i := 0; i < 4; i++ {
		if payload.IPAddress[i] != 1 {
			t.Fatal("NewTunPayload returned an incorrect IP address mapping.")
		}
	}

	for i := 0; i < 12; i++ {
		if payload.Nonce[i] != 2 {
			t.Fatal("NewTunPayload returned an incorrect Nonce mapping.")
		}
	}

	for i := 0; i < 2; i++ {
		if payload.Packet[i] != 3 {
			t.Fatal("NewTunPayload returned an incorrect Packet mapping.")
		}
	}
}

func TestNewSockPayload(t *testing.T) {
	payload := NewSockPayload(testPacket, 18)
	for i := 0; i < 4; i++ {
		if payload.IPAddress[i] != 1 {
			t.Fatal("NewTunPayload returned an incorrect IP address mapping.")
		}
	}

	for i := 0; i < 12; i++ {
		if payload.Nonce[i] != 2 {
			t.Fatal("NewTunPayload returned an incorrect Nonce mapping.")
		}
	}

	for i := 0; i < 2; i++ {
		if payload.Packet[i] != 3 {
			t.Fatal("NewTunPayload returned an incorrect Packet mapping.")
		}
	}
}
