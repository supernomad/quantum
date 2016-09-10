package common

import (
	"encoding/binary"
	"net"
	"strconv"
)

const (
	// MTU - The max size packet to recieve from the TUN device
	MTU = 65475

	// HeaderSize - The size of the perpended data
	HeaderSize = 16

	// FooterSize - The size of the appended data
	FooterSize = 16

	// MaxPacketLength - The maximum packet size to send via the UDP device
	MaxPacketLength = HeaderSize + MTU + FooterSize
)

const (
	// IPStart - The ip start position
	IPStart = 0

	// IPEnd - The ip end position
	IPEnd = 4

	// NonceStart - The nonce start position
	NonceStart = 4

	// NonceEnd - The nonce end postion
	NonceEnd = 16

	// PacketStart - The packet start position
	PacketStart = 16
)

// IPtoInt takes a string ip in the form '0.0.0.0' and returns a uint32 that represents that ipaddress
func IPtoInt(IP string) uint32 {
	buf := net.ParseIP(IP).To4()
	return binary.LittleEndian.Uint32(buf)
}

// IncrementIP will increment the given ip in place.
func IncrementIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] > 0 {
			break
		}
	}
}

// ToStringArray taks in a slice of ints and returns a slice of strings
func ToStringArray(ints []int) []string {
	strs := make([]string, len(ints))
	for i := 0; i < len(ints); i++ {
		strs[i] = strconv.Itoa(ints[i])
	}
	return strs
}

// ToIntArray takes in a slice of strings and returns a slice of ints or an error
func ToIntArray(strs []string) ([]int, error) {
	ints := make([]int, len(strs))
	for i := 0; i < len(strs); i++ {
		j, err := strconv.Atoi(strs[i])
		if err != nil {
			return nil, err
		}
		ints[i] = j
	}
	return ints, nil
}
