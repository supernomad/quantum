// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package common

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"net"
	"syscall"
)

// Mapping represents the relationship between a public/private address and encryption metadata
type Mapping struct {
	MachineID     string                 `json:"machineID"`
	PrivateIP     net.IP                 `json:"privateIP"`
	PublicKey     []byte                 `json:"publicKey"`
	IPv4          net.IP                 `json:"ipv4,omitempty"`
	IPv6          net.IP                 `json:"ipv6,omitempty"`
	Port          int                    `json:"port"`
	Cipher        cipher.AEAD            `json:"-"`
	SockaddrInet4 *syscall.SockaddrInet4 `json:"-"`
	SockaddrInet6 *syscall.SockaddrInet6 `json:"-"`
}

// Bytes returns the mapping as a byte slice
func (mapping *Mapping) Bytes() []byte {
	buf, _ := json.Marshal(mapping)
	return buf
}

// String returns the mapping as a string
func (mapping *Mapping) String() string {
	return string(mapping.Bytes())
}

// ParseMapping creates a new mapping based on the output of Mapping.Bytes
func ParseMapping(str string, privkey []byte) (*Mapping, error) {
	data := []byte(str)
	var mapping Mapping
	json.Unmarshal(data, &mapping)

	secret := GenerateSharedSecret(mapping.PublicKey, privkey)

	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	mapping.Cipher = aesgcm

	if mapping.IPv4 != nil {
		sa := &syscall.SockaddrInet4{Port: mapping.Port}
		copy(sa.Addr[:], mapping.IPv4.To4())
		mapping.SockaddrInet4 = sa
	}
	if mapping.IPv6 != nil {
		sa := &syscall.SockaddrInet6{Port: mapping.Port}
		copy(sa.Addr[:], mapping.IPv6.To16())
		mapping.SockaddrInet6 = sa
	}

	return &mapping, nil
}

// NewMapping generates a new basic Mapping
func NewMapping(machineID string, privateIP, publicV4, publicV6 net.IP, port int, pubkey []byte) *Mapping {
	return &Mapping{
		MachineID: machineID,
		IPv4:      publicV4,
		IPv6:      publicV6,
		Port:      port,
		PrivateIP: privateIP,
		PublicKey: pubkey,
	}
}
