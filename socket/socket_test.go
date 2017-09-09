// Copyright (c) 2016-2017 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package socket

import (
	"net"
	"syscall"
	"testing"

	"github.com/supernomad/quantum/common"
)

const (
	caFile         = "../dist/ssl/certs/ec-ca.crt"
	serverCertFile = "../dist/ssl/certs/ec-server.crt"
	serverKeyFile  = "../dist/ssl/keys/ec-server.key"
	clientCertFile = "../dist/ssl/certs/ec-client.crt"
	clientKeyFile  = "../dist/ssl/keys/ec-client.key"
)

func TestMock(t *testing.T) {
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

func testUDPEndToEndV4(t *testing.T) {
	done := make(chan bool)

	clientSa := &syscall.SockaddrInet4{Port: 9999}
	copy(clientSa.Addr[:], net.ParseIP("0.0.0.0").To4()[:])

	clientMapping := &common.Mapping{
		Sockaddr: clientSa,
	}

	client, err := New(UDPSocket, &common.Config{
		NumWorkers:    1,
		ReuseFDS:      false,
		IsIPv6Enabled: false,
		ListenAddr:    clientSa,
	})
	if err != nil {
		t.Fatalf("Failed to generate client UDP socket: %s", err.Error())
	}
	if client == nil {
		t.Fatal("Failed to generate client UDP socket: unhandled error")
	}
	if len(client.Queues()) != 1 {
		t.Fatal("Failed to generate client UDP socket: invalid socket queue generation")
	}

	serverSa := &syscall.SockaddrInet4{Port: 9998}
	copy(serverSa.Addr[:], net.ParseIP("0.0.0.0").To4()[:])

	serverMapping := &common.Mapping{
		Sockaddr: serverSa,
	}

	server, err := New(UDPSocket, &common.Config{
		NumWorkers:    1,
		ReuseFDS:      false,
		IsIPv6Enabled: false,
		ListenAddr:    serverSa,
	})
	if err != nil {
		t.Fatalf("Failed to generate server UDP socket: %s", err.Error())
	}
	if server == nil {
		t.Fatal("Failed to generate server UDP socket: unhandled error")
	}
	if len(server.Queues()) != 1 {
		t.Fatal("Failed to generate server UDP socket: invalid socket queue generation")
	}

	sendstr := "hello"
	sendbuf := []byte(sendstr)
	sendbufLen := len(sendbuf)
	readbuf := make([]byte, sendbufLen)

	errorstr := ""
	go func() {
		payload, ok := server.Read(0, readbuf)
		if !ok {
			errorstr = "Failed to read the payload correctly."
			done <- true
			return
		}

		if payload.Length != sendbufLen || sendstr != string(payload.Raw[:payload.Length]) {
			errorstr = "Failed to read the sent payload properly."
			done <- true
			return
		}

		ok = server.Write(0, payload, clientMapping)
		if !ok {
			errorstr = "Failed to write the payload correctly."
			done <- true
			return
		}

		done <- false
	}()

	go func() {
		payload := &common.Payload{
			Raw:    sendbuf,
			Length: sendbufLen,
		}

		ok := client.Write(0, payload, serverMapping)
		if !ok {
			errorstr = "Failed to write the payload correctly."
			done <- true
			return
		}

		recvPayload, ok := client.Read(0, readbuf)
		if !ok {
			errorstr = "Failed to read the payload correctly."
			done <- true
			return
		}

		if recvPayload.Length != sendbufLen || sendstr != string(recvPayload.Raw[:recvPayload.Length]) {
			errorstr = "Failed to read the sent payload properly."
			done <- true
			return
		}

		done <- false
	}()

	for i := 0; i < 2; i++ {
		select {
		case failed := <-done:
			if failed {
				t.Fatal(errorstr)
			}
		}
	}

	client.Close()
	server.Close()
}

func testUDPEndToEndV6(t *testing.T) {
	done := make(chan bool)

	clientSa := &syscall.SockaddrInet6{Port: 9999}
	copy(clientSa.Addr[:], net.ParseIP("::").To4()[:])

	clientMapping := &common.Mapping{
		Sockaddr: clientSa,
	}

	client, err := New(UDPSocket, &common.Config{
		NumWorkers:    1,
		ReuseFDS:      false,
		IsIPv6Enabled: true,
		ListenAddr:    clientSa,
	})
	if err != nil {
		t.Fatalf("Failed to generate client UDP socket: %s", err.Error())
	}
	if client == nil {
		t.Fatal("Failed to generate client UDP socket: unhandled error")
	}
	if len(client.Queues()) != 1 {
		t.Fatal("Failed to generate client UDP socket: invalid socket queue generation")
	}

	serverSa := &syscall.SockaddrInet6{Port: 9998}
	copy(serverSa.Addr[:], net.ParseIP("::").To4()[:])

	serverMapping := &common.Mapping{
		Sockaddr: serverSa,
	}

	server, err := New(UDPSocket, &common.Config{
		NumWorkers:    1,
		ReuseFDS:      false,
		IsIPv6Enabled: true,
		ListenAddr:    serverSa,
	})
	if err != nil {
		t.Fatalf("Failed to generate server UDP socket: %s", err.Error())
	}
	if server == nil {
		t.Fatal("Failed to generate server UDP socket: unhandled error")
	}
	if len(server.Queues()) != 1 {
		t.Fatal("Failed to generate server UDP socket: invalid socket queue generation")
	}

	sendstr := "hello"
	sendbuf := []byte(sendstr)
	sendbufLen := len(sendbuf)
	readbuf := make([]byte, sendbufLen)

	errorstr := ""
	go func() {
		payload, ok := server.Read(0, readbuf)
		if !ok {
			errorstr = "Failed to read the payload correctly."
			done <- true
			return
		}

		if payload.Length != sendbufLen || sendstr != string(payload.Raw[:payload.Length]) {
			errorstr = "Failed to read the sent payload properly."
			done <- true
			return
		}

		ok = server.Write(0, payload, clientMapping)
		if !ok {
			errorstr = "Failed to write the payload correctly."
			done <- true
			return
		}

		done <- false
	}()

	go func() {
		payload := &common.Payload{
			Raw:    sendbuf,
			Length: sendbufLen,
		}

		ok := client.Write(0, payload, serverMapping)
		if !ok {
			errorstr = "Failed to write the payload correctly."
			done <- true
			return
		}

		recvPayload, ok := client.Read(0, readbuf)
		if !ok {
			errorstr = "Failed to read the payload correctly."
			done <- true
			return
		}

		if recvPayload.Length != sendbufLen || sendstr != string(recvPayload.Raw[:recvPayload.Length]) {
			errorstr = "Failed to read the sent payload properly."
			done <- true
			return
		}

		done <- false
	}()

	for i := 0; i < 2; i++ {
		select {
		case failed := <-done:
			if failed {
				t.Fatal(errorstr)
			}
		}
	}

	client.Close()
	server.Close()
}

func TestUDP(t *testing.T) {
	t.Run("end-to-end", func(t *testing.T) {
		t.Run("IPv4", testUDPEndToEndV4)
		t.Run("IPv6", testUDPEndToEndV6)
	})
}

func testDTLSEndToEndV4(t *testing.T) {
	done := make(chan bool)

	lip := net.ParseIP("127.0.0.1").To4()

	clientSa := &syscall.SockaddrInet4{Port: 9999}
	copy(clientSa.Addr[:], lip[:])

	clientMapping := &common.Mapping{
		Address: "127.0.0.1",
		Port:    9999,
	}

	client, err := New(DTLSSocket, &common.Config{
		NumWorkers:     1,
		ReuseFDS:       false,
		DTLSCA:         caFile,
		DTLSCert:       clientCertFile,
		DTLSKey:        clientKeyFile,
		DTLSSkipVerify: false,
		IsIPv6Enabled:  false,
		ListenIP:       lip,
		ListenPort:     9999,
		ListenAddr:     clientSa,
		Log:            common.NewLogger(common.DebugLogger),
	})
	if err != nil {
		t.Fatalf("Failed to generate client UDP socket: %s", err.Error())
	}
	if client == nil {
		t.Fatal("Failed to generate client UDP socket: unhandled error")
	}
	if len(client.Queues()) != 1 {
		t.Fatal("Failed to generate client UDP socket: invalid socket queue generation")
	}

	serverSa := &syscall.SockaddrInet4{Port: 9998}
	copy(serverSa.Addr[:], lip[:])

	serverMapping := &common.Mapping{
		Address: "127.0.0.1",
		Port:    9998,
	}

	server, err := New(DTLSSocket, &common.Config{
		NumWorkers:     1,
		ReuseFDS:       false,
		DTLSCA:         caFile,
		DTLSCert:       serverCertFile,
		DTLSKey:        serverKeyFile,
		DTLSSkipVerify: false,
		IsIPv6Enabled:  false,
		ListenIP:       lip,
		ListenPort:     9998,
		ListenAddr:     serverSa,
		Log:            common.NewLogger(common.DebugLogger),
	})
	if err != nil {
		t.Fatalf("Failed to generate server UDP socket: %s", err.Error())
	}
	if server == nil {
		t.Fatal("Failed to generate server UDP socket: unhandled error")
	}
	if len(server.Queues()) != 1 {
		t.Fatal("Failed to generate server UDP socket: invalid socket queue generation")
	}

	sendstr := "hello"
	sendbuf := []byte(sendstr)
	sendbufLen := len(sendbuf)
	readbuf := make([]byte, sendbufLen)

	errorstr := ""
	go func() {
		payload, ok := server.Read(0, readbuf)
		if !ok {
			errorstr = "Server failed to read the payload correctly."
			done <- true
			return
		}

		if payload.Length != sendbufLen || sendstr != string(payload.Raw[:payload.Length]) {
			errorstr = "Server failed to read the sent payload properly."
			done <- true
			return
		}

		ok = server.Write(0, payload, clientMapping)
		if !ok {
			errorstr = "Server failed to write the payload correctly."
			done <- true
			return
		}

		done <- false
	}()

	go func() {
		payload := &common.Payload{
			Raw:    sendbuf,
			Length: sendbufLen,
		}

		ok := client.Write(0, payload, serverMapping)
		if !ok {
			errorstr = "Client failed to write the payload correctly."
			done <- true
			return
		}

		recvPayload, ok := client.Read(0, readbuf)
		if !ok {
			errorstr = "Client failed to read the payload correctly."
			done <- true
			return
		}

		if recvPayload.Length != sendbufLen || sendstr != string(recvPayload.Raw[:recvPayload.Length]) {
			errorstr = "Client failed to read the sent payload properly."
			done <- true
			return
		}

		done <- false
	}()

	for i := 0; i < 2; i++ {
		select {
		case failed := <-done:
			if failed {
				t.Fatal(errorstr)
			}
		}
	}

	client.Close()
	server.Close()
}

func testDTLSEndToEndV6(t *testing.T) {
	done := make(chan bool)

	lip := net.ParseIP("::1").To16()

	clientSa := &syscall.SockaddrInet6{Port: 9999}
	copy(clientSa.Addr[:], lip[:])

	clientMapping := &common.Mapping{
		Address: "::1",
		Port:    9999,
	}

	client, err := New(DTLSSocket, &common.Config{
		NumWorkers:     1,
		ReuseFDS:       false,
		DTLSCA:         caFile,
		DTLSCert:       clientCertFile,
		DTLSKey:        clientKeyFile,
		DTLSSkipVerify: false,
		IsIPv6Enabled:  true,
		ListenIP:       lip,
		ListenPort:     9999,
		ListenAddr:     clientSa,
		Log:            common.NewLogger(common.DebugLogger),
	})
	if err != nil {
		t.Fatalf("Failed to generate client UDP socket: %s", err.Error())
	}
	if client == nil {
		t.Fatal("Failed to generate client UDP socket: unhandled error")
	}
	if len(client.Queues()) != 1 {
		t.Fatal("Failed to generate client UDP socket: invalid socket queue generation")
	}

	serverSa := &syscall.SockaddrInet6{Port: 9998}
	copy(serverSa.Addr[:], lip[:])

	serverMapping := &common.Mapping{
		Address: "::1",
		Port:    9998,
	}

	server, err := New(DTLSSocket, &common.Config{
		NumWorkers:     1,
		ReuseFDS:       false,
		DTLSCA:         caFile,
		DTLSCert:       serverCertFile,
		DTLSKey:        serverKeyFile,
		DTLSSkipVerify: false,
		IsIPv6Enabled:  true,
		ListenIP:       lip,
		ListenPort:     9998,
		ListenAddr:     serverSa,
		Log:            common.NewLogger(common.DebugLogger),
	})
	if err != nil {
		t.Fatalf("Failed to generate server UDP socket: %s", err.Error())
	}
	if server == nil {
		t.Fatal("Failed to generate server UDP socket: unhandled error")
	}
	if len(server.Queues()) != 1 {
		t.Fatal("Failed to generate server UDP socket: invalid socket queue generation")
	}

	sendstr := "hello"
	sendbuf := []byte(sendstr)
	sendbufLen := len(sendbuf)
	readbuf := make([]byte, sendbufLen)

	errorstr := ""
	go func() {
		payload, ok := server.Read(0, readbuf)
		if !ok {
			errorstr = "Server failed to read the payload correctly."
			done <- true
			return
		}

		if payload.Length != sendbufLen || sendstr != string(payload.Raw[:payload.Length]) {
			errorstr = "Server failed to read the sent payload properly."
			done <- true
			return
		}

		ok = server.Write(0, payload, clientMapping)
		if !ok {
			errorstr = "Server failed to write the payload correctly."
			done <- true
			return
		}

		done <- false
	}()

	go func() {
		payload := &common.Payload{
			Raw:    sendbuf,
			Length: sendbufLen,
		}

		ok := client.Write(0, payload, serverMapping)
		if !ok {
			errorstr = "Client failed to write the payload correctly."
			done <- true
			return
		}

		recvPayload, ok := client.Read(0, readbuf)
		if !ok {
			errorstr = "Client failed to read the payload correctly."
			done <- true
			return
		}

		if recvPayload.Length != sendbufLen || sendstr != string(recvPayload.Raw[:recvPayload.Length]) {
			errorstr = "Client failed to read the sent payload properly."
			done <- true
			return
		}

		done <- false
	}()

	for i := 0; i < 2; i++ {
		select {
		case failed := <-done:
			if failed {
				t.Fatal(errorstr)
			}
		}
	}

	client.Close()
	server.Close()
}

func TestDTLS(t *testing.T) {
	t.Run("end-to-end", func(t *testing.T) {
		t.Run("IPv4", testDTLSEndToEndV4)
		t.Run("IPv6", testDTLSEndToEndV6)
	})
}
