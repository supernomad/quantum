// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE
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
	Queues         []*Stats          `json:"queues,omitempty"`
}

// String the Stats object
func (stats *Stats) String() string {
	data, _ := json.MarshalIndent(stats, "", "    ")
	return string(data)
}

// NewStats object with links
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
