// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package rest

import (
	"net/http"
	"testing"
	"time"

	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/metric"
)

func TestRest(t *testing.T) {
	cfg := &common.Config{
		Log:          common.NewLogger(common.NoopLogger),
		StatsRoute:   "/metrics",
		StatsPort:    1099,
		StatsAddress: "127.0.0.1",
		NumWorkers:   1,
	}

	aggregator := metric.New(cfg)
	api := New(cfg, aggregator)

	api.Start()
	aggregator.Start()

	aggregator.Metrics <- &metric.Metric{
		Type:      metric.Tx,
		Dropped:   false,
		PrivateIP: "10.99.0.1",
		Bytes:     20,
	}
	aggregator.Metrics <- &metric.Metric{
		Type:      metric.Tx,
		Dropped:   false,
		PrivateIP: "10.99.0.1",
		Bytes:     20,
	}
	aggregator.Metrics <- &metric.Metric{
		Type:    metric.Rx,
		Dropped: true,
		Bytes:   20,
	}
	aggregator.Metrics <- &metric.Metric{
		Type:      metric.Rx,
		Dropped:   false,
		PrivateIP: "10.99.0.1",
		Bytes:     20,
	}

	time.Sleep(1 * time.Second)

	_, err := http.Get("http://127.0.0.1:1099/metrics")
	if err != nil {
		t.Fatal(err)
	}

	aggregator.Stop()
	api.Stop()
}
