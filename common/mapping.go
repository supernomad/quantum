package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/json"
	"github.com/Supernomad/quantum/ecdh"
	"net"
	"strconv"
	"strings"
	"syscall"
)

type Mapping struct {
	Address      string
	PublicKey    []byte
	VerifyKeyBuf []byte
	VerifyKey    *ecdsa.PublicKey       `json:"-"`
	Sockaddr     *syscall.SockaddrInet4 `json:"-"`
	SecretKey    []byte                 `json:"-"`
	Cipher       cipher.AEAD            `json:"-"`
}

func (m *Mapping) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

func ParseMapping(data string, privkey []byte) (*Mapping, error) {
	var mapping Mapping
	var addr [4]byte
	json.Unmarshal([]byte(data), &mapping)

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
	x, y := elliptic.Unmarshal(elliptic.P256(), mapping.VerifyKeyBuf)
	mapping.VerifyKey = &ecdsa.PublicKey{
		X: x,
		Y: y,
	}
	mapping.VerifyKey.Curve = elliptic.P256()
	return &mapping, nil
}

func NewMapping(address string, pubkey []byte, verifykey *ecdsa.PublicKey) *Mapping {
	return &Mapping{
		Address:      address,
		PublicKey:    pubkey,
		VerifyKeyBuf: elliptic.Marshal(elliptic.P256(), verifykey.X, verifykey.Y),
		VerifyKey:    verifykey,
	}
}
