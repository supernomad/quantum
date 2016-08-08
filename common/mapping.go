package common

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"github.com/Supernomad/quantum/ecdh"
	"net"
	"strconv"
	"strings"
	"syscall"
)

// Mapping represents the relationship between a public/private address and encryption metadata
type Mapping struct {
	Address   string
	MachineID string
	PrivateIP string
	PublicKey []byte
	Sockaddr  *syscall.SockaddrInet4 `json:"-"`
	SecretKey []byte                 `json:"-"`
	Cipher    cipher.AEAD            `json:"-"`
}

// String returns the mapping as a string
func (m *Mapping) String() string {
	return string(m.Bytes())
}

// Bytes returns the mapping as a byte slice
func (m *Mapping) Bytes() []byte {
	buf, _ := json.Marshal(m)
	return buf
}

// ParseMapping creates a new mapping based on the output of Mapping.Bytes
func ParseMapping(data []byte, privkey []byte) (*Mapping, error) {
	var mapping Mapping
	var addr [4]byte

	json.Unmarshal(data, &mapping)

	split := strings.Split(mapping.Address, ":")

	copy(addr[:], net.ParseIP(split[0]).To4())
	port, _ := strconv.Atoi(split[1])

	mapping.Sockaddr = &syscall.SockaddrInet4{
		Port: port,
		Addr: addr,
	}
	mapping.SecretKey = ecdh.GenerateSharedSecret(mapping.PublicKey, privkey)

	block, err := aes.NewCipher(mapping.SecretKey)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	mapping.Cipher = aesgcm
	return &mapping, nil
}

// NewMapping generates a new basic Mapping
func NewMapping(privateIP, address, machineID string, pubkey []byte) *Mapping {
	return &Mapping{
		Address:   address,
		MachineID: machineID,
		PrivateIP: privateIP,
		PublicKey: pubkey,
	}
}
