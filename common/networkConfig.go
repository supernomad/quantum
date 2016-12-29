// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package common

import (
	"encoding/json"
	"errors"
	"net"
	"time"
)

// DefaultNetworkConfig to use when the NetworkConfig is not specified in the backend datastore.
var DefaultNetworkConfig *NetworkConfig

// NetworkConfig object to represent the current network.
type NetworkConfig struct {
	Network     string        `json:"network"`
	StaticRange string        `json:"staticRange"`
	LeaseTime   time.Duration `json:"leaseTime"`
	BaseIP      net.IP        `json:"-"`
	IPNet       *net.IPNet    `json:"-"`
	StaticNet   *net.IPNet    `json:"-"`
}

// ParseNetworkConfig from the return of the backend datastore
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

	if networkCfg.StaticRange == "" {
		return &networkCfg, nil
	}

	staticBase, staticNet, err := net.ParseCIDR(networkCfg.StaticRange)
	if err != nil {
		return nil, err
	} else if !ipnet.Contains(staticBase) {
		return nil, errors.New("network configuration has staticRange defined but the range does not exist in the configured network")
	}

	networkCfg.StaticNet = staticNet
	return &networkCfg, nil
}

// Bytes representation of a NetworkConfig object
func (networkCfg *NetworkConfig) Bytes() []byte {
	buf, _ := json.Marshal(networkCfg)
	return buf
}

// String representation of a NetworkConfig object
func (networkCfg *NetworkConfig) String() string {
	return string(networkCfg.Bytes())
}

func init() {
	defaultLeaseTime, _ := time.ParseDuration("48h")
	DefaultNetworkConfig = &NetworkConfig{
		Network:     "10.99.0.0/16",
		StaticRange: "10.99.0.0/23",
		LeaseTime:   defaultLeaseTime,
	}

	baseIP, ipnet, _ := net.ParseCIDR(DefaultNetworkConfig.Network)
	DefaultNetworkConfig.BaseIP = baseIP
	DefaultNetworkConfig.IPNet = ipnet

	_, staticNet, _ := net.ParseCIDR(DefaultNetworkConfig.StaticRange)
	DefaultNetworkConfig.StaticNet = staticNet
}
