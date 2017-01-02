// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package device

import (
	"errors"

	"github.com/Supernomad/quantum/common"
)

// Type defines what kind of virual network device to use.
type Type int

const (
	// TUNDevice type creates and manages a TUN based network device.
	TUNDevice Type = iota

	// MOCKDevice type creates and manages a mocked out network device for testing.
	MOCKDevice
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
	Read(buf []byte, queue int) (*common.Payload, bool)

	// Write should handle being passed a formatted *common.Payload, and write the underlying raw data to the specified device queue.
	Write(payload *common.Payload, queue int) bool

	// Close should gracefully destroy the virtual network device.
	Close() error

	// Queues should return all underlying queue file descriptors to pass along during a rolling restart.
	Queues() []int
}

// New will generate a new Device struct based on the supplied device deviceType and user configuration
func New(deviceType Type, cfg *common.Config) (Device, error) {
	switch deviceType {
	case TUNDevice:
		return newTUN(cfg)
	case MOCKDevice:
		return newMock(cfg)
	}
	return nil, errors.New("build error device type undefined")
}
