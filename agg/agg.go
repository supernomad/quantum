// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE
package agg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Supernomad/quantum/common"
)

const (
	// Incoming i.e. RX stats
	Incoming = iota // 0
	// Outgoing i.e. TX stats
	Outgoing // 1
)

// Agg a statistics aggregation object
type Agg struct {
	log      *common.Logger
	cfg      *common.Config
	stop     chan struct{}
	statsLog *StatsLog

	// Aggs is a channel to send Data objects to the centralized aggregator
	Aggs chan *Data
}

// StatsLog object to hold statistics information for quantum
type StatsLog struct {
	// TxStats holds the packet and byte counts for transmissions
	TxStats *common.Stats
	// RxStats holds the packet and byte counts for reception
	RxStats *common.Stats
}

// Bytes will return a byte slice representing the StatsLog object
func (statsl *StatsLog) Bytes() []byte {
	data, _ := json.Marshal(statsl)
	return data
}

// Data to use for statistics collection
type Data struct {
	PrivateIP string
	Queue     int
	Bytes     uint64
	Direction int
	Dropped   bool
}

func handleStats(stats *common.Stats, aggData *Data) {
	if !aggData.Dropped {
		stats.Bytes += aggData.Bytes
		stats.Packets++
	} else {
		stats.DroppedBytes += aggData.Bytes
		stats.DroppedPackets++
	}
}

func (agg *Agg) returnStats(w http.ResponseWriter, r *http.Request) {
	header := w.Header()
	header.Set("Content-Type", "application/json")
	header.Set("Server", "quantum")

	_, err := w.Write(agg.statsLog.Bytes())
	if err != nil {
		agg.log.Error.Println(err.Error())
	}
}

func (agg *Agg) pipeline(aggData *Data) {
	var stats *common.Stats
	switch aggData.Direction {
	case Incoming:
		stats = agg.statsLog.RxStats
	case Outgoing:
		stats = agg.statsLog.TxStats
	}

	handleStats(stats, aggData)
	handleStats(stats.Queues[aggData.Queue], aggData)

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
	listenAddress := fmt.Sprintf("%s:%d", agg.cfg.StatsAddress, agg.cfg.StatsPort)
	http.HandleFunc(agg.cfg.StatsRoute, agg.returnStats)
	for {
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
		log: log,
		cfg: cfg,
		statsLog: &StatsLog{
			RxStats: common.NewStats(cfg.NumWorkers),
			TxStats: common.NewStats(cfg.NumWorkers),
		},
		stop: make(chan struct{}),
		Aggs: make(chan *Data, 1024*1024),
	}
}
