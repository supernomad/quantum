// Copyright (c) 2016-2018 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package common

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"syscall"
	"testing"
	"time"
)

const (
	ymlConfFile         = "../dist/test/quantum.yml"
	jsonConfFile        = "../dist/test/quantum.json"
	txtConfFile         = "../dist/test/quantum.txt"
	nonExistentConfFile = "../dist/test/doesnt_exist.yml"
)

var (
	testPacket []byte
)

func init() {
	testPacket = make([]byte, 6)
	// IP (1.1.1.1)
	testPacket[0] = 1
	testPacket[1] = 1
	testPacket[2] = 1
	testPacket[3] = 1

	// Packet data
	testPacket[4] = 3
	testPacket[5] = 3
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

func ExampleIPtoInt() {
	ipAddr := net.ParseIP("1.0.0.0")
	ipInt := IPtoInt(ipAddr)

	fmt.Println(ipInt)
	// Output: 1
}

func ExampleIncrementIP() {
	ipAddr := net.ParseIP("0.0.0.1")
	IncrementIP(ipAddr)

	fmt.Println(ipAddr)
	// Output: 0.0.0.2
}

func ExampleArrayEquals() {
	a := []byte{0, 1}
	b := []byte{0, 1}
	c := []byte{1, 1}

	fmt.Println(ArrayEquals(a, b), ArrayEquals(nil, nil), ArrayEquals(a, c), ArrayEquals(a, nil))
	// Output: true true false false
}

func ExampleStringInSlice() {
	slice := []string{"encryption", "compression"}

	fmt.Println(StringInSlice("encryption", slice), StringInSlice("compression", slice), StringInSlice("nonexistent", slice))
	// Output: true true false
}

func TestStringInSlice(t *testing.T) {
	slice := []string{"encryption", "compression"}

	if !StringInSlice("encryption", slice) {
		t.Fatal("StringInSlice returned false checking if string exists in the supplied slice, when the string does indeed exist.")
	}
	if !StringInSlice("compression", slice) {
		t.Fatal("StringInSlice returned false checking if string exists in the supplied slice, when the string does indeed exist.")
	}
	if StringInSlice("nonexistent", slice) {
		t.Fatal("StringInSlice returned true checking if string exists in the supplied slice, when the string does not exist.")
	}
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

func testYamlConfig(t *testing.T, args []string) {
	os.Setenv("QUANTUM_DEVICE_NAME", "different")
	os.Setenv("QUANTUM_LISTEN_PORT", "1")
	os.Setenv("QUANTUM_CONF_FILE", ymlConfFile)
	os.Setenv("QUANTUM_PID_FILE", "../quantum.pid")
	os.Setenv("QUANTUM_FLOATING_IPS", "10.99.1.1,10.99.1.2,hello")
	os.Setenv("_QUANTUM_REAL_DEVICE_NAME_", "quantum0")

	os.Args = append(args, "-n", "100", "--datastore-prefix", "woot", "--datastore-tls-skip-verify", "-6", "fd00:dead:beef::2", "--network", "", "--network-backend", "", "--network-lease-time", "0")
	cfg, err := NewConfig(NewLogger(NoopLogger))
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
	if cfg.DatastorePassword != "Password1" {
		t.Fatalf("NewConfig didn't pick up the config file replacement for Password")
	}
	if cfg.DatastorePrefix != "woot" {
		t.Fatal("NewConfig didn't pick up the cli replacement for Prefix")
	}
	if cfg.NumWorkers != runtime.NumCPU() {
		t.Fatal("NewConfig didn't pick up the cli replacement for NumWorkers")
	}
	if !cfg.DatastoreTLSSkipVerify {
		t.Fatal("NewConfig didn't pick up the cli replacement for DatastoreTLSSkipVerify")
	}
	if len(cfg.Plugins) != 2 {
		t.Fatal("NewConfig didn't pick up file replacement for Plugins")
	}
	if len(cfg.FloatingIPs) != 2 {
		t.Fatal("NewConfig didn't pick up environment variable replacement for FloatingIPs")
	}
	if cfg.FloatingIPs[0].String() != "10.99.1.1" {
		t.Fatal("NewConfig didn't pick up environment variable replacement value for FloatingIPs[0]")
	}
	if cfg.FloatingIPs[1].String() != "10.99.1.2" {
		t.Fatal("NewConfig didn't pick up environment variable replacement value for FloatingIPs[1]")
	}

	// Reset os.Args
	os.Args = args
}

func testJSONConfig(t *testing.T, args []string) {
	os.Setenv("QUANTUM_DEVICE_NAME", "different")
	os.Setenv("QUANTUM_LISTEN_PORT", "1")
	os.Setenv("QUANTUM_CONF_FILE", jsonConfFile)
	os.Setenv("QUANTUM_PID_FILE", "../quantum.pid")
	os.Setenv("QUANTUM_FLOATING_IPS", "")
	os.Setenv("_QUANTUM_REAL_DEVICE_NAME_", "quantum0")

	os.Args = append(args, "-n", "100", "--datastore-prefix", "woot", "--datastore-tls-skip-verify", "-6", "fd00:dead:beef::2", "--network", "", "--network-backend", "", "--network-lease-time", "0")
	cfg, err := NewConfig(NewLogger(NoopLogger))
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
	if cfg.DatastorePassword != "Password1" {
		t.Fatalf("NewConfig didn't pick up the config file replacement for Password")
	}
	if cfg.DatastorePrefix != "woot" {
		t.Fatal("NewConfig didn't pick up the cli replacement for Prefix")
	}
	if cfg.NumWorkers != runtime.NumCPU() {
		t.Fatal("NewConfig didn't pick up the cli replacement for NumWorkers")
	}
	if !cfg.DatastoreTLSSkipVerify {
		t.Fatal("NewConfig didn't pick up the cli replacement for DatastoreTLSSkipVerify")
	}
	if len(cfg.Plugins) != 1 || cfg.Plugins[0] != "compression" {
		t.Fatal("NewConfig didn't pick up file replacement for Plugins")
	}

	// Reset os.Args
	os.Args = args
}

func testInvalidFileConfig(t *testing.T, args []string) {
	os.Setenv("QUANTUM_CONF_FILE", txtConfFile)
	_, err := NewConfig(NewLogger(NoopLogger))
	if err == nil {
		t.Fatal("NewConfig shuld have returned an error for a txt file.")
	}
}

func testNonexistentFileConfig(t *testing.T, args []string) {
	os.Setenv("QUANTUM_CONF_FILE", nonExistentConfFile)
	_, err := NewConfig(NewLogger(NoopLogger))
	if err == nil {
		t.Fatal("NewConfig shuld have returned an error for a nonexistent file.")
	}
}

func testInvalidIntConfig(t *testing.T, args []string) {
	os.Setenv("QUANTUM_WORKERS", "1.23")
	_, err := NewConfig(NewLogger(NoopLogger))
	if err == nil {
		t.Fatal("NewConfig shuld have returned an error for a float passed into an int field.")
	}
	os.Setenv("QUANTUM_WORKERS", "")
}

func testInvalidDurationConfig(t *testing.T, args []string) {
	os.Setenv("QUANTUM_NETWORK_LEASE_TIME", "hello")
	_, err := NewConfig(NewLogger(NoopLogger))
	if err == nil {
		t.Fatal("NewConfig shuld have returned an error for a string passed into a duration field.")
	}
	os.Setenv("QUANTUM_NETWORK_LEASE_TIME", "")
}

func testInvalidIPConfig(t *testing.T, args []string) {
	os.Setenv("QUANTUM_PRIVATE_IP", "w.o.o.t")
	_, err := NewConfig(NewLogger(NoopLogger))
	if err == nil {
		t.Fatal("NewConfig shuld have returned an error for a incorrectly formatted ip address.")
	}
	os.Setenv("QUANTUM_PRIVATE_IP", "")
}

func testInvalidBoolConfig(t *testing.T, args []string) {
	os.Setenv("QUANTUM_DTLS_SKIP_VERIFY", "yes")
	_, err := NewConfig(NewLogger(NoopLogger))
	if err == nil {
		t.Fatal("NewConfig shuld have returned an error for a string pasing into a bool field.")
	}
	os.Setenv("QUANTUM_DTLS_SKIP_VERIFY", "")
}

func testUsageConfig(t *testing.T, args []string) {
	os.Setenv("QUANTUM_PID_FILE", "../quantum.pid")

	cfg, err := NewConfig(NewLogger(NoopLogger))
	if err != nil {
		t.Fatalf("NewConfig returned an error, %s", err)
	}
	if cfg == nil {
		t.Fatal("NewConfig returned a blank config")
	}

	os.Args = append(args, "-h")
	cfg.parseSpecial(false)

	// Reset os.Args
	os.Args = args
}

func testVersionConfig(t *testing.T, args []string) {
	os.Setenv("QUANTUM_PID_FILE", "../quantum.pid")

	cfg, err := NewConfig(NewLogger(NoopLogger))
	if err != nil {
		t.Fatalf("NewConfig returned an error, %s", err)
	}
	if cfg == nil {
		t.Fatal("NewConfig returned a blank config")
	}

	os.Args = append(args, "-v")
	cfg.parseSpecial(false)

	// Reset os.Args
	os.Args = args
}

func TestNewConfig(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("int", func(t *testing.T) {
			testInvalidIntConfig(t, os.Args)
		})
		t.Run("duration", func(t *testing.T) {
			testInvalidDurationConfig(t, os.Args)
		})
		t.Run("ip", func(t *testing.T) {
			testInvalidIPConfig(t, os.Args)
		})
		t.Run("bool", func(t *testing.T) {
			testInvalidBoolConfig(t, os.Args)
		})
	})

	t.Run("special", func(t *testing.T) {
		t.Run("usage", func(t *testing.T) {
			testUsageConfig(t, os.Args)
		})
		t.Run("version", func(t *testing.T) {
			testVersionConfig(t, os.Args)
		})
	})

	t.Run("files", func(t *testing.T) {
		t.Run("json", func(t *testing.T) {
			testJSONConfig(t, os.Args)
		})
		t.Run("invalid", func(t *testing.T) {
			testInvalidFileConfig(t, os.Args)
		})
		t.Run("nonexistent", func(t *testing.T) {
			testNonexistentFileConfig(t, os.Args)
		})
		t.Run("yaml", func(t *testing.T) {
			testYamlConfig(t, os.Args)
		})
	})
}

