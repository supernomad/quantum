// Copyright (c) 2016-2018 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package metric

import (
	"testing"
	"time"

	"github.com/supernomad/quantum/common"
)

func TestAggregator(t *testing.T) {
	cfg := &common.Config{
		Log:        common.NewLogger(common.NoopLogger),
		NumWorkers: 1,
	}

	aggregator := New(cfg)

	aggregator.Start()

	aggregator.Metrics <- &Metric{
		Type:      Tx,
		Dropped:   false,
		PrivateIP: "10.99.0.1",
		Bytes:     20,
	}
	aggregator.Metrics <- &Metric{
		Type:      Tx,
		Dropped:   false,
		PrivateIP: "10.99.0.1",
		Bytes:     20,
	}
	aggregator.Metrics <- &Metric{
		Type:    Rx,
		Dropped: true,
		Bytes:   20,
	}
	aggregator.Metrics <- &Metric{
		Type:      Rx,
		Dropped:   false,
		PrivateIP: "10.99.0.1",
		Bytes:     20,
	}

	time.Sleep(1 * time.Millisecond)

	buf := aggregator.Bytes(true)
	if buf == nil {
		t.Fatal("Bytes returned a nil slice when asking for a prettified version.")
	}

	buf = aggregator.Bytes(false)
	if buf == nil {
		t.Fatal("Bytes returned a nil slice when asking for a flattened version.")
	}

	aggregator.Stop()
}
