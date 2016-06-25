package tun

import (
	"bytes"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/logger"
	"github.com/vishvananda/netlink"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

const (
	IF_NAME_SIZE    = 16
	IFF_TUN         = 0x0001
	IFF_NO_PI       = 0x1000
	IFF_MULTI_QUEUE = 0x8000
)

type Tun struct {
	Name  string
	log   *logger.Logger
	queue *os.File
}

func (tun *Tun) Close() error {
	return tun.queue.Close()
}

func (tun *Tun) Read() (*common.Payload, bool) {
	buf := make([]byte, common.MaxPacketLength)
	n, err := tun.queue.Read(buf[common.PacketStart:])

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

func (tun *Tun) Write(payload *common.Payload) bool {
	_, err := tun.queue.Write(payload.Packet)
	if err != nil {
		tun.log.Warn("[TUN]", "Write Error:", err)
		return false
	}
	return true
}

func New(ifPattern string, cidr string, log *logger.Logger) (*Tun, error) {
	ifName, queue, err := createTun(ifPattern)
	if err != nil {
		return nil, err
	}

	err = initTun(ifName, cidr)
	if err != nil {
		return nil, err
	}

	return &Tun{ifName, log, queue}, nil
}

type ifReq struct {
	Name  [IF_NAME_SIZE]byte
	Flags uint16
}

func initTun(name, cidr string) error {
	quantum0, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}
	addr, err := netlink.ParseAddr(cidr)
	if err != nil {
		return err
	}
	err = netlink.LinkSetUp(quantum0)
	if err != nil {
		return err
	}
	return netlink.AddrAdd(quantum0, addr)
}

func createTun(name string) (string, *os.File, error) {
	var req ifReq
	req.Flags = syscall.IFF_TUN | syscall.IFF_NO_PI

	copy(req.Name[:], bytes.TrimRight([]byte(name), "\000"))

	file, err := os.OpenFile("/dev/net/tun", os.O_RDWR, 0)
	if err != nil {
		file.Close()
		return "", nil, err
	}

	_, _, errNo := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), uintptr(syscall.TUNSETIFF), uintptr(unsafe.Pointer(&req)))
	if errNo != 0 {
		file.Close()
		return "", nil, err
	}

	return string(req.Name[:]), file, nil
}
