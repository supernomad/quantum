package agg

import (
	"github.com/Supernomad/quantum/common"
	"testing"
	"time"
)

func TestAgg(t *testing.T) {
	stats := make([]*common.Stats, 3)
	for i := 0; i < 3; i++ {
		stats[i] = common.NewStats()
	}
	agg := New(&common.Config{
		StatsWindow: 1,
		NumWorkers:  3,
	}, stats, stats)
	agg.Start()
	time.Sleep(2 * time.Second)
	agg.Stop()
}
