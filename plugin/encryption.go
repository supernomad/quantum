// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package plugin

import (
	"github.com/Supernomad/quantum/common"
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
		err := mapping.AES.Decrypt(payload.Packet, payload.IPAddress)
		if err != nil {
			return payload, mapping, false
		}
		payload.Packet = payload.Packet[:mapping.AES.DecryptedSize(payload.Packet)]
	case Outgoing:
		payload.Length = mapping.AES.EncryptedSize(payload.Packet) + common.HeaderSize
		err := mapping.AES.Encrypt(payload.Raw[common.PacketStart:], len(payload.Packet), payload.IPAddress)
		if err != nil {
			return payload, mapping, false
		}
	}
	return payload, mapping, true
}

// Close which is a noop.
func (enc *Encryption) Close() error {
	return nil
}

func newEncryption(cfg *common.Config) (Plugin, error) {
	return &Encryption{
		cfg: cfg,
	}, nil
}
