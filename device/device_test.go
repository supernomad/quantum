// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package device

import (
	"testing"

	"github.com/Supernomad/quantum/common"
)

func TestMock(t *testing.T) {
	mock, _ := New(MOCKDevice, &common.Config{})
	buf := make([]byte, common.MaxPacketLength)

	payload, ok := mock.Read(0, buf)
	if payload == nil || !ok {
		t.Fatal("Mock Read should always return a valid payload and nil error.")
	}

	if mock.Name() == "" {
		t.Fatal("Mock Name should return a non empty string.")
	}

	if !mock.Write(0, payload) {
		t.Fatal("Mock Write should always return true.")
	}

	if mock.Queues() != nil {
		t.Fatal("Mock Queues should always return nil.")
	}

	if mock.Close() != nil {
		t.Fatal("Mock Close should always return nil.")
	}
}
