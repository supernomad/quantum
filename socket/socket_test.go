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
