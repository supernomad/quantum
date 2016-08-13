package socket

import (
	"github.com/Supernomad/quantum/common"
	"github.com/vishvananda/netlink"
	"net"
	"strings"
	"syscall"
	"unsafe"
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

// Tun interface
type Tun struct {
	name   string
	queues []int
	cfg    *common.Config
}

// Name of the Tun device
func (tun *Tun) Name() string {
	return tun.name
}

// Open the Tun device for communication to begin
func (tun *Tun) Open() error {
	first := true
	for i := 0; i < tun.cfg.NumWorkers; i++ {
		ifName, queue, err := createTUN(tun.name)
		if err != nil {
			return err
		}
		tun.queues[i] = queue

		if first {
			first = false
			tun.name = ifName
		}
	}

	err := initTUN(tun.name, tun.cfg.PrivateIP, tun.cfg.NetworkConfig)
	if err != nil {
		return err
	}
	return nil
}

// Close the tun
func (tun *Tun) Close() error {
	for i := 0; i < len(tun.queues); i++ {
		if err := syscall.Close(tun.queues[i]); err != nil {
			return err
		}
	}
	return nil
}

// Read a packet off the tun
func (tun *Tun) Read(buf []byte, queue int) (*common.Payload, bool) {
	n, err := syscall.Read(tun.queues[queue], buf[common.PacketStart:])
	if err != nil {
		return nil, false
	}
	return common.NewTunPayload(buf, n), true
}

// Write a packet to the tun
func (tun *Tun) Write(payload *common.Payload, mapping *common.Mapping, queue int) bool {
	_, err := syscall.Write(tun.queues[queue], payload.Packet)
	if err != nil {
		return false
	}
	return true
}

func newTUN(cfg *common.Config) Socket {
	queues := make([]int, cfg.NumWorkers)
	name := cfg.InterfaceName

	return &Tun{name: name, cfg: cfg, queues: queues}
}

func initTUN(name, src string, networkCfg *common.NetworkConfig) error {
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

func createTUN(name string) (string, int, error) {
	var req ifReq
	req.Flags = iffTun | iffNoPi | iffMultiQueue

	copy(req.Name[:15], name)

	queue, err := syscall.Open("/dev/net/tun", syscall.O_RDWR, 0)
	if err != nil {
		syscall.Close(queue)
		return "", -1, err
	}

	_, _, errNo := syscall.Syscall(syscall.SYS_IOCTL, uintptr(queue), uintptr(syscall.TUNSETIFF), uintptr(unsafe.Pointer(&req)))
	if errNo != 0 {
		syscall.Close(queue)
		return "", -1, err
	}

	return string(req.Name[:strings.Index(string(req.Name[:]), "\000")]), queue, nil
}