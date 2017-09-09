// Copyright (c) 2016-2017 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package plugin

import (
	"errors"

	"github.com/supernomad/quantum/common"
)

const (
	// CompressionPlugin configures and injects a compression based plugin.
	CompressionPlugin = "compression"

	// EncryptionPlugin configures and injects an encryption based plugin.
	EncryptionPlugin = "encryption"

	// MockPlugin configures and injects a mock plugin for testing.
	MockPlugin = "mock"
)

const (
	// CompressionPluginOrder is the location of the compression plugin in the overall plugins enabled within quantum.
	CompressionPluginOrder = iota

	// EncryptionPluginOrder is the location of the encryption plugin in the overall plugins enabled within quantum.
	EncryptionPluginOrder

	// MockPluginOrder is the location of the mock plugin in the overall plugins enabled within quantum.
	MockPluginOrder
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

	// Name should return the name of the plugin.
	Name() string

	// Order returns the location of the specified plugin in the overall plugins enabled within quantum.
	Order() int
}

// Plugins is a collection of plugin structs for quantum to use.
type Plugins []Plugin

// Len returns the number of plugins.
func (plugins Plugins) Len() int {
	return len(plugins)
}

// Swap will swap the specified plugins in place.
func (plugins Plugins) Swap(i, j int) {
	plugins[i], plugins[j] = plugins[j], plugins[i]
}

// Sorter is used to sort the various enabled plugins within quantum.
type Sorter struct {
	Plugins
}

// Less returns whether or not the plugin at index 'i' is less than the that of the plugin at index 'j'.
func (sorter Sorter) Less(i, j int) bool {
	return sorter.Plugins[i].Order() < sorter.Plugins[j].Order()
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
