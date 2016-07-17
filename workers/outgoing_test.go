package workers

import (
	"crypto/rand"
	"encoding/binary"
	"github.com/Supernomad/quantum/common"
	"testing"
)

func benchmarkOutgoingResolve(payload *common.Payload, b *testing.B) {
	b.ResetTimer()
	// really compiler optimizations... really
	var m *common.Mapping
	var p *common.Payload

	for n := 0; n < b.N; n++ {
		if resolved, mapping, pass := outgoing.Resolve(payload); pass {
			m = mapping
			p = resolved
		} else {
			panic("Resolve failed something is wrong.")
		}
	}
	resolveOutgoingResult = p
	resolveOutgoingMapping = m
}

func BenchmarkOutgoingResolve(b *testing.B) {
	buf := make([]byte, common.MaxPacketLength)
	payload := common.NewTunPayload(buf, common.MTU)

	binary.LittleEndian.PutUint32(payload.Packet[16:20], intRemoteIP)
	benchmarkOutgoingResolve(payload, b)
}

func benchmarkSeal(payload *common.Payload, mapping *common.Mapping, b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if sealed, pass := outgoing.Seal(payload, mapping); pass {
			sealResult = sealed
		} else {
			panic("Seal failed something is wrong.")
		}
	}
}

func BenchmarkSeal(b *testing.B) {
	buf := make([]byte, common.MaxPacketLength)
	rand.Read(buf)

	payload := common.NewTunPayload(buf, common.MTU)

	benchmarkSeal(payload, testMapping, b)
}
