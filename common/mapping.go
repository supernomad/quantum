// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package common

import (
	"encoding/json"
	"errors"
	"net"
	"syscall"
)

// Mapping represents the relationship between a public/private address along with encryption metadata for a particular node in the quantum network.
type Mapping struct {
	// The unique machine id within the quantum network.
	MachineID string `json:"machineID"`

	// The private ip address within the quantum network.
	PrivateIP net.IP `json:"privateIP"`

	// The public ipv4 address of the node represented by this mapping, which may or may not exist.
	IPv4 net.IP `json:"ipv4,omitempty"`

	// The public ipv6 address of the node represented by this mapping, which may or may not exist.
	IPv6 net.IP `json:"ipv6,omitempty"`

	// The plugins that the node represented by this mapping supports.
	SupportedPlugins []string `json:"plugins,omitempty"`

	// The port where quantum is listening for remote packets.
	Port int `json:"port"`

	// The resulting endpoint to send data to the node represented by this mapping.
	Address string `json:"-"`

	// The resulting endpoint to send data to the node represented by this mapping.
	Sockaddr syscall.Sockaddr `json:"-"`
}

// Bytes returns a byte slice representation of a Mapping object, if there is an error while marshalling data a nil slice is returned.
func (mapping *Mapping) Bytes() []byte {
	buf, _ := json.Marshal(mapping)
	return buf
}

// Bytes returns a string representation of a Mapping object, if there is an error while marshalling data an empty string is returned.
func (mapping *Mapping) String() string {
	return string(mapping.Bytes())
}

// ParseMapping creates a new mapping based on the output of a Mapping.Bytes call.
func ParseMapping(str string, cfg *Config) (*Mapping, error) {
	data := []byte(str)
	var mapping Mapping
	json.Unmarshal(data, &mapping)

	if cfg.IsIPv6Enabled && mapping.IPv6 != nil {
		sa := &syscall.SockaddrInet6{Port: mapping.Port}
		copy(sa.Addr[:], mapping.IPv6.To16())

		mapping.Sockaddr = sa
		mapping.Address = mapping.IPv6.String()
	} else if cfg.IsIPv4Enabled && mapping.IPv4 != nil {
		sa := &syscall.SockaddrInet4{Port: mapping.Port}
		copy(sa.Addr[:], mapping.IPv4.To4())

		mapping.Sockaddr = sa
		mapping.Address = mapping.IPv4.String()
	} else {
		return nil, errors.New("mapping not compatible with this node due to networking conflicts: " + mapping.String())
	}

	return &mapping, nil
}

// NewMapping generates a new basic Mapping with no cryptographic metadata.
func NewMapping(cfg *Config) *Mapping {
	return &Mapping{
		MachineID:        cfg.MachineID,
		IPv4:             cfg.PublicIPv4,
		IPv6:             cfg.PublicIPv6,
		Port:             cfg.ListenPort,
		PrivateIP:        cfg.PrivateIP,
		SupportedPlugins: cfg.Plugins,
	}
}