func TestNewMapping(t *testing.T) {
	cfg := &Config{
		PrivateIP:  net.ParseIP("0.0.0.0"),
		PublicIPv4: net.ParseIP("1.1.1.1"),
		PublicIPv6: net.ParseIP("dead::beef"),
		ListenPort: 80,
		MachineID:  "123456",
	}

	actual := NewMapping(cfg)
	if !testEq(actual.IPv4, cfg.PublicIPv4) || !testEq(actual.IPv6, cfg.PublicIPv6) || actual.Port != cfg.ListenPort || !testEq(actual.PrivateIP, cfg.PrivateIP) {
		t.Fatalf("NewMapping did not return the right value, got: %v", actual)
	}
}

func TestParseMapping(t *testing.T) {
	cfg := &Config{
		PrivateIP:     net.ParseIP("0.0.0.0"),
		PublicIPv4:    net.ParseIP("1.1.1.1"),
		IsIPv4Enabled: true,
		PublicIPv6:    net.ParseIP("dead::beef"),
		IsIPv6Enabled: true,
		ListenPort:    80,
		MachineID:     "123456",
		PublicKey:     []byte("AES256Key-32Characters1234567890"),
		PublicSalt:    []byte("AES256Salt32Characters1234567890"),
	}

	expected := NewMapping(cfg)
	actual, err := ParseMapping(expected.String(), cfg)
	if err != nil {
		t.Fatalf("Error occurred during test: %s", err)
	}
	if !testEq(actual.IPv4, expected.IPv4) || actual.Port != expected.Port || !testEq(actual.PrivateIP, expected.PrivateIP) {
		t.Fatalf("ParseMapping did not return the right value, got: %v, expected: %v", actual, expected)
	}

	cfg.IsIPv6Enabled = false
	expected = NewMapping(cfg)
	actual, err = ParseMapping(expected.String(), cfg)
	if err != nil {
		t.Fatalf("Error occurred during test: %s", err)
	}
	if !testEq(actual.IPv4, expected.IPv4) || actual.Port != expected.Port || !testEq(actual.PrivateIP, expected.PrivateIP) {
		t.Fatalf("ParseMapping did not return the right value, got: %v, expected: %v", actual, expected)
	}

	cfg.IsIPv4Enabled = false
	expected = NewMapping(cfg)
	actual, err = ParseMapping(expected.String(), cfg)
	if err == nil {
		t.Fatalf("Error occurred during test of ParseMapping it should have returned an error and didn't.")
	}
}

