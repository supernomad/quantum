// Copyright (c) 2016-2018 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package device

import (
	"errors"
	"strings"
	"syscall"
	"unsafe"

	"github.com/supernomad/quantum/common"
	"github.com/vishvananda/netlink"
)

// Tun device struct for managing a multi-queue TUN networking device.
type Tun struct {
	name            string
	queues          []int
	oldDefaultRoute *netlink.Route
	cfg             *common.Config
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

	if tun.cfg.Forward {
		if err := netlink.RouteReplace(tun.oldDefaultRoute); err != nil {
			return errors.New("error adding old default route: " + err.Error())
		}
	}
	return nil
}

// Queues returns the underlying device queue file descriptors.
func (tun *Tun) Queues() []int {
	return tun.queues
}

// Read a packet off the specified device queue and return a *common.Payload representation of the packet.
func (tun *Tun) Read(queue int, buf []byte) (*common.Payload, bool) {
	n, err := syscall.Read(tun.queues[queue], buf[common.PacketStart:])
	if err != nil {
		return nil, false
	}
	return common.NewTunPayload(buf, n), true
}

// Write a *common.Payload to the specified device queue.
func (tun *Tun) Write(queue int, payload *common.Payload) bool {
	_, err := syscall.Write(tun.queues[queue], payload.Packet)
	return err == nil
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
		err := tun.initTun()
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
		return "", -1, errors.New("error setting the TUN device parameters")
	}

	return string(req.Name[:strings.Index(string(req.Name[:]), "\000")]), queue, nil
}

func (tun *Tun) initTun() error {
	link, err := netlink.LinkByName(tun.name)
	if err != nil {
		return errors.New("error getting the virtual network device from the kernel: " + err.Error())
	}
	err = netlink.LinkSetUp(link)
	if err != nil {
		return errors.New("error upping the virtual network device: " + err.Error())
	}
	err = netlink.LinkSetMTU(link, common.MTU)
	if err != nil {
		return errors.New("error setting the virtual network device MTU: " + err.Error())
	}
	addr, err := netlink.ParseAddr(tun.cfg.PrivateIP.String() + "/32")
	if err != nil {
		return errors.New("error parsing the virtual network device address: " + err.Error())
	}
	err = netlink.AddrAdd(link, addr)
	if err != nil {
		return errors.New("error setting the virtual network device address: " + err.Error())
	}
	route := &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Scope:     netlink.SCOPE_LINK,
		Protocol:  2,
		Src:       tun.cfg.PrivateIP,
		Dst:       tun.cfg.NetworkConfig.IPNet,
	}
	err = netlink.RouteAdd(route)
	if err != nil {
		return errors.New("error setting the virtual network device network routes: " + err.Error())
	}

	if tun.cfg.Forward {
		routes, _ := netlink.RouteList(nil, netlink.FAMILY_V4)
		for _, r := range routes {
			if r.Dst == nil {
				tun.oldDefaultRoute = &netlink.Route{
					LinkIndex: r.LinkIndex,
					Dst:       r.Dst,
					Src:       r.Src,
					Gw:        r.Gw,
				}
				if err := netlink.RouteDel(&r); err != nil {
					return errors.New("error removing old default route: " + err.Error())
				}
			}
		}
		route := &netlink.Route{
			LinkIndex: link.Attrs().Index,
			Src:       tun.cfg.PrivateIP,
			Dst:       nil,
		}
		err = netlink.RouteAdd(route)
		if err != nil {
			return errors.New("error setting the virtual network device network routes: " + err.Error())
		}
	}

	for i := 0; i < len(tun.cfg.FloatingIPs); i++ {
		additional, err := netlink.ParseAddr(tun.cfg.FloatingIPs[i].String() + "/32")
		if err != nil {
			return errors.New("error parsing the virtual network device address: " + err.Error())
		}
		err = netlink.AddrAdd(link, additional)
		if err != nil {
			return errors.New("error setting the virtual network device address: " + err.Error())
		}
	}

	return nil
}
