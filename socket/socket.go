// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package socket

import (
	"errors"
	"syscall"

	"github.com/Supernomad/quantum/common"
)

const (
	// UDPSocket type creates and manages a UDP based socket.
	UDPSocket = "udp"

	// DTLSSocket type creates and manages a UDP based socket that is encrypted using DTLS.
	DTLSSocket = "dtls"

	// MOCKSocket type creates and manages a mocked out socket for testing.
	MOCKSocket = "mock"
)

// Socket interface for a generic multi-queue socket interface.
type Socket interface {
	// Read should return a formatted *common.Payload, based on the provided byte slice, off the specified socket queue.
	Read(queue int, buf []byte) (*common.Payload, bool)

	// Write should handle being passed a formatted *common.Payload + *common.Mapping, and write the underlying raw data using the specified socket queue.
	Write(queue int, payload *common.Payload, mapping *common.Mapping) bool

	// Close should gracefully destroy the socket.
	Close() error

	// Queues should return all underlying queue file descriptors to pass along during a rolling restart.
	Queues() []int
}

// New generates a socket based on the supplied type and configuration.
func New(socketType string, cfg *common.Config) (Socket, error) {
	switch socketType {
	case UDPSocket:
		return newUDP(cfg)
	case DTLSSocket:
		return newDTLS(cfg)
	case MOCKSocket:
		return newMock(cfg)
	}
	return nil, errors.New("build error socket type undefined")
}

func createUDPSocket(ipv6Enabled bool, sa syscall.Sockaddr) (int, error) {
	// Grab the correct socket family.
	family := syscall.AF_INET
	if ipv6Enabled {
		family = syscall.AF_INET6
	}

	// Create the socket.
	fd, err := syscall.Socket(family, syscall.SOCK_DGRAM, 0)
	if err != nil {
		return -1, errors.New("error setting the UDP socket parameters: " + err.Error())
	}

	// Set the various socket options required.
	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		return -1, errors.New("error setting the UDP socket parameters: " + err.Error())
	}

	// Bind the newly created and configured socket.
	err = syscall.Bind(fd, sa)
	if err != nil {
		return -1, errors.New("error binding the UDP socket to the configured listen address: " + err.Error())
	}

	return fd, nil
}
