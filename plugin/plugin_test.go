// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package plugin

import (
	"math/rand"
	"testing"

	"github.com/Supernomad/quantum/common"
)

var mapping = &common.Mapping{
	SupportedPlugins: []string{"comp"},
}

func fillSlice(buf []byte) error {
	_, err := rand.Read(buf)
	return err
}

func TestCompression(t *testing.T) {
	comp, err := New(CompressionPlugin, &common.Config{})
	if err != nil {
		t.Fatal("Failed to create new compression plugin.")
	}

	buf := make([]byte, common.MaxPacketLength)
	err = fillSlice(buf)

	out := common.NewTunPayload(buf, common.MTU)

	t.Logf("Uncompressed length %dB", out.Length)
	compressed, _, ok := comp.Apply(Outgoing, out, mapping)
	if !ok {
		t.Fatal("Failed to compress the outgoing payload.")
	}
	t.Logf("Compressed length %dB", compressed.Length)

	in := common.NewSockPayload(compressed.Raw, compressed.Length)

	t.Logf("Compressed length %dB", in.Length)
	uncompressed, _, ok := comp.Apply(Incoming, in, mapping)
	if !ok {
		t.Fatalf("Failed to decompress the incoming payload. %s", string(uncompressed.Raw))
	}
	t.Logf("Uncompressed length %dB", uncompressed.Length)

	comp.Close()
}

func TestMock(t *testing.T) {
	mock, _ := New(MockPlugin, &common.Config{})

	if mock.Name() != "mock" {
		t.Fatal("Mock Name should always return 'mock'.")
	}

	if payload, mapping, ok := mock.Apply(Outgoing, nil, nil); !ok || payload != nil || mapping != nil {
		t.Fatal("Mock Apply should always return ok.")
	}

	if mock.Close() != nil {
		t.Fatal("Mock Close should always return nil.")
	}
}
