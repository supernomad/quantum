// Copyright (c) 2016-2018 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package metric

import (
	"github.com/supernomad/quantum/common"
)

const (
	metricsBackLog = 1000
)

// Aggregator is a struct for monitoring and aggregating quantum metrics via an embedded channel.
type Aggregator struct {
	cfg        *common.Config
	stop       chan struct{}
	metricsLog *MetricsLog

	// Metrics is the channel Metric structs are sent to for aggregation and export via the rest api
	Metrics chan *Metric
}

func handleMetric(metrics *Metrics, metric *Metric) {
	if metric.Dropped {
		metrics.DroppedBytes += metric.Bytes
		metrics.DroppedPackets++
	} else {
		metrics.Bytes += metric.Bytes
		metrics.Packets++
	}
}

func (aggregator *Aggregator) pipeline(metric *Metric) {
	aggregator.cfg.Log.Debug.Println("[AGGREGATOR]", "Metric data received:", metric)

	var metrics *Metrics
	switch metric.Type {
	case Rx:
		metrics = aggregator.metricsLog.RxMetrics
	case Tx:
		metrics = aggregator.metricsLog.TxMetrics
	}

	handleMetric(metrics, metric)

	if queueMetrics, ok := metrics.Queues[metric.Queue]; ok {
		handleMetric(queueMetrics, metric)
	} else {
		queueMetrics = &Metrics{}
		handleMetric(queueMetrics, metric)

		metrics.Queues[metric.Queue] = queueMetrics
	}

	if metric.PrivateIP == "" {
		return
	}

	if linkMetrics, ok := metrics.Links[metric.PrivateIP]; ok {
		handleMetric(linkMetrics, metric)
	} else {
		linkMetrics = &Metrics{}
		handleMetric(linkMetrics, metric)

		metrics.Links[metric.PrivateIP] = linkMetrics
	}
}

// Start aggregating and serving requests for statistics data.
func (aggregator *Aggregator) Start() {
	go func() {
	loop:
		for {
			select {
			case <-aggregator.stop:
				break loop
			case metric := <-aggregator.Metrics:
				aggregator.pipeline(metric)
			}
		}
		close(aggregator.stop)
		close(aggregator.Metrics)
	}()
}

// Stop aggregating and receiving requests for statistics data.
func (aggregator *Aggregator) Stop() {
	aggregator.stop <- struct{}{}
}

// Bytes returns a byte slice json representation of the underlying MetricsLog struct in either flat or prettified notation, if there is an error while marshalling data a nil slice is returned.
func (aggregator *Aggregator) Bytes(pretty bool) []byte {
	return aggregator.metricsLog.Bytes(pretty)
}

// New generates an Aggregator instance for aggregating statistics data for quantum.
func New(cfg *common.Config) *Aggregator {
	return &Aggregator{
		cfg:        cfg,
		stop:       make(chan struct{}),
		metricsLog: newMetricsLog(cfg.NumWorkers),
		Metrics:    make(chan *Metric, metricsBackLog),
	}
}
