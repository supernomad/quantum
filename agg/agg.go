// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package agg

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Supernomad/quantum/common"
)

// Agg a statistics aggregation struct.
type Agg struct {
	log      *common.Logger
	cfg      *common.Config
	stop     chan struct{}
	statsLog *common.StatsLog

	// Aggs is the channel Data structs are sent to for aggregation and export via the rest api
	Aggs chan *common.Stat
}

func handleStats(stats *common.Stats, aggStat *common.Stat) {
	if !aggStat.Dropped {
		stats.Bytes += aggStat.Bytes
		stats.Packets++
	} else {
		stats.DroppedBytes += aggStat.Bytes
		stats.DroppedPackets++
	}
}

func (agg *Agg) returnStats(w http.ResponseWriter, r *http.Request) {
	agg.log.Debug.Println("[AGG]", "Recieved an api request:", r)

	header := w.Header()
	header.Set("Content-Type", "application/json")
	header.Set("Server", "quantum")

	_, err := w.Write(agg.statsLog.Bytes(strings.Contains(r.RequestURI, "pretty")))
	if err != nil {
		agg.log.Error.Println("[AGG]", "Error writing stats api response:", err.Error())
	}
}

func (agg *Agg) pipeline(aggStat *common.Stat) {
	agg.log.Debug.Println("[AGG]", "Statistics data recieved:", aggStat)

	var stats *common.Stats
	switch aggStat.Direction {
	case common.IncomingStat:
		stats = agg.statsLog.RxStats
	case common.OutgoingStat:
		stats = agg.statsLog.TxStats
	}

	handleStats(stats, aggStat)
	handleStats(stats.Queues[aggStat.Queue], aggStat)

	if aggStat.PrivateIP == "" {
		return
	}

	if linkStats, ok := stats.Links[aggStat.PrivateIP]; ok {
		handleStats(linkStats, aggStat)
	} else {
		linkStats = &common.Stats{}
		handleStats(linkStats, aggStat)

		stats.Links[aggStat.PrivateIP] = linkStats
	}
}

func (agg *Agg) server() {
	listenAddress := fmt.Sprintf("%s:%d", agg.cfg.StatsAddress, agg.cfg.StatsPort)
	http.HandleFunc(agg.cfg.StatsRoute, agg.returnStats)
	for {
		err := http.ListenAndServe(listenAddress, nil)
		if err != nil {
			agg.log.Error.Println("[AGG]", "Error initializing stats api:", err.Error())
		}

		time.Sleep(10 * time.Second)
	}
}

// Start aggregating and serving requests for statistics data.
func (agg *Agg) Start() {
	go agg.server()
	go func() {
	loop:
		for {
			select {
			case <-agg.stop:
				break loop
			case aggStat := <-agg.Aggs:
				agg.pipeline(aggStat)
			}
		}
		close(agg.stop)
	}()
}

// Stop aggregating and recieving requests for statistics data.
func (agg *Agg) Stop() {
	agg.stop <- struct{}{}
}

// New generates an Agg instance for aggregating statistics data for quantum and exposing those statistics via a REST api interface.
func New(log *common.Logger, cfg *common.Config) *Agg {
	return &Agg{
		log: log,
		cfg: cfg,
		statsLog: &common.StatsLog{
			RxStats: common.NewStats(cfg.NumWorkers),
			TxStats: common.NewStats(cfg.NumWorkers),
		},
		stop: make(chan struct{}),
		Aggs: make(chan *common.Stat, 1024*1024),
	}
}
