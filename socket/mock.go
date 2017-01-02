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

// Close which is a noop.
func (mock *Mock) Close() error {
	return nil
}

// Queues which is a noop.
func (mock *Mock) Queues() []int {
	return nil
}

func newMock(cfg *common.Config) (Socket, error) {
	return &Mock{}, nil
}
