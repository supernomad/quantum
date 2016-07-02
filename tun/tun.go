package tun

import (
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/logger"
	"github.com/vishvananda/netlink"
	"strings"
	"syscall"
	"unsafe"
)

const (
	IF_NAME_SIZE    = 16
	IFF_TUN         = 0x0001
	IFF_NO_PI       = 0x1000
	IFF_MULTI_QUEUE = 0x0100
)

type ifReq struct {
	Name  [IF_NAME_SIZE]byte
	Flags uint16
}

type Tun struct {
	Name   string
	log    *logger.Logger
	queues []int
}

func (tun *Tun) Close() error {
	for i := 0; i < len(tun.queues); i++ {
		if err := syscall.Close(tun.queues[i]); err != nil {
			return err
		}
	}
	return nil
}

func (tun *Tun) Read(buf []byte, queue int) (*common.Payload, bool) {
	n, err := syscall.Read(tun.queues[queue], buf[common.PacketStart:])

	if err != nil {
		tun.log.Warn("[TUN]", "Read Error:", err)
		return nil, false
	}

	if buf[common.PacketStart]>>4 != 4 {
		tun.log.Error("[TUN]", "Unknown IP version recieved")
		return nil, false
	}

	return common.NewTunPayload(buf, n), true
}

func (tun *Tun) Write(payload *common.Payload, queue int) bool {
	_, err := syscall.Write(tun.queues[queue], payload.Packet)
	if err != nil {
		tun.log.Warn("[TUN]", "Write Error:", err)
		return false
	}
	return true
}

func New(ifPattern string, cidr string, numWorkers int, log *logger.Logger) (*Tun, error) {
	queues := make([]int, numWorkers)
	first := true
	name := ifPattern

	for i := 0; i < numWorkers; i++ {
		ifName, queue, err := createTun(name)
		if err != nil {
			return nil, err
		}
		queues[i] = queue

		if first {
			first = false
			name = ifName
		}
	}

	err := initTun(name, cidr)
	if err != nil {
		return nil, err
	}

	return &Tun{Name: name, queues: queues, log: log}, nil
}

func initTun(name, cidr string) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}
	addr, err := netlink.ParseAddr(cidr)
	if err != nil {
		return err
	}
	err = netlink.LinkSetUp(link)
	if err != nil {
		return err
	}
	return netlink.AddrAdd(link, addr)
}

func createTun(name string) (string, int, error) {
	var req ifReq
	req.Flags = IFF_TUN | IFF_NO_PI | IFF_MULTI_QUEUE

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
