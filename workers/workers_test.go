package workers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"github.com/Supernomad/quantum/agg"
	"github.com/Supernomad/quantum/backend"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/device"
	"github.com/Supernomad/quantum/socket"
	"net"
	"sync"
	"testing"
	"time"
)

var (
	outgoing  *Outgoing
	incoming  *Incoming
	dev       device.Device
	sock      socket.Socket
	store     *backend.Mock
	privateIP = "10.1.1.1"
	wg        = &sync.WaitGroup{}
)

var (
	testMapping *common.Mapping
)

func init() {
	ip := net.ParseIP("10.8.0.1")
	ipv6 := net.ParseIP("dead::beef")

	store = &backend.Mock{}
	dev = device.New(device.MOCKDevice, nil)
	sock = socket.New(socket.MOCKSocket, nil)

	key := make([]byte, 32)
	rand.Read(key)

	block, _ := aes.NewCipher(key)
	aesgcm, _ := cipher.NewGCM(block)

	testMapping = &common.Mapping{IPv4: ip, IPv6: ipv6, PublicKey: make([]byte, 32), Cipher: aesgcm}

	store.Mapping = testMapping
	aggregator := agg.New(
		common.NewLogger(),
		&common.Config{
			StatsRoute:   "/stats",
			StatsPort:    1099,
			StatsAddress: "127.0.0.1",
			NumWorkers:   1,
		})
	wg.Add(1)
	aggregator.Start(wg)

	incoming = NewIncoming(&common.Config{NumWorkers: 1, PrivateIP: ip, IsIPv6Enabled: true, IsIPv4Enabled: true}, aggregator, store, dev, sock)
	outgoing = NewOutgoing(&common.Config{NumWorkers: 1, PrivateIP: ip, IsIPv6Enabled: true, IsIPv4Enabled: true}, aggregator, store, dev, sock)
}

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
