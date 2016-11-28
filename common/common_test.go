package common

import (
	"net"
	"os"
	"testing"
)

const (
	confFile = "../bin/quantum-test.yml"
)

var (
	testPacket []byte
)

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

func testEq(a, b []byte) bool {

	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func TestIPtoInt(t *testing.T) {
	var expected uint32
	actual := IPtoInt(net.ParseIP("0.0.0.0"))
	if expected != actual {
		t.Fatalf("IPtoInt did not return the right value, got: %d, expected: %d", actual, expected)
	}
}

func TestIncrementIP(t *testing.T) {
	expected := net.ParseIP("10.0.0.1")

	actual := net.ParseIP("10.0.0.0")
	IncrementIP(actual)

	if !testEq(expected, actual) {
		t.Fatalf("IncrementIP did not return the right value, got: %s, expected: %s", actual, expected)
	}
}

func TestNewConfig(t *testing.T) {
	os.Setenv("QUANTUM_INTERFACE_NAME", "different")
	os.Setenv("QUANTUM_LISTEN_PORT", "1")
	os.Setenv("QUANTUM_CONF_FILE", confFile)
	os.Setenv("QUANTUM_PID_FILE", "../quantum.pid")

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig returned an error, %s", err)
	}
	t.Log(cfg)
	if cfg == nil {
		t.Fatal("NewConfig returned a blank config")
	}
	if cfg.InterfaceName != "different" {
		t.Fatalf("NewConfig didn't pick up the environment variable replacement for InterfaceName")
	}
	if cfg.ListenPort != 1 {
		t.Fatalf("NewConfig didn't pick up the environment variable replacement for ListenPort")
	}
	if cfg.Datastore != "consul" {
		t.Fatalf("NewConfig didn't pick up the config file replacement for datastore")
	}
}

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

func TestNewMapping(t *testing.T) {
	privateIP := net.ParseIP("0.0.0.0")
	publicip := net.ParseIP("1.1.1.1")
	publicip6 := net.ParseIP("dead::beef")
	publicport := 80
	publicKey := make([]byte, 32)

	actual := NewMapping(privateIP, publicip, publicip6, publicport, publicKey)
	if !testEq(actual.IPv4, publicip) || !testEq(actual.IPv6, publicip6) || actual.Port != publicport || !testEq(actual.PrivateIP, privateIP) || !testEq(actual.PublicKey, publicKey) {
		t.Fatalf("NewMapping did not return the right value, got: %v", actual)
	}
}

func TestParseMapping(t *testing.T) {
	privateIP := net.ParseIP("0.0.0.0")
	publicip := net.ParseIP("1.1.1.1")
	publicip6 := net.ParseIP("dead::beef")
	publicport := 80
	publicKey := make([]byte, 32)

	expected := NewMapping(privateIP, publicip, publicip6, publicport, publicKey)
	actual, err := ParseMapping(expected.Bytes(), make([]byte, 32))
	if err != nil {
		t.Fatalf("Error occured during test: %s", err)
	}
	if !testEq(actual.IPv4, expected.IPv4) || actual.Port != expected.Port || !testEq(actual.PrivateIP, expected.PrivateIP) || !testEq(actual.PublicKey, expected.PublicKey) {
		t.Fatalf("ParseMapping did not return the right value, got: %v, expected: %v", actual, expected)
	}
}

func TestParseNetworkConfig(t *testing.T) {
	actual, err := ParseNetworkConfig(DefaultNetworkConfig.Bytes())
	if err != nil {
		t.Fatal("ParseNetworkConfig returned an error:", err)
	}
	if actual.Network != DefaultNetworkConfig.Network || actual.LeaseTime != DefaultNetworkConfig.LeaseTime {
		t.Fatalf("ParseNetworkConfig returned the wrong value, got: %v, expected: %v", actual, DefaultNetworkConfig)
	}
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

func TestNewStats(t *testing.T) {
	stats := NewStats()
	if stats.Packets != 0 {
		t.Fatalf("NewStats did not return the correct default for Packets, got: %d, expected: %d", stats.Packets, 0)
	}
	if stats.Bytes != 0 {
		t.Fatalf("NewStats did not return the correct default for Bytes, got: %d, expected: %d", stats.Bytes, 0)
	}
	if stats.Links == nil {
		t.Fatalf("NewStats did not return the correct default for Links, got: %v, expected: %v", stats.Links, make(map[string]*Stats))
	}
	str := stats.String()
	if str == "" {
		t.Fatalf("String didn't return the correct value.")
	}
}

func TestNewLogger(t *testing.T) {
	log := NewLogger()
	if log.Error == nil {
		t.Fatal("NewLogger returned a nil Error log.")
	}
	if log.Info == nil {
		t.Fatal("NewLogger returned a nil Error log.")
	}
}
