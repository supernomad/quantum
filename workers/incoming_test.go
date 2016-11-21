package workers

import (
	"github.com/Supernomad/quantum/common"
	"math/rand"
	"testing"
	"time"
)

func benchmarkIncomingPipeline(buf []byte, queue int, b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		incoming.pipeline(buf, queue)
	}
}

func BenchmarkIncomingPipeline(b *testing.B) {
	buf := make([]byte, common.MaxPacketLength)
	rand.Read(buf)

	payload := common.NewTunPayload(buf, common.MTU)
	if sealed, pass := outgoing.seal(payload, testMapping); pass {
		benchmarkIncomingPipeline(sealed.Raw, 0, b)
	} else {
		panic("Seal failed something is wrong")
	}

}

func TestIncomingPipeline(t *testing.T) {
	buf := make([]byte, common.MaxPacketLength)
	rand.Read(buf)

	payload := common.NewTunPayload(buf, common.MTU)
	if sealed, pass := outgoing.seal(payload, testMapping); pass {
		if !incoming.pipeline(sealed.Raw, 0) {
			panic("Somthing is wrong.")
		}
	} else {
		panic("Seal failed something is wrong")
	}
}

func TestIncoming(t *testing.T) {
	incoming.Start(0)
	time.Sleep(2 * time.Second)
	incoming.Stop()
}
