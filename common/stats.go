package common

import (
	"encoding/json"
)

// Stats object for incoming/outgoing statistics
type Stats struct {
	DroppedPackets uint64            `json:"droppedPackets"`
	Packets        uint64            `json:"packets"`
	DroppedBytes   uint64            `json:"droppedBytes"`
	Bytes          uint64            `json:"bytes"`
	Links          map[string]*Stats `json:"links,omitempty"`
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
