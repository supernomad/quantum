// Copyright (c) 2016-2017 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package common

import (
	"encoding/json"
	"errors"
	"net"
	"time"
)

// NetworkConfig object to represent the current network setup.
type NetworkConfig struct {
	// The backend to use for communication.
	Backend string `json:"backend"`

	// The network range that represents the quantum network.
	Network string `json:"network"`

	// The reserved static ip address range which should be skipped for floating and DHCP assignments.
	StaticRange string `json:"staticRange"`

	// The reserved floating ip address range which should be skipped for static and DHCP assignments.
	FloatingRange string `json:"floatingRange"`

	// The length of time to hold the assigned DHCP lease.
	LeaseTime time.Duration `json:"leaseTime"`

	// The base ip address of the quantum network.
	BaseIP net.IP `json:"-"`

	// The IPNet representation of the quantum network.
	IPNet *net.IPNet `json:"-"`

	// The IPNet representation of the reserved static ip address range.
	StaticNet *net.IPNet `json:"-"`

	// The IPNet representation of the reserved floating ip address range.
	FloatingNet *net.IPNet `json:"-"`
}

// ParseNetworkConfig from the data stored in the datastore.
func ParseNetworkConfig(data []byte) (*NetworkConfig, error) {
	var networkCfg NetworkConfig
	json.Unmarshal(data, &networkCfg)

	if networkCfg.LeaseTime == 0 {
		networkCfg.LeaseTime = 48 * time.Hour
	}

	baseIP, ipnet, err := net.ParseCIDR(networkCfg.Network)
	if err != nil {
		return nil, err
	}

	networkCfg.BaseIP = baseIP
	networkCfg.IPNet = ipnet

	if networkCfg.StaticRange != "" {
		staticBase, staticNet, err := net.ParseCIDR(networkCfg.StaticRange)
		if err != nil {
			return nil, err
		} else if !ipnet.Contains(staticBase) {
			return nil, errors.New("network configuration has staticRange defined but the range does not exist in the configured network")
		}

		networkCfg.StaticNet = staticNet
	}

	if networkCfg.FloatingRange != "" {
		floatingBase, floatingNet, err := net.ParseCIDR(networkCfg.FloatingRange)
		if err != nil {
			return nil, err
		} else if !ipnet.Contains(floatingBase) {
			return nil, errors.New("network configuration has floatingRange defined but the range does not exist in the configured network")
		} else if networkCfg.StaticNet != nil && networkCfg.StaticNet.Contains(floatingBase) {
			return nil, errors.New("network configuration has floatingRange and staticRange defined but the ranges conflict with each other")
		}

		networkCfg.FloatingNet = floatingNet
	}

	return &networkCfg, nil
}

// Bytes returns a byte slice representation of a NetworkConfig object, if there is an error while marshalling data a nil slice is returned.
func (networkCfg *NetworkConfig) Bytes() []byte {
	buf, _ := json.Marshal(networkCfg)
	return buf
}

// Bytes returns a string representation of a NetworkConfig object, if there is an error while marshalling data an empty string is returned.
func (networkCfg *NetworkConfig) String() string {
	return string(networkCfg.Bytes())
}