func TestParseNetworkConfig(t *testing.T) {
	defaultLeaseTime, _ := time.ParseDuration("48h")
	DefaultNetworkConfig := &NetworkConfig{
		Backend:       "udp",
		Network:       "10.99.0.0/16",
		StaticRange:   "10.99.0.0/23",
		FloatingRange: "10.99.2.0/23",
		LeaseTime:     defaultLeaseTime,
	}

	baseIP, ipnet, _ := net.ParseCIDR(DefaultNetworkConfig.Network)
	DefaultNetworkConfig.BaseIP = baseIP
	DefaultNetworkConfig.IPNet = ipnet

	_, staticNet, _ := net.ParseCIDR(DefaultNetworkConfig.StaticRange)
	DefaultNetworkConfig.StaticNet = staticNet

	_, floatingNet, _ := net.ParseCIDR(DefaultNetworkConfig.FloatingRange)
	DefaultNetworkConfig.FloatingNet = floatingNet

	actual, err := ParseNetworkConfig([]byte(DefaultNetworkConfig.String()))
	if err != nil {
		t.Fatal("ParseNetworkConfig returned an error:", err)
	}
	if actual.Network != DefaultNetworkConfig.Network || actual.LeaseTime != DefaultNetworkConfig.LeaseTime {
		t.Fatalf("ParseNetworkConfig returned the wrong value, got: %v, expected: %v", actual, DefaultNetworkConfig)
	}
}

