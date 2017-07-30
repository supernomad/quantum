// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package metric

import (
	"encoding/json"
)

const (
	// Rx metric
	Rx = iota

	// Tx metric
	Tx
)

// Metric is used to represent a single incoming or outgoing packet's metric.
type Metric struct {
	// The remote private ip associated with the packet.
	PrivateIP string

	// The queue handling the packet.
	Queue int

	// The size of the packet in bytes.
	Bytes uint64

	// The type of the packet, either Rx or Tx.
	Type int

	// Whether or not the packet was dropped.
	Dropped bool
}

// Metrics struct for storing aggregated incoming or outgoing statistics.
type Metrics struct {
	// The number of packets quantum has dropped.
	DroppedPackets uint64 `json:"droppedPackets"`

	// The number of packets successfully handled by quantum.
	Packets uint64 `json:"packets"`

	// The number of bytes quantum has dropped.
	DroppedBytes uint64 `json:"droppedBytes"`

	// The number of bytes successfully handled by quantum.
	Bytes uint64 `json:"bytes"`

	// The stats for individual links that represent the network traffic of this node in relation to remote nodes.
	Links map[string]*Metrics `json:"links,omitempty"`

	// The stats for individual queues within quantm.
	Queues []*Metrics `json:"queues,omitempty"`
}

// MetricsLog struct which contains the packet and byte statistics information for quantum.
type MetricsLog struct {
	// TxMetrics holds the packet and byte counts for packet transmission.
	TxMetrics *Metrics

	// RxMetrics holds the packet and byte counts for packet reception.
	RxMetrics *Metrics
}

// Bytes returns a byte slice json representation of the MetricsLog struct in either flat or prettified notation, if there is an error while marshalling data a nil slice is returned.
func (metricsLog *MetricsLog) Bytes(pretty bool) []byte {
	var data []byte
	if pretty {
		data, _ = json.MarshalIndent(metricsLog, "", "    ")
	} else {
		data, _ = json.Marshal(metricsLog)
	}
	return data
}

func newMetricsLog(numWorkers int) *MetricsLog {
	metricsLog := &MetricsLog{
		TxMetrics: &Metrics{
			Links:  make(map[string]*Metrics),
			Queues: make([]*Metrics, numWorkers),
		},
		RxMetrics: &Metrics{
			Links:  make(map[string]*Metrics),
			Queues: make([]*Metrics, numWorkers),
		},
	}

	for i := 0; i < numWorkers; i++ {
		metricsLog.TxMetrics.Queues[i] = &Metrics{}
		metricsLog.RxMetrics.Queues[i] = &Metrics{}
	}
	return metricsLog
}
