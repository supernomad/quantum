// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package plugin

import (
	"errors"

	"github.com/Supernomad/quantum/common"
)

const (
	// CompressionPlugin configures and injects a compression based plugin.
	CompressionPlugin = "compression"

	// EncryptionPlugin configures and injects an encryption based plugin.
	EncryptionPlugin = "encryption"

	// MockPlugin configures and injects a mock plugin for testing.
	MockPlugin = "mock"
)

// Direction of the packet when supplied to the plugin in question.
type Direction int

const (
	// Incoming a packet is coming from a remote node destined for the local node.
	Incoming = iota

	// Outgoing a packet is coming from the local node destined for a remote node.
	Outgoing
)

// Plugin interface for a generic multi-queue network device.
type Plugin interface {
	// Apply should apply the plugin to the specified payload and mapping.
	Apply(direction Direction, payload *common.Payload, mapping *common.Mapping) (*common.Payload, *common.Mapping, bool)

	// Close should gracefully destroy the plugin.
	Close() error
}

// New will generate a new Plugin struct based on the supplied device pluginType and user configuration.
func New(pluginType string, cfg *common.Config) (Plugin, error) {
	switch pluginType {
	case CompressionPlugin:
		return newCompression(cfg)
	case EncryptionPlugin:
		return newEncryption(cfg)
	case MockPlugin:
		return newMock(cfg)
	}
	return nil, errors.New("specified plugin is not supported")
}
