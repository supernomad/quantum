package agg

import (
	"github.com/Supernomad/quantum/common"
	"sync"
	"testing"
	"time"
)

func TestAgg(t *testing.T) {
	stats := make([]*common.Stats, 3)
	for i := 0; i < 3; i++ {
		stats[i] = common.NewStats()
		stats[i].Links["10.0.0.0"] = common.NewStats()
	}
	stats[0].Links["10.0.0.1"] = common.NewStats()

	agg := New(
		common.NewLogger(),
		&common.Config{
			StatsWindow: 1 * time.Second,
			NumWorkers:  3,
		}, stats, stats)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	agg.Start(wg)
	time.Sleep(2 * time.Second)
	agg.Stop()
}
