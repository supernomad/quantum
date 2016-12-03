package device

import (
	"github.com/Supernomad/quantum/common"
	"github.com/vishvananda/netlink"
	"net"
)

const (
	// TUNDevice type
	TUNDevice int = 0
	// TAPDevice type
	TAPDevice int = 1
	// MOCKDevice type
	MOCKDevice int = 2

	ifNameSize    = 16
	iffTun        = 0x0001
	iffTap        = 0x0002
	iffNoPi       = 0x1000
	iffMultiQueue = 0x0100
)

type ifReq struct {
	Name  [ifNameSize]byte
	Flags uint16
}

// Device is a generic multi-queue network device
type Device interface {
	Name() string
	Read(buf []byte, queue int) (*common.Payload, bool)
	Write(payload *common.Payload, queue int) bool
	Open() error
	Close() error
	Queues() []int
}

// New Device object
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
