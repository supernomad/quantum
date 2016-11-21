package workers

import (
	"github.com/Supernomad/quantum/common"
	"math/rand"
	"testing"
	"time"
)

func benchmarkOutgoingPipeline(buf []byte, queue int, b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if !outgoing.pipeline(buf, queue) {
			panic("Somthing is wrong.")
		}
	}
}

func BenchmarkOutgoingPipeline(b *testing.B) {
	buf := make([]byte, common.MaxPacketLength)
	rand.Read(buf)

	benchmarkOutgoingPipeline(buf, 0, b)
}

func TestOutgoingPipeline(t *testing.T) {
	buf := make([]byte, common.MaxPacketLength)
	rand.Read(buf)

	if !outgoing.pipeline(buf, 0) {
		panic("Somthing is wrong.")
	}
}

func TestOutgoing(t *testing.T) {
	outgoing.Start(0)
	time.Sleep(2 * time.Second)
	outgoing.Stop()
}