func TestParseNetworkConfigOnlyNetwork(t *testing.T) {
	netCfg := &NetworkConfig{Network: "10.99.0.0/16"}
	actual, err := ParseNetworkConfig(netCfg.Bytes())
	if err != nil {
		t.Fatal("ParseNetworkConfig returned an error:", err)
	}
	if actual.Network != netCfg.Network || actual.LeaseTime != 48*time.Hour {
		t.Fatalf("ParseNetworkConfig returned the wrong value, got: %v, expected: %v", actual, netCfg)
	}
}

func TestParseNetworkConfigIncorrectFormat(t *testing.T) {
	netCfg := &NetworkConfig{Network: "10.99.0."}
	_, err := ParseNetworkConfig(netCfg.Bytes())
	if err == nil {
		t.Fatal("ParseNetworkConfig should have errored")
	}

	netCfg.Network = "10.99.0.0/16"
	netCfg.StaticRange = "10.99.0./23"

	_, err = ParseNetworkConfig(netCfg.Bytes())
	if err == nil {
		t.Fatal("ParseNetworkConfig should have errored")
	}

	netCfg.StaticRange = "10.20.0.0/23"
	_, err = ParseNetworkConfig(netCfg.Bytes())
	if err == nil {
		t.Fatal("ParseNetworkConfig should have errored")
	}

	netCfg.StaticRange = "10.99.0.0/23"

	netCfg.FloatingRange = "10.99.0./23"
	_, err = ParseNetworkConfig(netCfg.Bytes())
	if err == nil {
		t.Fatal("ParseNetworkConfig should have errored")
	}

	netCfg.FloatingRange = "10.20.0.0/23"
	_, err = ParseNetworkConfig(netCfg.Bytes())
	if err == nil {
		t.Fatal("ParseNetworkConfig should have errored")
	}

	netCfg.FloatingRange = "10.99.0.0/23"
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

	for i := 0; i < 2; i++ {
		if payload.Packet[i] != 3 {
			t.Fatal("NewTunPayload returned an incorrect Packet mapping.")
		}
	}
}

func TestNewSockPayload(t *testing.T) {
	payload := NewSockPayload(testPacket, 6)
	for i := 0; i < 4; i++ {
		if payload.IPAddress[i] != 1 {
			t.Fatal("NewTunPayload returned an incorrect IP address mapping.")
		}
	}

	for i := 0; i < 2; i++ {
		if payload.Packet[i] != 3 {
			t.Fatal("NewTunPayload returned an incorrect Packet mapping.")
		}
	}
}

