package agg

import (
	"github.com/Supernomad/quantum/common"
	"sync"
	"testing"
	"time"
)

func TestAgg(t *testing.T) {
	agg := New(
		common.NewLogger(),
		&common.Config{
			StatsRoute:   "/stats",
			StatsPort:    1099,
			StatsAddress: "127.0.0.1",
		})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	agg.Start(wg)
	time.Sleep(2 * time.Second)
	agg.Stop()
}
