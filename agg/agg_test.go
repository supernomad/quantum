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
			NumWorkers:   1,
		})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	agg.Start(wg)
	agg.Aggs <- &Data{
		Direction: Outgoing,
		Dropped:   false,
		PrivateIP: "10.10.0.1",
	}
	agg.Aggs <- &Data{
		Direction: Incoming,
		Dropped:   true,
	}
	time.Sleep(2 * time.Second)
	agg.Stop()
	wg.Wait()
}
