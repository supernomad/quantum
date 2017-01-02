// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package socket

import (
	"github.com/Supernomad/quantum/common"
)

// Mock socket struct to use for testing.
type Mock struct {
}

// Read which just returns the supplied buffer in the form of a *common.Payload.
func (mock *Mock) Read(buf []byte, queue int) (*common.Payload, bool) {
	return common.NewSockPayload(buf, len(buf)), true
}

// Write which is a noop.
func (mock *Mock) Write(payload *common.Payload, queue int) bool {
	return true
}

// Open which is a noop.
func (mock *Mock) Open() error {
	return nil
}

// Close which is a noop.
func (mock *Mock) Close() error {
	return nil
}

// GetFDs which is a noop.
func (mock *Mock) GetFDs() []int {
	return nil
}

func newMock(cfg *common.Config) *Mock {
	return &Mock{}
}
