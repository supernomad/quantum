// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE
package socket

import (
	"github.com/Supernomad/quantum/common"
)

// Mock socket
type Mock struct {
}

// Read a packet off the socket
func (mock *Mock) Read(buf []byte, queue int) (*common.Payload, bool) {
	return common.NewSockPayload(buf, len(buf)), true
}

// Write a packet to the socket
func (mock *Mock) Write(payload *common.Payload, queue int) bool {
	return true
}

// Open the socket
func (mock *Mock) Open() error {
	return nil
}

// Close the socket
func (mock *Mock) Close() error {
	return nil
}

// GetFDs will return the underlying queue fds
func (mock *Mock) GetFDs() []int {
	return nil
}

func newMock(cfg *common.Config) *Mock {
	return &Mock{}
}
