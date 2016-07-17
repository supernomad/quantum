package workers

import (
	"crypto/rand"
	"encoding/binary"
	"github.com/Supernomad/quantum/common"
	"testing"
)

func benchmarkIncomingResolve(payload *common.Payload, b *testing.B) {
	b.ResetTimer()
	// really compiler optimizations... really
	var m *common.Mapping
	var p *common.Payload

	for n := 0; n < b.N; n++ {
		if resolved, mapping, pass := incoming.Resolve(payload); pass {
			m = mapping
			p = resolved
		} else {
			panic("Resolve failed something is wrong.")
		}
	}
	resolveIncomingResult = p
	resolveIncomingMapping = m
}

func BenchmarkIncomingResolve(b *testing.B) {
	buf := make([]byte, common.MaxPacketLength)
	payload := common.NewSockPayload(buf, common.MaxPacketLength)

	binary.LittleEndian.PutUint32(payload.IPAddress, intLocalIP)
	benchmarkIncomingResolve(payload, b)
}

func benchmarkUnseal(payload []*common.Payload, mapping *common.Mapping, b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if unsealed, pass := incoming.Unseal(payload[n], mapping); pass {
			unsealResult = unsealed
		} else {
			panic("Unseal failed something is wrong.")
		}
	}
}

func BenchmarkUnseal(b *testing.B) {
	buf := make([]byte, common.MaxPacketLength)
	rand.Read(buf)

	payload := common.NewTunPayload(buf, common.MTU)
	if sealed, pass := outgoing.Seal(payload, testMapping); pass {
		payloads := make([]*common.Payload, b.N)
		for n := 0; n < b.N; n++ {
			newBuf := make([]byte, common.MaxPacketLength)
			copy(newBuf, sealed.Raw)
			payloads[n] = common.NewSockPayload(newBuf, common.MaxPacketLength)
		}
		benchmarkUnseal(payloads, testMapping, b)
	} else {
		panic("Seal failed something is wrong")
	}
}
