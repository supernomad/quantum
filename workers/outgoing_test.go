package workers

import (
	"encoding/binary"
	"github.com/Supernomad/quantum/common"
	"math/rand"
	"testing"
	"time"
)

func benchmarkOutgoingResolve(payload *common.Payload, b *testing.B) {
	b.ResetTimer()
	// really compiler optimizations... really
	var m *common.Mapping
	var p *common.Payload

	for n := 0; n < b.N; n++ {
		if resolved, mapping, pass := outgoing.resolve(payload); pass {
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

func benchmarkOutgoingSeal(payload *common.Payload, mapping *common.Mapping, b *testing.B) {
	b.ResetTimer()
	var s *common.Payload
	for n := 0; n < b.N; n++ {
		if sealed, pass := outgoing.seal(payload, mapping); pass {
			s = sealed
		} else {
			panic("Seal failed something is wrong.")
		}
	}
	sealResult = s
}

func BenchmarkOutgoingSeal(b *testing.B) {
	buf := make([]byte, common.MaxPacketLength)
	rand.Read(buf)

	payload := common.NewTunPayload(buf, common.MTU)

	benchmarkOutgoingSeal(payload, testMapping, b)
}

func benchmarkOutgoingStats(payload *common.Payload, mapping *common.Mapping, b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		outgoing.stats(payload, mapping, 0)
	}
}

func BenchmarkOutgoingStats(b *testing.B) {
	buf := make([]byte, common.MaxPacketLength)
	payload := common.NewTunPayload(buf, common.MTU)
	benchmarkOutgoingStats(payload, testMapping, b)
}

func benchmarkOutgoingPipeline(buf []byte, queue int, b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		binary.LittleEndian.PutUint32(buf[common.PacketStart+16:common.PacketStart+20], intRemoteIP)
		if !outgoing.pipeline(buf, queue) {
			panic("Somthing is wrong.")
		}
	}
}

func BenchmarkOutgoingPipeline(b *testing.B) {
	buf := make([]byte, common.MaxPacketLength)

	benchmarkOutgoingPipeline(buf, 0, b)
}

func TestOutgoingPipeline(t *testing.T) {
	buf := make([]byte, common.MaxPacketLength)
	binary.LittleEndian.PutUint32(buf[common.PacketStart+16:common.PacketStart+20], intRemoteIP)
	if !outgoing.pipeline(buf, 0) {
		panic("Somthing is wrong.")
	}
}

func TestOutgoing(t *testing.T) {
	outgoing.Start(0)
	time.Sleep(2 * time.Second)
	outgoing.Stop()
}
