// Package common ip addressing func's mainly used for dhcp
// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE
package common

import (
	"errors"
	"net"
)

func getLocalMappingIfExists(machineID string, mappings map[uint32]*Mapping) (*Mapping, bool) {
	for _, mapping := range mappings {
		if mapping.MachineID == machineID {
			return mapping, true
		}
	}
	return nil, false
}

func ipExists(ip net.IP, mappings map[uint32]*Mapping) bool {
	for _, mapping := range mappings {
		if ArrayEquals(ip.To4(), mapping.PrivateIP.To4()) {
			return true
		}
	}
	return false
}

func getFreeIP(cfg *Config, mappings map[uint32]*Mapping) (net.IP, error) {
	for ip := cfg.NetworkConfig.BaseIP.Mask(cfg.NetworkConfig.IPNet.Mask); cfg.NetworkConfig.IPNet.Contains(ip); IncrementIP(ip) {
		if ip[3] == 0 || ip[3] == 255 ||
			(cfg.NetworkConfig.StaticNet != nil && cfg.NetworkConfig.StaticNet.Contains(ip)) ||
			ipExists(ip, mappings) {
			continue
		}
		return ip, nil
	}
	return nil, errors.New("there are no available ip addresses in the configured network")
}

// GenerateLocalMapping will take in the user defined configuration and the currently defined mappings to determine the local node mapping.
func GenerateLocalMapping(cfg *Config, mappings map[uint32]*Mapping) (*Mapping, error) {
	if cfg.PrivateIP == nil {
		if mapping, exists := getLocalMappingIfExists(cfg.MachineID, mappings); exists {
			cfg.PrivateIP = mapping.PrivateIP
		} else {
			ip, err := getFreeIP(cfg, mappings)
			if err != nil {
				return nil, err
			}
			cfg.PrivateIP = ip
		}
	} else if _, exists := getLocalMappingIfExists(cfg.MachineID, mappings); !exists && ipExists(cfg.PrivateIP, mappings) {
		return nil, errors.New("statically assigned private ip address belongs to another server")
	}

	return NewMapping(cfg.MachineID, cfg.PrivateIP, cfg.PublicIPv4, cfg.PublicIPv6, cfg.ListenPort, cfg.PublicKey), nil
}
