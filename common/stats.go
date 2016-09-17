package common

import (
	"encoding/json"
)

// Stats object for incoming/outgoing statistics
type Stats struct {
	DroppedPackets uint64
	Packets        uint64

	DroppedBytes uint64
	Bytes        uint64

	Links map[string]*Stats `json:",omitempty"`
}

// String the Stats object
func (stats *Stats) String() string {
	data, _ := json.MarshalIndent(stats, "", "    ")
	return string(data)
}

// NewStats object with links
func NewStats() *Stats {
	return &Stats{
		Links: make(map[string]*Stats),
	}
}
