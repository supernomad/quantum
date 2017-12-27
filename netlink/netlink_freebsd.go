// Copyright (c) 2016-2017 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package netlink

import (
	"errors"
	"net"
	"os/exec"
	"regexp"
)

const (
	routeCmdName    = "/sbin/route"
	ifconfigCmdName = "/sbin/ifconfig"

	routeCmdv4Args    = []string{"-n", "get", "-inet", "-host"}
	ifconfigCmdv4Args = []string{"inet"}
	routeCmdv6Args    = []string{"-n", "get", "-inet6", "-host"}
	ifconfigCmdv6Args = []string{"inet6"}
)

var (
	interfaceRE = regexp.MustCompile(`interface: (.+)`)
	addressRE   = regexp.MustCompile(`(inet|inet6)\s([0-9.:a-f]+)\s`)
)

// RouteGet returns a list of routes that match the provided destination IP address.
func RouteGet(dst net.IP) (Routes, error) {
	routeCmd := exec.Command(routeCmdName)
	isIpv6 := false

	switch {
	case dst.To4() != nil:
		routeCmd.Args = append(routeCmdv4Args, dst.String())
	case dst.To16() != nil:
		isIpv6 = true
		routeCmd.Args = append(routeCmdv6Args, dst.String())
	default:
		return nil, errors.New("invalid IP address supplied to RouteGet")
	}

	routeOutput, err := routeCmd.Output()
	if err != nil {
		return nil, errors.New("route command '" + routeCmdName + "' failed to return routes: " + err.Error())
	}

	routeResults := interfaceRE.FindStringSubmatch(routeOutput)
	if routeResults == nil {
		return nil, errors.New("route command did not return a valid interface for the provided route")
	}

	iname := routeResults[1]
	ifconfigCmd := exec.Command(ifconfigCmdName)

	if isIpv6 {
		ifconfigCmd.Args = append([]string{iname}, ifconfigCmdv6Args...)
	} else {
		ifconfigCmd.Args = append([]string{iname}, ifconfigCmdv4Args...)
	}

	ifconfigOutput, err := ifconfigCmd.Output()
	if err != nil {
		return nil, errors.New("ifconfig command '" + ifconfigCmdName + "' failed to return addresses: " + err.Error())
	}

	addressResults := addressRE.FindStringSubmatch(ifconfigOutput)

	if addressResults == nil {
		return nil, errors.New("ifconfig command did not return valid addresses for the provided route")
	}

	route := &Route{
		Src: net.ParseIP(addressResults[2]),
		Dst: nil,
	}

	return []*Route{route}, nil
}

// LinkSetup will configure and UP the specified link with the given configuration options.
func LinkSetup(name string, src net.IP, additionalIPs []net.IP, dst *net.IPNet, forward bool, mtu int) error {
	return nil
}
