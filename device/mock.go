// Copyright (c) 2016-2017 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package device

import (
	"github.com/supernomad/quantum/common"
)

// Mock device struct to use for testing.
type Mock struct {
}

// Name of the mock device.
func (mock *Mock) Name() string {
	return "Mocked Device"
}

// Read which just returns the supplied buffer in the form of a *common.Payload.
func (mock *Mock) Read(queue int, buf []byte) (*common.Payload, bool) {
	return common.NewTunPayload(buf, common.MTU), true
}

// Write which is a noop.
func (mock *Mock) Write(queue int, payload *common.Payload) bool {
	return true
}

// Close which is a noop.
func (mock *Mock) Close() error {
	return nil
}

// Queues which is a noop.
func (mock *Mock) Queues() []int {
	return nil
}

func newMock(cfg *common.Config) (Device, error) {
	return &Mock{}, nil
}
