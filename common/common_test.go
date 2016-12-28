// Package common testing
// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE
package common

import (
	"net"
	"os"
	"runtime"
	"testing"
	"time"
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

func TestArrayEquals(t *testing.T) {
	if !ArrayEquals(nil, nil) {
		t.Fatal("ArrayEquals returned false comparing nil/nil")
	}
	if ArrayEquals([]byte{0}, nil) {
		t.Fatal("ArrayEquals returned true comparing nil/non-nil")
	}
	if ArrayEquals([]byte{0, 1}, []byte{0}) {
		t.Fatal("ArrayEquals returned true comparing mismatched lengths")
	}
	if !ArrayEquals([]byte{0, 1}, []byte{0, 1}) {
		t.Fatal("ArrayEquals returned false for equal arrays")
	}
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
	os.Setenv("QUANTUM_DEVICE_NAME", "different")
	os.Setenv("QUANTUM_LISTEN_PORT", "1")
	os.Setenv("QUANTUM_CONF_FILE", confFile)
	os.Setenv("QUANTUM_PID_FILE", "../quantum.pid")
	os.Setenv("_QUANTUM_REAL_DEVICE_NAME_", "quantum0")

	os.Args = append(os.Args, "-n", "100", "--prefix", "woot", "--tls-skip-verify", "-6", "fd00:dead:beef::2")
	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig returned an error, %s", err)
	}
	if cfg == nil {
		t.Fatal("NewConfig returned a blank config")
	}
	if cfg.DeviceName != "different" {
		t.Fatalf("NewConfig didn't pick up the environment variable replacement for DeviceName")
	}
	if cfg.ListenPort != 1 {
		t.Fatalf("NewConfig didn't pick up the environment variable replacement for ListenPort")
	}
	if cfg.Password != "Password1" {
		t.Fatalf("NewConfig didn't pick up the config file replacement for Password")
	}
	if cfg.Prefix != "woot" {
		t.Fatal("NewConfig didn't pick up the cli replacement for Prefix")
	}
	if cfg.NumWorkers != runtime.NumCPU() {
		t.Fatal("NewConfig didn't pick up the cli replacement for NumWorkers")
	}
	if !cfg.TLSSkipVerify {
		t.Fatal("NewConfig didn't pick up the cli replacement for TLSSkipVerify")
	}

	cfg.usage(false)
	cfg.version(false)
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
	machineID := "123456"

	actual := NewMapping(machineID, privateIP, publicip, publicip6, publicport, publicKey)
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
	machineID := "123456"

	expected := NewMapping(machineID, privateIP, publicip, publicip6, publicport, publicKey)
	actual, err := ParseMapping(expected.String(), make([]byte, 32))
	if err != nil {
		t.Fatalf("Error occurred during test: %s", err)
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

func TestParseNetworkConfigOnlyNetwork(t *testing.T) {
	netCfg := &NetworkConfig{Network: "10.10.0.0/16"}
	actual, err := ParseNetworkConfig(netCfg.Bytes())
	if err != nil {
		t.Fatal("ParseNetworkConfig returned an error:", err)
	}
	if actual.Network != netCfg.Network || actual.LeaseTime != 48*time.Hour {
		t.Fatalf("ParseNetworkConfig returned the wrong value, got: %v, expected: %v", actual, netCfg)
	}
}

func TestParseNetworkConfigIncorrectFormat(t *testing.T) {
	netCfg := &NetworkConfig{Network: "10.10.0."}
	_, err := ParseNetworkConfig(netCfg.Bytes())
	if err == nil {
		t.Fatal("ParseNetworkConfig should have errored")
	}

	netCfg.Network = "10.10.0.0/16"
	netCfg.StaticRange = "10.10.0./23"

	_, err = ParseNetworkConfig(netCfg.Bytes())
	if err == nil {
		t.Fatal("ParseNetworkConfig should have errored")
	}

	netCfg.StaticRange = "10.20.0.0/23"
	_, err = ParseNetworkConfig(netCfg.Bytes())
	if err == nil {
		t.Fatal("ParseNetworkConfig should have errored")
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
	stats := NewStats(1)
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
	log := NewLogger(false, false, false, false)
	if log.Error == nil {
		t.Fatal("NewLogger returned a nil Error log.")
	}
	if log.Info == nil {
		t.Fatal("NewLogger returned a nil Error log.")
	}
}

func TestGenerateLocalMapping(t *testing.T) {
	cfg := &Config{
		PrivateIP:     net.ParseIP("10.10.0.1"),
		PublicIPv4:    net.ParseIP("192.167.0.1"),
		PublicIPv6:    net.ParseIP("fd00:dead:beef::2"),
		ListenPort:    1099,
		PublicKey:     make([]byte, 32),
		NetworkConfig: DefaultNetworkConfig,
		MachineID:     "123",
	}

	mappings := make(map[uint32]*Mapping)
	mapping, err := GenerateLocalMapping(cfg, mappings)
	if err != nil {
		t.Fatal(err)
	}

	if !testEq(mapping.PrivateIP.To4(), cfg.PrivateIP.To4()) {
		t.Fatal("GenerateLocalMapping created the wrong mapping.")
	}

	mappings[IPtoInt(cfg.PrivateIP)] = mapping

	_, err = GenerateLocalMapping(cfg, mappings)
	if err != nil {
		t.Fatal(err)
	}

	mapping.MachineID = "456"

	_, err = GenerateLocalMapping(cfg, mappings)
	if err == nil {
		t.Fatal("GenerateLocalMapping failed to properly handle an existing ip address")
	}

	cfg.PrivateIP = nil
	mapping.MachineID = "123"

	_, err = GenerateLocalMapping(cfg, mappings)
	if err != nil {
		t.Fatal(err)
	}

	cfg.PrivateIP = nil
	_, err = GenerateLocalMapping(cfg, make(map[uint32]*Mapping))
	if err != nil {
		t.Fatal(err)
	}
}
