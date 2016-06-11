package tun

import (
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/logger"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
)

/*
#include <sys/ioctl.h>
#include <sys/socket.h>
#include <linux/if.h>
#include <linux/if_tun.h>

#define IFREQ_SIZE sizeof(struct ifreq)
*/
import "C"

type Tun struct {
	Name   string
	log    *logger.Logger
	queues []*os.File
}

func (tun *Tun) Close() error {
	for i := 0; i < len(tun.queues); i++ {
		err := tun.queues[i].Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (tun *Tun) Read(queue int) (*common.Payload, bool) {
	buf := make([]byte, common.MaxPacketLength)
	n, err := tun.queues[queue].Read(buf[common.PacketStart:])

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
	_, err := tun.queues[queue].Write(payload.Packet)
	if err != nil {
		tun.log.Warn("[TUN]", "Write Error:", err)
		return false
	}
	return true
}

func New(ifPattern string, cidr string, numQueues int, log *logger.Logger) (*Tun, error) {
	ifName, queues, err := createTun(ifPattern, numQueues)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("ip", "link", "set", "dev", ifName, "up")
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	cmd = exec.Command("ip", "addr", "add", cidr, "dev", ifName)
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	return &Tun{ifName, log, queues}, nil
}

type ifReq struct {
	Name  [C.IFNAMSIZ]byte
	Flags uint16
	pad   [C.IFREQ_SIZE - C.IFNAMSIZ - 2]byte
}

func createTun(ifPattern string, numQueues int) (string, []*os.File, error) {
	name := ifPattern
	first := true
	queues := make([]*os.File, numQueues)

	for i := 0; i < numQueues; i++ {
		var req ifReq
		req.Flags = C.IFF_TUN | C.IFF_NO_PI | C.IFF_MULTI_QUEUE

		copy(req.Name[:15], name)

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
		queues[i] = file

		if first {
			first = false
			name = string(req.Name[:strings.Index(string(req.Name[:]), "\000")])
		}
	}

	return name, queues, nil
}
