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

func (agg *Agg) aggData() (incomingStats *common.Stats, outgoingStats *common.Stats) {
	incomingStats = common.NewStats()
	outgoingStats = common.NewStats()

	for i := 0; i < agg.cfg.NumWorkers; i++ {
		incomingStats.Packets += agg.incomingStats[i].Packets
		incomingStats.Bytes += agg.incomingStats[i].Bytes

		for k, v := range agg.incomingStats[i].Links {
			if link, ok := incomingStats.Links[k]; !ok {
				incomingStats.Links[k] = &common.Stats{
					Packets: v.Packets,
					Bytes:   v.Bytes,
				}
			} else {
				link.Packets += v.Packets
				link.Bytes += v.Bytes
			}
		}

		outgoingStats.Packets += agg.outgoingStats[i].Packets
		outgoingStats.Bytes += agg.outgoingStats[i].Bytes

		for k, v := range agg.outgoingStats[i].Links {
			if link, ok := outgoingStats.Links[k]; !ok {
				outgoingStats.Links[k] = &common.Stats{
					Packets: v.Packets,
					Bytes:   v.Bytes,
				}
			} else {
				link.Packets += v.Packets
				link.Bytes += v.Bytes
			}
		}
	}
	return
}

func (agg *Agg) diffData(elapsed float64, incomingStats *common.Stats, outgoingStats *common.Stats) {
	incomingStats.PacketsDiff = incomingStats.Packets - agg.lastIncomingStats.Packets
	incomingStats.PPS = float64(incomingStats.PacketsDiff) / elapsed
	incomingStats.BytesDiff = incomingStats.Bytes - agg.lastIncomingStats.Bytes
	incomingStats.Bandwidth = float64(incomingStats.BytesDiff) / elapsed

	for key, incomingLink := range incomingStats.Links {
		if lastIncomingLink, exists := agg.lastIncomingStats.Links[key]; exists {
			incomingLink.PacketsDiff = incomingLink.Packets - lastIncomingLink.Packets
			incomingLink.PPS = float64(incomingLink.PacketsDiff) / elapsed
			incomingLink.BytesDiff = incomingLink.Bytes - lastIncomingLink.Bytes
			incomingLink.Bandwidth = float64(incomingLink.BytesDiff) / elapsed
		} else {
			incomingLink.PacketsDiff = incomingLink.Packets
			incomingLink.PPS = float64(incomingLink.PacketsDiff) / elapsed
			incomingLink.BytesDiff = incomingLink.Bytes
			incomingLink.Bandwidth = float64(incomingLink.BytesDiff) / elapsed
		}
	}

	outgoingStats.PacketsDiff = outgoingStats.Packets - agg.lastOutgoingStats.Packets
	outgoingStats.PPS = float64(outgoingStats.PacketsDiff) / elapsed
	outgoingStats.BytesDiff = outgoingStats.Bytes - agg.lastOutgoingStats.Bytes
	outgoingStats.Bandwidth = float64(outgoingStats.BytesDiff) / elapsed

	for key, outgoingLink := range outgoingStats.Links {
		if lastOutgoingLink, exists := agg.lastOutgoingStats.Links[key]; exists {
			outgoingLink.PacketsDiff = outgoingLink.Packets - lastOutgoingLink.Packets
			outgoingLink.PPS = float64(outgoingLink.PacketsDiff) / elapsed
			outgoingLink.BytesDiff = outgoingLink.Bytes - lastOutgoingLink.Bytes
			outgoingLink.Bandwidth = float64(outgoingLink.BytesDiff) / elapsed
		} else {
			outgoingLink.PacketsDiff = outgoingLink.Packets
			outgoingLink.PPS = float64(outgoingLink.PacketsDiff) / elapsed
			outgoingLink.BytesDiff = outgoingLink.Bytes
			outgoingLink.Bandwidth = float64(outgoingLink.BytesDiff) / elapsed
		}
	}

	return
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

	incomingStats, outgoingStats := agg.aggData()
	agg.diffData(elapsedSec, incomingStats, outgoingStats)

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
