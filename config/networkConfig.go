package config

import (
	"encoding/json"
	"net"
	"time"
)

var DefaultNetworkConfig *NetworkConfig

type NetworkConfig struct {
	Network   string
	LeaseTime time.Duration

	BaseIP net.IP     `json:"-"`
	IPNet  *net.IPNet `json:"-"`
}

func ParseNetworkConfig(data []byte) (*NetworkConfig, error) {
	var networkCfg NetworkConfig
	json.Unmarshal(data, &networkCfg)

	baseIP, ipnet, err := net.ParseCIDR(networkCfg.Network)
	if err != nil {
		return nil, err
	}

	networkCfg.BaseIP = baseIP
	networkCfg.IPNet = ipnet
	return &networkCfg, nil
}

func (networkCfg *NetworkConfig) Bytes() []byte {
	buf, _ := json.Marshal(networkCfg)
	return buf
}

func init() {
	defaultLeaseTime, _ := time.ParseDuration("48h")
	DefaultNetworkConfig = &NetworkConfig{
		Network:   "10.9.0.0/16",
		LeaseTime: defaultLeaseTime,
	}
	baseIP, ipnet, _ := net.ParseCIDR(DefaultNetworkConfig.Network)
	DefaultNetworkConfig.BaseIP = baseIP
	DefaultNetworkConfig.IPNet = ipnet
}
