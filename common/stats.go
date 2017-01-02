// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package common

import (
	"encoding/json"
)

// StatDirection determines which work the stat object came from either incoming or outgoing.
type StatDirection int

const (
	// IncomingStat i.e. RX stats
	IncomingStat StatDirection = iota // 0

	// OutgoingStat i.e. TX stats
	OutgoingStat // 1
)

// Stats object for monitoring incoming or outgoing statistics
type Stats struct {
	// The number of packets quantum has dropped due to failing either nating or cryptography.
	DroppedPackets uint64 `json:"droppedPackets"`

	// The number of packets successfully handled by quantum.
	Packets uint64 `json:"packets"`

	// The number of bytes quantum has dropped due to failing either nating or cryptography.
	DroppedBytes uint64 `json:"droppedBytes"`

	// The number of bytes successfully handled by quantum.
	Bytes uint64 `json:"bytes"`

	// The stats for individual links that represent the network traffic of this node in relation to remote nodes.
	Links map[string]*Stats `json:"links,omitempty"`

	// The stats for indivifual queues within quantm.
	Queues []*Stats `json:"queues,omitempty"`
}

// String returns a string representation of the Stats object, if there is an error while marshalling data an empty string is returned.
func (stats *Stats) String() string {
	data, _ := json.MarshalIndent(stats, "", "    ")
	return string(data)
}

// NewStats generates a new stats object to monitor quantum with.
func NewStats(numQueues int) *Stats {
	queues := make([]*Stats, numQueues)
	for i := 0; i < numQueues; i++ {
		queues[i] = &Stats{}
	}
	return &Stats{
		Links:  make(map[string]*Stats),
		Queues: queues,
	}
}

// StatsLog struct which contains the packet and byte statistics information for quantum.
type StatsLog struct {
	// TxStats holds the packet and byte counts for packet transmission.
	TxStats *Stats

	// RxStats holds the packet and byte counts for packet reception.
	RxStats *Stats
}

// Bytes returns a byte slice representation of the StatsLog object, if there is an error while marshalling data a nil slice is returned.
func (statsl *StatsLog) Bytes(pretty bool) []byte {
	var data []byte
	if pretty {
		data, _ = json.MarshalIndent(statsl, "", "    ")
	} else {
		data, _ = json.Marshal(statsl)
	}
	return data
}

// String returns a string representation of the StatsLog object, if there is an error while marshalling data an empty string is returned.
func (statsl *StatsLog) String(pretty bool) string {
	return string(statsl.Bytes(pretty))
}

// Stat is used to represent statistics about a single incoming or outgoing packet.
type Stat struct {
	// The remote private ip associated with the packet.
	PrivateIP string

	// The queue handling the packet.
	Queue int

	// The size of the packet in bytes.
	Bytes uint64

	// The direction of the packet, either incoming or outgoing.
	Direction StatDirection

	// Whether or not the packet was dropped.
	Dropped bool
}
