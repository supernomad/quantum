// Copyright (c) 2016-2018 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package plugin

import (
	"github.com/supernomad/quantum/common"
)

// Encryption plugin struct to use for encrypting outgoing packets or decrypting incoming packets.
type Encryption struct {
	cfg *common.Config
}

// Apply returns the payload/mapping encrypted if the direction is Outgoing and decrypted if the direction is Incoming.
func (enc *Encryption) Apply(direction Direction, payload *common.Payload, mapping *common.Mapping) (*common.Payload, *common.Mapping, bool) {
	if !common.StringInSlice(EncryptionPlugin, mapping.SupportedPlugins) {
		return payload, mapping, true
	}

	switch direction {
	case Incoming:
		length, err := mapping.AES.Decrypt(payload.Packet, payload.IPAddress)
		if err != nil {
			return payload, mapping, false
		}

		payload.Packet = payload.Raw[common.PacketStart : common.PacketStart+length]
		payload.Length = common.HeaderSize + length
	case Outgoing:
		length, err := mapping.AES.Encrypt(payload.Raw[common.PacketStart:], len(payload.Packet), payload.IPAddress)
		if err != nil {
			return payload, mapping, false
		}

		payload.Packet = payload.Raw[common.PacketStart : common.PacketStart+length]
		payload.Length = common.HeaderSize + length
	}
	return payload, mapping, true
}

// Close which is a noop.
func (enc *Encryption) Close() error {
	return nil
}

// Name returns 'encryption'.
func (enc *Encryption) Name() string {
	return EncryptionPlugin
}

// Order returns the EncryptionPluginOrder value.
func (enc *Encryption) Order() int {
	return EncryptionPluginOrder
}

func newEncryption(cfg *common.Config) (Plugin, error) {
	return &Encryption{
		cfg: cfg,
	}, nil
}
