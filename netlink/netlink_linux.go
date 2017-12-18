// Copyright (c) 2016-2017 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package netlink

import (
	"errors"
	"net"

	"github.com/vishvananda/netlink"
)

// RouteGet returns a list of routes that match the provided destination IP address.
func RouteGet(dst net.IP) (Routes, error) {
	tmpRoutes, err := netlink.RouteGet(dst)
	if err != nil {
		return nil, err
	}

	routes := make([]*Route, len(tmpRoutes))
	for i := 0; i < len(tmpRoutes); i++ {
		route := &Route{
			Src: tmpRoutes[i].Src,
			Dst: tmpRoutes[i].Dst,
		}

		routes[i] = route
	}

	return routes, nil
}

// LinkSetup will configure and UP the specified link with the given configuration options.
func LinkSetup(name string, src net.IP, additionalIPs []net.IP, dst *net.IPNet, forward bool, mtu int) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return errors.New("error getting the virtual network device from the kernel: " + err.Error())
	}
	err = netlink.LinkSetUp(link)
	if err != nil {
		return errors.New("error upping the virtual network device: " + err.Error())
	}
	err = netlink.LinkSetMTU(link, mtu)
	if err != nil {
		return errors.New("error setting the virtual network device MTU: " + err.Error())
	}
	addr, err := netlink.ParseAddr(src.String() + "/32")
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
		Src:       src,
		Dst:       dst,
	}
	err = netlink.RouteAdd(route)
	if err != nil {
		return errors.New("error setting the virtual network device network routes: " + err.Error())
	}

	if forward {
		routes, _ := netlink.RouteList(nil, netlink.FAMILY_V4)
		for _, r := range routes {
			if r.Dst == nil {
				if err := netlink.RouteDel(&r); err != nil {
					return errors.New("error removing old default route: " + err.Error())
				}
			}
		}
		route := &netlink.Route{
			LinkIndex: link.Attrs().Index,
			Src:       src,
			Dst:       nil,
		}
		err = netlink.RouteAdd(route)
		if err != nil {
			return errors.New("error setting the virtual network device network routes: " + err.Error())
		}
	}

	for i := 0; i < len(additionalIPs); i++ {
		additional, err := netlink.ParseAddr(additionalIPs[i].String() + "/32")
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
