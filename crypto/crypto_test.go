// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package crypto

import (
	"crypto/rand"
	"net"
	"sync"
	"syscall"
	"testing"
)

const (
	caFile         = "../dist/ssl/certs/ec-ca.crt"
	serverCertFile = "../dist/ssl/certs/ec-server.crt"
	serverKeyFile  = "../dist/ssl/keys/ec-server.key"
	clientCertFile = "../dist/ssl/certs/ec-client.crt"
	clientKeyFile  = "../dist/ssl/keys/ec-client.key"
	tagLen         = 16
	nonceLen       = 12
	bufLen         = 1500
	dataLen        = bufLen - tagLen - nonceLen
)

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

func fillSlice(buf []byte) {
	for i := 0; i < len(buf); i++ {
		buf[i] = 1
	}
}

func TestAES(t *testing.T) {
	key := []byte("AES256Key-32Characters1234567890")
	salt := make([]byte, SaltLength)

	_, err := rand.Read(salt)
	if err != nil {
		t.Fatalf("Unable to random salt: %s", err.Error())
	}

	aes, err := NewAES(key, salt)
	if err != nil {
		t.Fatalf("Unable to create the AES object: %s", err.Error())
	}

	buf := make([]byte, bufLen)
	expected := make([]byte, dataLen)
	fillSlice(buf[:dataLen])
	fillSlice(expected)

	minSize := aes.EncryptedSize(buf)
	if minSize != len(buf)+tagLen+nonceLen {
		t.Fatalf("The AES minimum size is incorrect, got: %d", minSize)
	}

	err = aes.Encrypt(buf, dataLen, nil)
	if err != nil {
		t.Fatalf("Errored trying to encrypt buffer: %s", err.Error())
	}

	if testEq(buf[:dataLen], expected) {
		t.Fatal("Encrypted output matches plaintext.")
	}

	err = aes.Decrypt(buf, nil)
	if err != nil {
		t.Fatalf("Errored trying to decrypt buffer: %s", err.Error())
	}

	if !testEq(buf[:dataLen], expected) {
		t.Fatal("Decrypted output does not match plaintext.")
	}
}