func TestNewLogger(t *testing.T) {
	log := NewLogger(NoopLogger)
	if log.Error == nil {
		t.Fatal("NewLogger returned a nil Error log.")
	}
	if log.Warn == nil {
		t.Fatal("NewLogger returned a nil Warn log.")
	}
	if log.Info == nil {
		t.Fatal("NewLogger returned a nil Info log.")
	}
	if log.Debug == nil {
		t.Fatal("NewLogger returned a nil Debug log.")
	}
}

func TestGenerateLocalMapping(t *testing.T) {
	defaultLeaseTime, _ := time.ParseDuration("48h")
	DefaultNetworkConfig := &NetworkConfig{
		Backend:     "udp",
		Network:     "10.99.0.0/16",
		StaticRange: "10.99.0.0/23",
		LeaseTime:   defaultLeaseTime,
	}

	baseIP, ipnet, _ := net.ParseCIDR(DefaultNetworkConfig.Network)
	DefaultNetworkConfig.BaseIP = baseIP
	DefaultNetworkConfig.IPNet = ipnet

	_, staticNet, _ := net.ParseCIDR(DefaultNetworkConfig.StaticRange)
	DefaultNetworkConfig.StaticNet = staticNet

	cfg := &Config{
		PrivateIP:     net.ParseIP("10.99.0.1"),
		PublicIPv4:    net.ParseIP("192.167.0.1"),
		PublicIPv6:    net.ParseIP("fd00:dead:beef::2"),
		ListenPort:    1099,
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

func TestGenerateFloatingMapping(t *testing.T) {
	defaultLeaseTime, _ := time.ParseDuration("48h")
	DefaultNetworkConfig := &NetworkConfig{
		Backend:       "udp",
		Network:       "10.99.0.0/16",
		StaticRange:   "10.99.0.0/23",
		FloatingRange: "10.99.2.0/23",
		LeaseTime:     defaultLeaseTime,
	}

	baseIP, ipnet, _ := net.ParseCIDR(DefaultNetworkConfig.Network)
	DefaultNetworkConfig.BaseIP = baseIP
	DefaultNetworkConfig.IPNet = ipnet

	_, staticNet, _ := net.ParseCIDR(DefaultNetworkConfig.StaticRange)
	DefaultNetworkConfig.StaticNet = staticNet

	_, floatingNet, _ := net.ParseCIDR(DefaultNetworkConfig.FloatingRange)
	DefaultNetworkConfig.FloatingNet = floatingNet

	cfg := &Config{
		PrivateIP:     net.ParseIP("10.99.0.1"),
		PublicIPv4:    net.ParseIP("192.167.0.1"),
		PublicIPv6:    net.ParseIP("fd00:dead:beef::2"),
		FloatingIPs:   []net.IP{net.ParseIP("10.99.2.1"), net.ParseIP("10.99.2.2")},
		ListenPort:    1099,
		NetworkConfig: DefaultNetworkConfig,
		MachineID:     "123",
	}

	mappings := make(map[uint32]*Mapping)
	mapping, err := GenerateFloatingMapping(cfg, 0, mappings)
	if err != nil {
		t.Fatal(err)
	}

	if !testEq(mapping.PrivateIP.To4(), cfg.FloatingIPs[0].To4()) {
		t.Fatal("GenerateFloatingMapping created the wrong mapping.")
	}

	mapping.Floating = false

	mappings[IPtoInt(cfg.PrivateIP)] = mapping

	_, err = GenerateFloatingMapping(cfg, 1, mappings)
	if err != nil {
		t.Fatal(err)
	}

	mapping.MachineID = "456"

	_, err = GenerateFloatingMapping(cfg, 0, mappings)
	if err == nil {
		t.Fatal("GenerateFloatingMapping failed to properly handle an existing ip address")
	}
}

func TestSignaler(t *testing.T) {
	log := NewLogger(NoopLogger)
	cfg, err := NewConfig(log)
	signaler := NewSignaler(log, cfg, []int{1}, map[string]string{"QUANTUM_TESTING": "woot"})

	go func() {
		signaler.signals <- syscall.SIGHUP
		signaler.signals <- syscall.SIGINT
	}()

	err = signaler.Wait(false)
	if err != nil {
		t.Fatal("Wait returned an error: " + err.Error())
	}
	err = signaler.Wait(false)
	if err != nil {
		t.Fatal("Wait returned an error: " + err.Error())
	}
}
