// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package device

import (
	"net"

	"github.com/Supernomad/quantum/common"
	"github.com/vishvananda/netlink"
)

// Type defines what kind of virual network device to use.
type Type int

const (
	// TUNDevice type creates and manages a TUN based network device.
	TUNDevice Type = iota

	// MOCKDevice type creates ana manages a mocked our network device for testing.
	MOCKDevice
)

const (
	ifNameSize    = 16
	iffTun        = 0x0001
	iffNoPi       = 0x1000
	iffMultiQueue = 0x0100
)

type ifReq struct {
	Name  [ifNameSize]byte
	Flags uint16
}

// Device interface for a generic multi-queue network device.
type Device interface {
	// Should return the name of the virual network device.
	Name() string

	// Read should return a formatted *common.Payload, based on the provided byte slice, off the specified device queue.
	Read(buf []byte, queue int) (*common.Payload, bool)

	// Write should handle being passed a formatted *common.Payload, and write the underlying raw data to the specified device queue.
	Write(payload *common.Payload, queue int) bool

	// Open should handle creating and configuring the virtual network device.
	Open() error

	// Close should gracefully destroy the virtual network device.
	Close() error

	// Queues should return all underlying queue file descriptors to pass along during a rolling restart.
	Queues() []int
}

// New will generate a new Device struct based on the supplied device kind and user configuration
func New(kind int, cfg *common.Config) Device {
	switch kind {
	case TUNDevice:
		return newTUN(cfg)
	case MOCKDevice:
		return newMock(cfg)
	}
	return nil
}

func initDevice(name, src string, networkCfg *common.NetworkConfig) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}
	err = netlink.LinkSetUp(link)
	if err != nil {
		return err
	}
	err = netlink.LinkSetMTU(link, common.MTU)
	if err != nil {
		return err
	}
	addr, err := netlink.ParseAddr(src + "/32")
	if err != nil {
		return err
	}
	err = netlink.AddrAdd(link, addr)
	if err != nil {
		return err
	}
	route := &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Scope:     netlink.SCOPE_LINK,
		Protocol:  2,
		Src:       net.ParseIP(src),
		Dst:       networkCfg.IPNet,
	}
	return netlink.RouteAdd(route)
}
