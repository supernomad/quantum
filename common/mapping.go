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

// Mapping represents the relationship between a public/private address along with encryption metadata for a particular node in the quantum network.
type Mapping struct {
	// The unique machine id within the quantum network.
	MachineID string `json:"machineID"`

	// The private ip address within the quantum network.
	PrivateIP net.IP `json:"privateIP"`

	// The public key to be used during encryption/decryption of packets.
	PublicKey []byte `json:"publicKey"`

	// The public ipv4 address of the node represented by this mapping, which may or may not exist.
	IPv4 net.IP `json:"ipv4,omitempty"`

	// The public ipv6 address of the node represented by this mapping, which may or may not exist.
	IPv6 net.IP `json:"ipv6,omitempty"`

	// The port where quantum is listening for remote packets.
	Port int `json:"port"`

	// The AEAD object to use for encryption/decryption of packets
	Cipher cipher.AEAD `json:"-"`

	// The ipv4 syscall.Sockaddr object for sending data to the node represented by this mapping.
	SockaddrInet4 *syscall.SockaddrInet4 `json:"-"`

	// The ipv6 syscall.Sockaddr object for sending data to the node represented by this mapping.
	SockaddrInet6 *syscall.SockaddrInet6 `json:"-"`
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

// NewMapping generates a new basic Mapping with no cryptographic metadata.
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
