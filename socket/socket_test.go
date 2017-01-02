// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package socket

import (
	"math/rand"
	"net"
	"syscall"
	"testing"

	"github.com/Supernomad/quantum/common"
)

func benchmarkWrite(sock Socket, payload *common.Payload, queue int, b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if !sock.Write(payload, queue) {
			b.Fatal("Failed to write")
		}
	}
}

func BenchmarkWrite(b *testing.B) {
	addr := net.ParseIP("127.0.0.1")
	paddr := net.ParseIP("127.0.0.2")

	sa := &syscall.SockaddrInet4{Port: 1099}
	copy(sa.Addr[:], addr[:])

	payloadAddr := &syscall.SockaddrInet4{Port: 1099}
	copy(payloadAddr.Addr[:], paddr[:])

	cfg := &common.Config{
		NumWorkers: 1,
		ListenAddr: sa,
	}
	sock, err := New(UDPSocket, cfg)
	if err != nil {
		b.Fatal(err)
	}

	buf := make([]byte, common.MaxPacketLength)
	rand.Read(buf)

	payload := common.NewTunPayload(buf, common.MTU)
	payload.Sockaddr = payloadAddr
	benchmarkWrite(sock, payload, 0, b)
}

func TestMock(t *testing.T) {
	mock, _ := New(MOCKSocket, &common.Config{})
	buf := make([]byte, common.MaxPacketLength)

	payload, ok := mock.Read(buf, 0)
	if payload == nil || !ok {
		t.Fatal("Mock Read should always return a valid payload and nil error.")
	}

	if !mock.Write(payload, 0) {
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
	addr := net.ParseIP("127.0.0.1")
	paddr := net.ParseIP("127.0.0.1")

	sa := &syscall.SockaddrInet4{Port: 1099}
	copy(sa.Addr[:], addr[:])

	payloadAddr := &syscall.SockaddrInet4{Port: 1099}
	copy(payloadAddr.Addr[:], paddr[:])

	cfg := &common.Config{
		NumWorkers: 1,
		ListenAddr: sa,
	}
	sock, err := New(UDPSocket, cfg)
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, common.MaxPacketLength)
	rand.Read(buf)

	payload := common.NewTunPayload(buf, common.MTU)
	payload.Sockaddr = payloadAddr
	if !sock.Write(payload, 0) {
		t.Fatal("Sock Write failed")
	}

	payload, ok := sock.Read(buf, 0)
	if !ok {
		t.Fatal("Sock Read failed")
	}

	if queues := sock.Queues(); len(queues) != 1 {
		t.Fatal("Sock Queues didn't return correctly")
	}

	if err := sock.Close(); err != nil {
		t.Fatal("Sock Close errored")
	}
}
