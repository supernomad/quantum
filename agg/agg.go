package agg

import (
	"github.com/Supernomad/quantum/common"
	"sync"
	"time"
)

// Agg a statistics aggregation object
type Agg struct {
	cfg *common.Config

	stop   chan struct{}
	ticker *time.Ticker

	incomingStats []*common.Stats
	outgoingStats []*common.Stats

	sinks []StatSink
}

// StatSink to dump statistics into
type StatSink interface {
	SendStats(statsl *StatsLog) error
}

func aggregateStats(stats []*common.Stats) *common.Stats {
	aggStats := common.NewStats()
	for i := 0; i < len(stats); i++ {
		aggStats.DroppedPackets += stats[i].DroppedPackets
		aggStats.Packets += stats[i].Packets
		aggStats.DroppedBytes += stats[i].DroppedBytes
		aggStats.Bytes += stats[i].Bytes

		for k, statLink := range stats[i].Links {
			if aggLink, ok := aggStats.Links[k]; !ok {
				aggStats.Links[k] = &common.Stats{
					DroppedPackets: statLink.DroppedPackets,
					Packets:        statLink.Packets,
					DroppedBytes:   statLink.DroppedBytes,
					Bytes:          statLink.Bytes,
				}
			} else {
				aggLink.DroppedPackets += statLink.DroppedPackets
				aggLink.Packets += statLink.Packets
				aggLink.DroppedBytes += statLink.DroppedBytes
				aggLink.Bytes += statLink.Bytes
			}
		}
	}
	return aggStats
}

func (agg *Agg) sendData(statsl *StatsLog) error {
	for i := 0; i < len(agg.sinks); i++ {
		if err := agg.sinks[i].SendStats(statsl); err != nil {
			return err
		}
	}
	return nil
}

func (agg *Agg) pipeline() {
	incomingStats := aggregateStats(agg.incomingStats)
	outgoingStats := aggregateStats(agg.outgoingStats)

	statsl := &StatsLog{
		TxStats:      outgoingStats,
		TxQueueStats: agg.outgoingStats,
		RxStats:      incomingStats,
		RxQueueStats: agg.incomingStats,
	}

	agg.sendData(statsl)
}

// Start aggregating and sending stats data
func (agg *Agg) Start(wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()
	loop:
		for {
			select {
			case <-agg.stop:
				break loop
			case <-agg.ticker.C:
				agg.pipeline()
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
func New(log *common.Logger, cfg *common.Config, incomingStats []*common.Stats, outgoingStats []*common.Stats) *Agg {
	return &Agg{
		cfg:           cfg,
		sinks:         []StatSink{&ConsoleSink{log: log}},
		ticker:        time.NewTicker(cfg.StatsWindow),
		stop:          make(chan struct{}),
		incomingStats: incomingStats,
		outgoingStats: outgoingStats,
	}
}
