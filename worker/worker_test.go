// Copyright (c) 2016-2017 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package worker

import (
	"crypto/rand"
	"net"
	"testing"
	"time"

	"github.com/supernomad/quantum/common"
	"github.com/supernomad/quantum/datastore"
	"github.com/supernomad/quantum/device"
	"github.com/supernomad/quantum/metric"
	"github.com/supernomad/quantum/plugin"
	"github.com/supernomad/quantum/router"
	"github.com/supernomad/quantum/socket"
)

var (
	testMapping *common.Mapping
	outgoing    *Outgoing
	incoming    *Incoming
	store       *datastore.Mock
	cfg         *common.Config
	rt          *router.Router

	dev       device.Device
	sock      socket.Socket
	privateIP = "10.1.1.1"
)

func init() {
	ip := net.ParseIP("10.8.0.1")
	ipv6 := net.ParseIP("dead::beef")

	base, ipnet, _ := net.ParseCIDR("10.8.0.0/24")

	store = &datastore.Mock{}
	dev, _ = device.New(device.MOCKDevice, nil)
	sock, _ = socket.New(socket.MOCKSocket, nil)

	key := make([]byte, 32)
	rand.Read(key)

	testMapping = &common.Mapping{IPv4: ip, IPv6: ipv6}
	store.InternalMapping = testMapping

	aggregator := metric.New(
		&common.Config{
			Log:        common.NewLogger(common.NoopLogger),
			NumWorkers: 1,
		})
	aggregator.Start()

	netCfg := &common.NetworkConfig{BaseIP: base, IPNet: ipnet}
	cfg = &common.Config{NumWorkers: 1, PrivateIP: ip, IsIPv6Enabled: true, IsIPv4Enabled: true, NetworkConfig: netCfg}
	rt = router.New(cfg, store)

	incoming = NewIncoming(cfg, aggregator, rt, []plugin.Plugin{}, dev, sock)
	outgoing = NewOutgoing(cfg, aggregator, rt, []plugin.Plugin{}, dev, sock)
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
	benchmarkIncomingPipeline(payload.Raw, 0, b)
}

func TestIncomingPipeline(t *testing.T) {
	buf := make([]byte, common.MaxPacketLength)
	rand.Read(buf)

	payload := common.NewTunPayload(buf, common.MTU)
	if !incoming.pipeline(payload.Raw, 0) {
		panic("Pipeline failed something is wrong.")
	}
}

func TestIncoming(t *testing.T) {
	incoming.Start(0)
	time.Sleep(5 * time.Millisecond)
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
	time.Sleep(5 * time.Millisecond)
	outgoing.Stop()
}
