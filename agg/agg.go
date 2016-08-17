package agg

import (
	"github.com/Supernomad/quantum/common"
	"time"
)

// Agg a statistics aggregation object
type Agg struct {
	cfg    *common.Config
	start  time.Time
	stop   chan struct{}
	ticker *time.Ticker

	lastIncomingStats *common.Stats
	lastOutgoingStats *common.Stats

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
		aggStats.Packets += stats[i].Packets
		aggStats.Bytes += stats[i].Bytes

		for k, statLink := range stats[i].Links {
			if aggLink, ok := aggStats.Links[k]; !ok {
				aggStats.Links[k] = &common.Stats{
					Packets: statLink.Packets,
					Bytes:   statLink.Bytes,
				}
			} else {
				aggLink.Packets += statLink.Packets
				aggLink.Bytes += statLink.Bytes
			}
		}
	}
	return aggStats
}

func diffStats(elapsed float64, current, last *common.Stats) {
	current.PacketsDiff = current.Packets - last.Packets
	current.PPS = float64(current.PacketsDiff) / elapsed

	current.BytesDiff = current.Bytes - last.Bytes
	current.Bandwidth = float64(current.BytesDiff) / elapsed

	for k, currentLink := range current.Links {
		if lastLink, ok := last.Links[k]; ok {
			currentLink.PacketsDiff = currentLink.Packets - lastLink.Packets
			currentLink.BytesDiff = currentLink.Bytes - lastLink.Bytes
		} else {
			currentLink.PacketsDiff = currentLink.Packets
			currentLink.BytesDiff = currentLink.Bytes
		}

		currentLink.PPS = float64(currentLink.PacketsDiff) / elapsed
		currentLink.Bandwidth = float64(currentLink.BytesDiff) / elapsed
	}
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
	elapsed := time.Since(agg.start)
	elapsedSec := elapsed.Seconds()

	incomingStats := aggregateStats(agg.incomingStats)
	outgoingStats := aggregateStats(agg.outgoingStats)

	diffStats(elapsedSec, incomingStats, agg.lastIncomingStats)
	diffStats(elapsedSec, outgoingStats, agg.lastOutgoingStats)

	statsl := &StatsLog{
		TimeSpan: elapsedSec,

		TxStats: outgoingStats,
		RxStats: incomingStats,
	}

	agg.lastIncomingStats = incomingStats
	agg.lastOutgoingStats = outgoingStats

	agg.sendData(statsl)
	agg.start = time.Now()
}

// Start aggregating and sending stats data
func (agg *Agg) Start() {
	go func() {
		for {
			select {
			case <-agg.stop:
				return
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
func New(cfg *common.Config, incomingStats []*common.Stats, outgoingStats []*common.Stats) *Agg {
	return &Agg{
		cfg:               cfg,
		start:             time.Now(),
		sinks:             []StatSink{&ConsoleSink{}},
		ticker:            time.NewTicker(cfg.StatsWindow * time.Second),
		stop:              make(chan struct{}),
		lastIncomingStats: common.NewStats(),
		lastOutgoingStats: common.NewStats(),
		incomingStats:     incomingStats,
		outgoingStats:     outgoingStats,
	}
}
