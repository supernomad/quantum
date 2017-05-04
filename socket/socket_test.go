// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package socket

import (
	"net"
	"os"
	"syscall"
	"testing"

	"github.com/Supernomad/quantum/common"
)

const (
	caFile         = "../dist/ssl/certs/ec-ca.crt"
	serverCertFile = "../dist/ssl/certs/ec-server.crt"
	serverKeyFile  = "../dist/ssl/keys/ec-server.key"
	clientCertFile = "../dist/ssl/certs/ec-client.crt"
	clientKeyFile  = "../dist/ssl/keys/ec-client.key"
)

func TestDTLS(t *testing.T) {
	if os.Getenv("IS_TRAVIS") == "true" {
		t.Skip("skipping end-to-end tests, these tests don't work in travis-ci")
	}

	listenIP := net.ParseIP("::")
	localAddr := &syscall.SockaddrInet6{Port: 9999}
	copy(localAddr.Addr[:], listenIP.To16()[:])

	localCfg := &common.Config{
		NumWorkers:     1,
		ReuseFDS:       false,
		IsIPv6Enabled:  true,
		DTLSCA:         caFile,
		DTLSCert:       serverCertFile,
		DTLSKey:        serverKeyFile,
		DTLSSkipVerify: false,
		ListenAddr:     localAddr,
		ListenIP:       listenIP,
		ListenPort:     9999,
		Log:            common.NewLogger(common.NoopLogger),
	}

	localSocket, err := New(DTLSSocket, localCfg)
	if err != nil {
		t.Fatalf("Failed generating the local dtls socket: %s", err.Error())
	}

	defer localSocket.Close()

	remoteAddr := &syscall.SockaddrInet6{Port: 9998}
	copy(remoteAddr.Addr[:], listenIP.To16()[:])

	remoteCfg := &common.Config{
		NumWorkers:     1,
		ReuseFDS:       false,
		IsIPv6Enabled:  true,
		DTLSCA:         caFile,
		DTLSCert:       serverCertFile,
		DTLSKey:        serverKeyFile,
		DTLSSkipVerify: false,
		ListenAddr:     remoteAddr,
		ListenIP:       listenIP,
		ListenPort:     9998,
		Log:            common.NewLogger(common.NoopLogger),
	}

	remoteMapping := &common.Mapping{
		Address: "::1",
		Port:    9998,
	}
	remoteSocket, err := New(DTLSSocket, remoteCfg)
	if err != nil {
		t.Fatalf("Failed generating the remote dtls socket: %s", err.Error())
	}

	defer remoteSocket.Close()

	sendbuf := []byte("hello")
	recvbuf := make([]byte, 2048)

	ok := localSocket.Write(0, &common.Payload{Raw: sendbuf, Length: len(sendbuf)}, remoteMapping)
	if !ok {
		t.Fatal("Failed writing to the remote peer.")
	}

	payload, ok := remoteSocket.Read(0, recvbuf)
	if !ok {
		t.Fatal("Failed reading from the local peer.")
	}

	if payload.Length != len(sendbuf) {
		t.Fatal("Read incorrect length")
	}

	if string(payload.Raw[:payload.Length]) != "hello" {
		t.Fatalf("Read incorrect data, got: %s", string(payload.Raw[:payload.Length]))
	}
}

func TestMock(t *testing.T) {
	if os.Getenv("IS_TRAVIS") == "true" {
		t.Skip("skipping end-to-end tests, these tests don't work in travis-ci")
	}

	mock, _ := New(MOCKSocket, &common.Config{})
	buf := make([]byte, common.MaxPacketLength)

	payload, ok := mock.Read(0, buf)
	if payload == nil || !ok {
		t.Fatal("Mock Read should always return a valid payload and nil error.")
	}

	if !mock.Write(0, payload, nil) {
		t.Fatal("Mock Write should always return true.")
	}

	if mock.Queues() != nil {
		t.Fatal("Mock Queues should always return nil.")
	}

	if mock.Close() != nil {
		t.Fatal("Mock Close should always return nil.")
	}
}

func TestUDP(t *testing.T) {
	if os.Getenv("IS_TRAVIS") == "true" {
		t.Skip("skipping end-to-end tests, these tests don't work in travis-ci")
	}

	listenIP := net.ParseIP("::")
	localAddr := &syscall.SockaddrInet6{Port: 9999}
	copy(localAddr.Addr[:], listenIP.To16()[:])

	localCfg := &common.Config{
		NumWorkers:    1,
		ReuseFDS:      false,
		IsIPv6Enabled: true,
		ListenAddr:    localAddr,
		ListenIP:      listenIP,
		ListenPort:    9999,
		Log:           common.NewLogger(common.NoopLogger),
	}

	localSocket, err := New(UDPSocket, localCfg)
	if err != nil {
		t.Fatalf("Failed generating the local dtls socket: %s", err.Error())
	}

	remoteAddr := &syscall.SockaddrInet6{Port: 9998}
	copy(remoteAddr.Addr[:], listenIP.To16()[:])

	remoteCfg := &common.Config{
		NumWorkers:    1,
		ReuseFDS:      false,
		IsIPv6Enabled: true,
		ListenAddr:    remoteAddr,
		ListenIP:      listenIP,
		ListenPort:    9998,
		Log:           common.NewLogger(common.NoopLogger),
	}

	remoteMapping := &common.Mapping{
		Sockaddr: remoteAddr,
		Address:  "::1",
		Port:     9998,
	}
	remoteSocket, err := New(UDPSocket, remoteCfg)
	if err != nil {
		t.Fatalf("Failed generating the remote dtls socket: %s", err.Error())
	}

	sendbuf := []byte("hello")
	recvbuf := make([]byte, 2048)

	ok := localSocket.Write(0, &common.Payload{Raw: sendbuf, Length: len(sendbuf)}, remoteMapping)
	if !ok {
		t.Fatal("Failed writing to the remote peer.")
	}

	payload, ok := remoteSocket.Read(0, recvbuf)
	if !ok {
		t.Fatal("Failed reading from the local peer.")
	}

	if payload.Length != len(sendbuf) {
		t.Fatal("Read incorrect length")
	}

	if string(payload.Raw[:payload.Length]) != "hello" {
		t.Fatalf("Read incorrect data, got: %s", string(payload.Raw[:payload.Length]))
	}
}