func BenchmarkAES(b *testing.B) {
	key := []byte("AES256Key-32Characters1234567890")
	salt := make([]byte, SaltLength)

	_, err := rand.Read(salt)
	if err != nil {
		b.Fatalf("Unable to random salt: %s", err.Error())
	}

	aes, err := NewAES(key, salt)
	if err != nil {
		b.Fatalf("Unable to create the AES object: %s", err.Error())
	}

	buf := make([]byte, bufLen)
	fillSlice(buf[:dataLen])

	for i := 0; i < b.N; i++ {
		err = aes.Encrypt(buf, dataLen, nil)
		if err != nil {
			b.Fatalf("Errored trying to encrypt buffer: %s", err.Error())
		}

		err = aes.Decrypt(buf, nil)
		if err != nil {
			b.Fatalf("Errored trying to decrypt buffer: %s", err.Error())
		}
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

func testBadCaCert(t *testing.T) {
	_, err := NewServerDTLSContext(-1, "::", 9999, true, true, "path/to/non/existent/CA/certificate/file.crt", serverCertFile, serverKeyFile)
	if err == nil {
		t.Fatal("NewServerDTLSContext failed to pick up a non-existent ca certificate file")
	}

	_, err = NewClientDTLSContext("::", true, true, "path/to/non/existent/CA/certificate/file.crt", serverCertFile, serverKeyFile)
	if err == nil {
		t.Fatal("NewClientDTLSContext failed to pick up a non-existent ca certificate file")
	}
}

func testBadCert(t *testing.T) {
	_, err := NewServerDTLSContext(-1, "::", 9999, true, true, caFile, "path/to/non/existent/certificate/file.crt", serverKeyFile)
	if err == nil {
		t.Fatal("NewServerDTLSContext failed to pick up a non-existent certificate file")
	}

	_, err = NewClientDTLSContext("::", true, true, caFile, "path/to/non/existent/certificate/file.crt", serverKeyFile)
	if err == nil {
		t.Fatal("NewClientDTLSContext failed to pick up a non-existent certificate file")
	}
}

func testBadKey(t *testing.T) {
	_, err := NewServerDTLSContext(-1, "::", 9999, true, true, caFile, serverCertFile, "path/to/non/existent/key/file.pem")
	if err == nil {
		t.Fatal("NewServerDTLSContext failed to pick up a non-existent key file")
	}

	_, err = NewClientDTLSContext("::", true, true, caFile, serverCertFile, "path/to/non/existent/key/file.pem")
	if err == nil {
		t.Fatal("NewClientDTLSContext failed to pick up a non-existent key file")
	}
}

func testMismatchedCertKey(t *testing.T) {
	_, err := NewServerDTLSContext(-1, "::", 9999, true, true, caFile, serverCertFile, clientKeyFile)
	if err == nil {
		t.Fatal("NewServerDTLSContext failed to pick up a mismatched certificate/key pair")
	}

	_, err = NewClientDTLSContext("::", true, true, caFile, serverCertFile, clientKeyFile)
	if err == nil {
		t.Fatal("NewClientDTLSContext failed to pick up a mismatched certificate/key pair")
	}
}

func testEndToEndV4(t *testing.T) {
	done := make(chan bool)

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	if err != nil {
		t.Fatal("error creating the DTLS socket: " + err.Error())
	}
	defer syscall.Close(fd)

	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		t.Fatal("error setting the DTLS socket parameters: " + err.Error())
	}

	sa := &syscall.SockaddrInet4{Port: 9999}
	copy(sa.Addr[:], net.ParseIP("0.0.0.0").To4()[:])

	err = syscall.Bind(fd, sa)
	if err != nil {
		t.Fatal("error binding the DTLS socket to the configured listen address: " + err.Error())
	}

	dtls, err := NewServerDTLSContext(fd, "0.0.0.0", 9999, false, true, caFile, serverCertFile, serverKeyFile)
	if err != nil {
		t.Fatal(err.Error())
	}
	if dtls == nil || dtls.ctx == nil {
		t.Fatal("Failed to create the server DTLS context.")
	}

	cdtls, err := NewClientDTLSContext("0.0.0.0", false, true, caFile, clientCertFile, clientKeyFile)
	if err != nil {
		t.Fatal(err.Error())
	}
	if cdtls == nil || cdtls.ctx == nil {
		t.Fatal("Failed to create the client DTLS context.")
	}

	sendstr := "hello"
	sendbuf := []byte(sendstr)
	sendbufLen := len(sendbuf)
	readbuf := make([]byte, sendbufLen)

	errorstr := ""
	go func() {
		session, err := dtls.Accept()
		if err != nil {
			errorstr = err.Error()
			done <- true
			return
		}

		if session == nil {
			errorstr = "Failed to accept incomming connection."
			done <- true
			return
		}

		n, ok := session.Read(readbuf)
		if !ok {
			errorstr = "Failed to read the buffer correctly."
			done <- true
			return
		}

		if n != sendbufLen || sendstr != string(readbuf[:n]) {
			errorstr = "Failed to read the sent buffer properly."
			done <- true
			return
		}

		done <- false
	}()

	go func() {
		session, err := cdtls.Connect("127.0.0.1", 9999)
		if err != nil {
			errorstr = err.Error()
			done <- true
			return
		}

		if session == nil {
			errorstr = "Failed to connect to the server."
			done <- true
			return
		}

		n, ok := session.Write(sendbuf)
		if !ok {
			errorstr = "Failed to write the buffer correctly."
			done <- true
			return
		}

		if n != sendbufLen {
			errorstr = "Failed to write properly to the server."
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
}

func testEndToEndV6(t *testing.T) {
	done := make(chan bool)

	fd, err := syscall.Socket(syscall.AF_INET6, syscall.SOCK_DGRAM, 0)
	if err != nil {
		t.Fatal("error creating the DTLS socket: " + err.Error())
	}
	defer syscall.Close(fd)

	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		t.Fatal("error setting the DTLS socket parameters: " + err.Error())
	}

	sa := &syscall.SockaddrInet6{Port: 9999}
	copy(sa.Addr[:], net.ParseIP("::").To16()[:])

	err = syscall.Bind(fd, sa)
	if err != nil {
		t.Fatal("error binding the DTLS socket to the configured listen address: " + err.Error())
	}

	dtls, err := NewServerDTLSContext(fd, "::", 9999, true, true, caFile, serverCertFile, serverKeyFile)
	if err != nil {
		t.Fatal(err.Error())
	}
	if dtls == nil || dtls.ctx == nil {
		t.Fatal("Failed to create the server DTLS context.")
	}

	cdtls, err := NewClientDTLSContext("::", true, true, caFile, clientCertFile, clientKeyFile)
	if err != nil {
		t.Fatal(err.Error())
	}
	if cdtls == nil || cdtls.ctx == nil {
		t.Fatal("Failed to create the client DTLS context.")
	}

	sendstr := "hello"
	sendbuf := []byte(sendstr)
	sendbufLen := len(sendbuf)
	readbuf := make([]byte, sendbufLen)

	errorstr := ""
	go func() {
		session, err := dtls.Accept()
		if err != nil {
			errorstr = err.Error()
			done <- true
			return
		}

		if session == nil {
			errorstr = "Failed to accept incomming connection."
			done <- true
			return
		}

		n, ok := session.Read(readbuf)
		if !ok {
			errorstr = "Failed to read the buffer correctly."
			done <- true
			return
		}

		if n != sendbufLen || sendstr != string(readbuf[:n]) {
			errorstr = "Failed to read the sent buffer properly."
			done <- true
			return
		}

		done <- false
	}()

	go func() {
		session, err := cdtls.Connect("::1", 9999)
		if err != nil {
			errorstr = err.Error()
			done <- true
			return
		}

		if session == nil {
			errorstr = "Failed to connect to the server."
			done <- true
			return
		}

		n, ok := session.Write(sendbuf)
		if !ok {
			errorstr = "Failed to write the buffer correctly."
			done <- true
			return
		}

		if n != sendbufLen {
			errorstr = "Failed to write properly to the server."
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
}

func TestDTLS(t *testing.T) {
	InitDTLS()

	t.Run("certificates", func(t *testing.T) {
		t.Run("bad-ca-cert", testBadCaCert)
		t.Run("bad-cert", testBadCert)
		t.Run("bad-key", testBadKey)
		t.Run("mismatched-cert-key-pair", testMismatchedCertKey)
	})

	t.Run("end-to-end", func(t *testing.T) {
		t.Run("IPv4", testEndToEndV4)
		t.Run("IPv6", testEndToEndV6)
	})

	DestroyDTLS()
}

func benchmarkDTLS(server, client *DTLSSession, b *testing.B) {
	sendBuf := make([]byte, 1500)
	recvBuf := make([]byte, 1500)
	for i := 0; i < len(sendBuf); i++ {
		sendBuf[i] = 1
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j, ok := server.Write(sendBuf)
		if !ok || j != len(sendBuf) {
			b.Error("Send error")
		}

		k, ok := client.Read(recvBuf)
		if !ok || j != k {
			b.Error("Read error")
		}
	}
}

func BenchmarkDTLS(b *testing.B) {
	InitDTLS()
	defer DestroyDTLS()

	var wg sync.WaitGroup
	wg.Add(2)

	fd, err := syscall.Socket(syscall.AF_INET6, syscall.SOCK_DGRAM, 0)
	if err != nil {
		b.Fatal("error creating the DTLS socket: " + err.Error())
	}
	defer syscall.Close(fd)

	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		b.Fatal("error setting the DTLS socket parameters: " + err.Error())
	}

	sa := &syscall.SockaddrInet6{Port: 9999}
	copy(sa.Addr[:], net.ParseIP("::").To16()[:])

	err = syscall.Bind(fd, sa)
	if err != nil {
		b.Fatal("error binding the DTLS socket to the configured listen address: " + err.Error())
	}

	dtls, err := NewServerDTLSContext(fd, "::", 9999, true, true, caFile, serverCertFile, serverKeyFile)
	if err != nil {
		b.Fatal(err.Error())
	}
	if dtls == nil || dtls.ctx == nil {
		b.Fatal("Failed to create the server DTLS context.")
	}

	cdtls, err := NewClientDTLSContext("::", true, true, caFile, clientCertFile, clientKeyFile)
	if err != nil {
		b.Fatal(err.Error())
	}
	if cdtls == nil || cdtls.ctx == nil {
		b.Fatal("Failed to create the client DTLS context.")
	}

	var server, client *DTLSSession
	go func() {
		defer wg.Done()
		server, err = dtls.Accept()
		if err != nil {
			b.Error(err.Error())
			return
		}

		if server == nil {
			b.Error("Failed to accept incomming connection.")
			return
		}
	}()

	go func() {
		defer wg.Done()
		client, err = cdtls.Connect("::1", 9999)
		if err != nil {
			b.Error(err.Error())
			return
		}

		if client == nil {
			b.Error("Failed to connect to the server.")
			return
		}
	}()

	wg.Wait()

	benchmarkDTLS(server, client, b)
}
