// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package plugin

import (
	"math/rand"
	"testing"

	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/crypto"
)

var mapping *common.Mapping

func init() {
	mapping = &common.Mapping{
		SupportedPlugins: []string{"compression", "encryption"},
	}
	key := []byte("AES256Key-32Characters1234567890")
	salt := make([]byte, crypto.SaltLength)

	rand.Read(salt)

	aes, _ := crypto.NewAES(key, salt)
	mapping.AES = aes
}

func testEq(a, b []byte) bool {

	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func fillSlice(buf []byte) error {
	_, err := rand.Read(buf)
	return err
}

func TestEncryption(t *testing.T) {
	encryption, err := New(EncryptionPlugin, &common.Config{})
	if err != nil {
		t.Fatal("Failed to create new compression plugin.")
	}

	buf := make([]byte, common.MaxPacketLength)
	err = fillSlice(buf)

	out := common.NewTunPayload(buf, common.MTU)

	encrypted, _, ok := encryption.Apply(Outgoing, out, mapping)
	if !ok {
		t.Fatal("Failed to encrypt the outgoing payload.")
	}

	in := common.NewSockPayload(encrypted.Raw, encrypted.Length)

	_, _, ok = encryption.Apply(Incoming, in, mapping)
	if !ok {
		t.Fatal("Failed to decrypt the incoming payload.")
	}

	if !testEq(out.Packet, in.Packet) {
		t.Fatal("The outgoing and incoming payloads don't match after encryption/decryption.")
	}

	encryption.Close()
}

func TestCompression(t *testing.T) {
	compression, err := New(CompressionPlugin, &common.Config{})
	if err != nil {
		t.Fatal("Failed to create new compression plugin.")
	}

	buf := make([]byte, common.MaxPacketLength)
	err = fillSlice(buf)

	out := common.NewTunPayload(buf, common.MTU)

	compressed, _, ok := compression.Apply(Outgoing, out, mapping)
	if !ok {
		t.Fatal("Failed to compress the outgoing payload.")
	}

	in := common.NewSockPayload(compressed.Raw, compressed.Length)

	_, _, ok = compression.Apply(Incoming, in, mapping)
	if !ok {
		t.Fatal("Failed to decompress the incoming payload.")
	}

	if !testEq(out.Packet, in.Packet) {
		t.Fatal("The outgoing and incoming payloads don't match after compression/decompression.")
	}

	compression.Close()
}

func TestMock(t *testing.T) {
	mock, _ := New(MockPlugin, &common.Config{})

	if payload, mapping, ok := mock.Apply(Outgoing, nil, nil); !ok || payload != nil || mapping != nil {
		t.Fatal("Mock Apply should always return ok.")
	}

	if mock.Close() != nil {
		t.Fatal("Mock Close should always return nil.")
	}
}
