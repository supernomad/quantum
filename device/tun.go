// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package device

import (
	"errors"
	"strings"
	"syscall"
	"unsafe"

	"github.com/Supernomad/quantum/common"
)

// Tun device struct for managing a multi-queue TUN networking device.
type Tun struct {
	name   string
	queues []int
	cfg    *common.Config
}

// Name of the Tun device.
func (tun *Tun) Name() string {
	return tun.name
}

// Close the Tun device and remove associated network configuration.
func (tun *Tun) Close() error {
	for i := 0; i < len(tun.queues); i++ {
		if err := syscall.Close(tun.queues[i]); err != nil {
			return errors.New("error closing the device queues: " + err.Error())
		}
	}
	return nil
}

// Queues returns the underlying device queue file descriptors.
func (tun *Tun) Queues() []int {
	return tun.queues
}

// Read a packet off the specified device queue and return a *common.Payload representation of the packet.
func (tun *Tun) Read(buf []byte, queue int) (*common.Payload, bool) {
	n, err := syscall.Read(tun.queues[queue], buf[common.PacketStart:])
	if err != nil {
		return nil, false
	}
	return common.NewTunPayload(buf, n), true
}

// Write a *common.Payload to the specified device queue.
func (tun *Tun) Write(payload *common.Payload, queue int) bool {
	_, err := syscall.Write(tun.queues[queue], payload.Packet)
	if err != nil {
		return false
	}
	return true
}

func newTUN(cfg *common.Config) (Device, error) {
	queues := make([]int, cfg.NumWorkers)
	name := cfg.DeviceName
	tun := &Tun{name: name, cfg: cfg, queues: queues}

	for i := 0; i < tun.cfg.NumWorkers; i++ {
		if !tun.cfg.ReuseFDS {
			ifName, queue, err := createTUN(tun.name)
			if err != nil {
				return nil, err
			}
			tun.queues[i] = queue
			tun.name = ifName
		} else {
			tun.queues[i] = 3 + i
			tun.name = tun.cfg.RealDeviceName
		}
	}

	if !tun.cfg.ReuseFDS {
		err := initDevice(tun.name, tun.cfg.PrivateIP.String(), tun.cfg.NetworkConfig)
		if err != nil {
			return nil, err
		}
	}

	return tun, nil
}

func createTUN(name string) (string, int, error) {
	var req ifReq
	req.Flags = iffTun | iffNoPi | iffMultiQueue

	copy(req.Name[:15], name)

	queue, err := syscall.Open("/dev/net/tun", syscall.O_RDWR, 0)
	if err != nil {
		syscall.Close(queue)
		return "", -1, errors.New("error opening the /dev/net/tun char file: " + err.Error())
	}

	_, _, errNo := syscall.Syscall(syscall.SYS_IOCTL, uintptr(queue), uintptr(syscall.TUNSETIFF), uintptr(unsafe.Pointer(&req)))
	if errNo != 0 {
		syscall.Close(queue)
		return "", -1, errors.New("error setting the TUN device parameters: " + err.Error())
	}

	return string(req.Name[:strings.Index(string(req.Name[:]), "\000")]), queue, nil
}
