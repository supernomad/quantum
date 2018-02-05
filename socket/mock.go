// Copyright (c) 2016-2018 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package socket

import (
	"github.com/supernomad/quantum/common"
)

// Mock socket struct to use for testing.
type Mock struct {
}

// Read which just returns the supplied buffer in the form of a *common.Payload.
func (mock *Mock) Read(queue int, buf []byte) (*common.Payload, bool) {
	return common.NewSockPayload(buf, len(buf)), true
}

// Write which is a noop.
func (mock *Mock) Write(queue int, payload *common.Payload, mapping *common.Mapping) bool {
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
