// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package plugin

import (
	"math/rand"
	"sort"
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

func TestSorter(t *testing.T) {
	encryption, err := New(EncryptionPlugin, &common.Config{})
	if err != nil {
		t.Fatal("Failed to create new encryption plugin.")
	}

	compression, err := New(CompressionPlugin, &common.Config{})
	if err != nil {
		t.Fatal("Failed to create new compression plugin.")
	}

	mock, err := New(MockPlugin, &common.Config{})
	if err != nil {
		t.Fatal("Failed to create new mock plugin.")
	}

	plugins := []Plugin{mock, encryption, compression}

	sort.Sort(Sorter{Plugins: plugins})

	if plugins[0].Name() != CompressionPlugin || plugins[1].Name() != EncryptionPlugin || plugins[2].Name() != MockPlugin {
		t.Fatal("Failed to properly sort the plugins")
	}

	sort.Sort(sort.Reverse(Sorter{Plugins: plugins}))

	if plugins[0].Name() != MockPlugin || plugins[1].Name() != EncryptionPlugin || plugins[2].Name() != CompressionPlugin {
		t.Fatal("Failed to properly reverse the plugins")
	}
}

func TestEncryption(t *testing.T) {
	encryption, err := New(EncryptionPlugin, &common.Config{})
	if err != nil {
		t.Fatal("Failed to create new compression plugin.")
	}

	buf := make([]byte, common.MaxPacketLength)
	expected := make([]byte, common.MaxPacketLength)

	err = fillSlice(buf)
	if err != nil {
		t.Fatal("Failed to fill buffer for encryption.")
	}

	copy(expected, buf)

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

	if !testEq(expected[:common.MTU], buf[:common.MTU]) {
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
	expected := make([]byte, common.MaxPacketLength)

	err = fillSlice(buf)
	if err != nil {
		t.Fatal("Failed to fill buffer for encryption.")
	}

	copy(expected, buf)

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

	if !testEq(expected[:common.MTU], buf[:common.MTU]) {
		t.Fatal("The outgoing and incoming payloads don't match after compression/decompression.")
	}

	compression.Close()
}

func TestMulti(t *testing.T) {
	encryption, err := New(EncryptionPlugin, &common.Config{})
	if err != nil {
		t.Fatal("Failed to create new encryption plugin.")
	}

	compression, err := New(CompressionPlugin, &common.Config{})
	if err != nil {
		t.Fatal("Failed to create new compression plugin.")
	}

	plugins := []Plugin{encryption, compression}

	sort.Sort(Sorter{Plugins: plugins})

	buf := make([]byte, common.MaxPacketLength)
	expected := make([]byte, common.MaxPacketLength)

	err = fillSlice(buf)
	if err != nil {
		t.Fatal("Failed to fill buffer for encryption.")
	}

	copy(expected, buf)

	payload := common.NewTunPayload(buf, common.MTU)

	var ok bool
	for i := 0; i < len(plugins); i++ {
		payload, mapping, ok = plugins[i].Apply(Outgoing, payload, mapping)
		if !ok {
			t.Fatalf("Failed to apply outgoing plugin: %s", plugins[i].Name())
		}
	}

	sort.Sort(sort.Reverse(Sorter{Plugins: plugins}))

	payload = common.NewSockPayload(payload.Raw, payload.Length)

	for i := 0; i < len(plugins); i++ {
		payload, mapping, ok = plugins[i].Apply(Incoming, payload, mapping)
		if !ok {
			t.Fatalf("Failed to apply incoming plugin: %s", plugins[i].Name())
		}
	}

	if payload.Length-common.HeaderSize != common.MTU {
		t.Fatal("The outgoing and incoming payloads have different lengths after applying all plugins.")
	}

	if !testEq(expected[:common.MTU], payload.Raw[:payload.Length-common.HeaderSize]) {
		t.Fatal("The outgoing and incoming payloads don't match after applying all plugins.")
	}
}

func TestMock(t *testing.T) {
	mock, _ := New(MockPlugin, &common.Config{})

	if payload, mapping, ok := mock.Apply(Outgoing, nil, nil); !ok || payload != nil || mapping != nil {
		t.Fatal("Mock Apply should always return ok.")
	}

	if mock.Close() != nil {
		t.Fatal("Mock Close should always return nil.")
	}

	if mock.Order() != MockPluginOrder {
		t.Fatal("Mock Order should always return MockPluginOrder.")
	}
}
