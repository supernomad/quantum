// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package datastore

import (
	"testing"

	"github.com/Supernomad/quantum/common"
)

func TestMock(t *testing.T) {
	mock, _ := New(MOCKDatastore, common.NewLogger(common.NoopLogger), &common.Config{})

	if mock.Init() != nil {
		t.Fatal("Mock Init should always return a nil error.")
	}

	mock.Start()

	mapping, ok := mock.Mapping(0)
	if !ok || mapping != nil {
		t.Fatal("Mock Mapping didn't return ok.")
	}

	mock.Stop()
}
