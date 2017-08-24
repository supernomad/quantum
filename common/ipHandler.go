// Copyright (c) 2016-2017 Christian Saide <Supernomad>
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

func nonFloatingIPExists(ip net.IP, mappings map[uint32]*Mapping) bool {
	for _, mapping := range mappings {
		if ArrayEquals(ip.To4(), mapping.PrivateIP.To4()) && !mapping.Floating {
			return true
		}
	}
	return false
}

func getFreeIP(cfg *Config, mappings map[uint32]*Mapping) (net.IP, error) {
	for ip := cfg.NetworkConfig.BaseIP.Mask(cfg.NetworkConfig.IPNet.Mask); cfg.NetworkConfig.IPNet.Contains(ip); IncrementIP(ip) {
		if ip[3] == 0 || ip[3] == 255 ||
			(cfg.NetworkConfig.StaticNet != nil && cfg.NetworkConfig.StaticNet.Contains(ip)) ||
			(cfg.NetworkConfig.FloatingNet != nil && cfg.NetworkConfig.FloatingNet.Contains(ip)) ||
			ipExists(ip, mappings) {
			continue
		}
		return ip, nil
	}
	return nil, errors.New("there are no available ip addresses in the configured network")
}

// GenerateLocalMapping will take in the user defined configuration plus the currently defined mappings, in order to determine the local mapping.
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
	} else if !cfg.NetworkConfig.IPNet.Contains(cfg.PrivateIP) {
		return nil, errors.New("statically assigned private ip address does not lie within the overall network range")
	}

	return NewMapping(cfg), nil
}

// GenerateFloatingMapping will take in the user defined configuration plus the currently defined mappins, in order to determine the floating mapping.
func GenerateFloatingMapping(cfg *Config, i int, mappings map[uint32]*Mapping) (*Mapping, error) {
	if nonFloatingIPExists(cfg.FloatingIPs[i], mappings) {
		return nil, errors.New("the floating ip '" + cfg.FloatingIPs[i].String() + "' is already assigned to a different node as a static or dhcp ip address")
	} else if cfg.NetworkConfig.StaticNet != nil && cfg.NetworkConfig.StaticNet.Contains(cfg.FloatingIPs[i]) {
		return nil, errors.New("the floating ip '" + cfg.FloatingIPs[i].String() + "' lies within the reserved static ip range")
	} else if cfg.NetworkConfig.FloatingNet != nil && !cfg.NetworkConfig.FloatingNet.Contains(cfg.FloatingIPs[i]) {
		return nil, errors.New("the floating ip '" + cfg.FloatingIPs[i].String() + "' does not lie within the reserved floating ip range")
	} else if !cfg.NetworkConfig.IPNet.Contains(cfg.FloatingIPs[i]) {
		return nil, errors.New("the floating ip '" + cfg.FloatingIPs[i].String() + "' does not lie within the overall network range")
	}

	return NewFloatingMapping(cfg, i), nil
}
