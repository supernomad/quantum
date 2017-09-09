// Copyright (c) 2016-2017 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package device

import (
	"errors"

	"github.com/supernomad/quantum/common"
)

const (
	// TUNDevice creates and manages a TUN based network device.
	TUNDevice = "tun"

	// MOCKDevice creates and manages a mocked out network device for testing.
	MOCKDevice = "mock"
)

const (
	ifNameSize    = 16
	iffTun        = 0x0001
	iffNoPi       = 0x1000
	iffMultiQueue = 0x0100
)

type ifReq struct {
	Name  [ifNameSize]byte
	Flags uint16
}

// Device interface for a generic multi-queue network device.
type Device interface {
	// Should return the name of the virual network device.
	Name() string

	// Read should return a formatted *common.Payload, based on the provided byte slice, off the specified device queue.
	Read(queue int, buf []byte) (*common.Payload, bool)

	// Write should handle being passed a formatted *common.Payload, and write the underlying raw data to the specified device queue.
	Write(queue int, payload *common.Payload) bool

	// Close should gracefully destroy the virtual network device.
	Close() error

	// Queues should return all underlying queue file descriptors to pass along during a rolling restart.
	Queues() []int
}

// New will generate a new Device struct based on the supplied device deviceType and user configuration
func New(deviceType string, cfg *common.Config) (Device, error) {
	switch deviceType {
	case TUNDevice:
		return newTUN(cfg)
	case MOCKDevice:
		return newMock(cfg)
	}
	return nil, errors.New("build error device type undefined")
}
