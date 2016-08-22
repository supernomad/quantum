package common

import (
	"encoding/json"
)

// Stats object for incoming/outgoing statistics
type Stats struct {
	DroppedPackets     uint64
	DroppedPacketsDiff uint64
	DroppedPPS         float64

	Packets     uint64
	PacketsDiff uint64
	PPS         float64

	DroppedBytes     uint64
	DroppedBytesDiff uint64
	DroppedBandwidth float64

	Bytes     uint64
	BytesDiff uint64
	Bandwidth float64

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
		Packets:   0,
		PPS:       0,
		Bytes:     0,
		Bandwidth: 0,
		Links:     make(map[string]*Stats),
	}
}
