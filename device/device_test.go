// Copyright (c) 2016-2018 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package device

import (
	"net"
	"syscall"
	"testing"
	"time"

	"github.com/supernomad/quantum/common"
	"golang.org/x/net/ipv4"
)

func TestMock(t *testing.T) {
	mock, _ := New(MOCKDevice, &common.Config{})
	buf := make([]byte, common.MaxPacketLength)

	payload, ok := mock.Read(0, buf)
	if payload == nil || !ok {
		t.Fatal("Mock Read should always return a valid payload and nil error.")
	}

	if mock.Name() == "" {
		t.Fatal("Mock Name should return a non empty string.")
	}

	if !mock.Write(0, payload) {
		t.Fatal("Mock Write should always return true.")
	}

	if mock.Queues() != nil {
		t.Fatal("Mock Queues should always return nil.")
	}

	if mock.Close() != nil {
		t.Fatal("Mock Close should always return nil.")
	}
}

func TestTUN(t *testing.T) {
	defaultLeaseTime, _ := time.ParseDuration("48h")
	DefaultNetworkConfig := &common.NetworkConfig{
		Backend:     "udp",
		Network:     "10.99.0.0/16",
		StaticRange: "10.99.0.0/23",
		LeaseTime:   defaultLeaseTime,
	}

	baseIP, ipnet, _ := net.ParseCIDR(DefaultNetworkConfig.Network)
	DefaultNetworkConfig.BaseIP = baseIP
	DefaultNetworkConfig.IPNet = ipnet

	_, staticNet, _ := net.ParseCIDR(DefaultNetworkConfig.StaticRange)
	DefaultNetworkConfig.StaticNet = staticNet

	tun, err := New(TUNDevice, &common.Config{
		NumWorkers:    1,
		DeviceName:    "quantum%d",
		PrivateIP:     net.ParseIP("10.99.0.1"),
		NetworkConfig: DefaultNetworkConfig,
		ReuseFDS:      false,
	})

	if err != nil {
		t.Fatalf("Failed to create TUN device: %s", err.Error())
	}
	if tun == nil {
		t.Fatal("Failed to create TUN device: unhandled error")
	}

	if tun.Name() != "quantum0" {
		t.Fatal("Failed to properly set the TUN device name.")
	}

	if len(tun.Queues()) != 1 {
		t.Fatal("Failed to properly create the right number of TUN queues.")
	}

	buf := make([]byte, 1024)
	payload, ok := tun.Read(0, buf)
	if !ok {
		t.Fatal("Failed to read the packet off the tun device")
	}

	iph := &ipv4.Header{
		Version:  4,
		Len:      20,
		TOS:      0,
		TotalLen: len(payload.Packet),
		ID:       0,
		Flags:    0,
		FragOff:  0,
		TTL:      128,
		Protocol: syscall.ETH_P_IP,
		Checksum: 0,
		Src:      net.ParseIP("10.99.0.2"),
		Dst:      net.ParseIP("10.99.0.1"),
		Options:  nil,
	}
	iphBuf, _ := iph.Marshal()
	copy(payload.Packet[0:20], iphBuf)

	ok = tun.Write(0, payload)
	if !ok {
		t.Fatal("Failed to write the packet to the tun device")
	}

	if err := tun.Close(); err != nil {
		t.Fatalf("Failed to close the TUN device: %s", err.Error())
	}
}
