package agg

import (
	"fmt"
	"github.com/Supernomad/quantum/common"
	"net/http"
	"sync"
	"time"
)

const (
	Incoming = iota // 0
	Outgoing        // 1
)

// Agg a statistics aggregation object
type Agg struct {
	log *common.Logger
	cfg *common.Config

	rx *common.Stats
	tx *common.Stats

	Aggs chan *AggData
	stop chan struct{}
}

type AggData struct {
	PrivateIP string
	Bytes     uint64
	Direction uint64
	Dropped   bool
}

func handleStats(stats *common.Stats, aggData *AggData) {
	if !aggData.Dropped {
		stats.Bytes += aggData.Bytes
		stats.Packets += 1
	} else {
		stats.DroppedBytes += aggData.Bytes
		stats.DroppedPackets += 1
	}
}

func (agg *Agg) returnStats(w http.ResponseWriter, r *http.Request) {
	statsl := &StatsLog{
		RxStats: agg.rx,
		TxStats: agg.tx,
	}
	fmt.Fprintf(w, statsl.String())
}

func (agg *Agg) pipeline(aggData *AggData) {
	var stats *common.Stats
	switch aggData.Direction {
	case Incoming:
		stats = agg.rx
	case Outgoing:
		stats = agg.tx
	}

	handleStats(stats, aggData)

	if aggData.PrivateIP == "" {
		return
	}

	if linkStats, ok := stats.Links[aggData.PrivateIP]; ok {
		handleStats(linkStats, aggData)
	} else {
		linkStats = &common.Stats{}
		handleStats(linkStats, aggData)

		stats.Links[aggData.PrivateIP] = linkStats
	}
}

func (agg *Agg) server() {
	for {
		listenAddress := fmt.Sprintf("%s:%d", agg.cfg.StatsAddress, agg.cfg.StatsPort)

		http.HandleFunc(agg.cfg.StatsRoute, agg.returnStats)
		err := http.ListenAndServe(listenAddress, nil)
		if err != nil {
			agg.log.Error.Println(err.Error())
		}
		time.Sleep(10 * time.Second)
	}
}

// Start aggregating stats data
func (agg *Agg) Start(wg *sync.WaitGroup) {
	go agg.server()
	go func() {
		defer wg.Done()
	loop:
		for {
			select {
			case <-agg.stop:
				break loop
			case aggData := <-agg.Aggs:
				agg.pipeline(aggData)
			}
		}
	}()
}

// Stop aggregating and sending stats data
func (agg *Agg) Stop() {
	go func() {
		agg.stop <- struct{}{}
	}()
}

// New Agg instance pointer
func New(log *common.Logger, cfg *common.Config) *Agg {
	return &Agg{
		log:  log,
		cfg:  cfg,
		rx:   common.NewStats(),
		tx:   common.NewStats(),
		stop: make(chan struct{}),
		Aggs: make(chan *AggData, 1024*1024),
	}
}
