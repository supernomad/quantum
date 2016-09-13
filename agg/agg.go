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

func diffStats(elapsed float64, current, last *common.Stats) {
	current.DroppedPacketsDiff = current.DroppedPackets - last.DroppedPackets
	current.DroppedPPS = float64(current.DroppedPacketsDiff) / elapsed

	current.PacketsDiff = current.Packets - last.Packets
	current.PPS = float64(current.PacketsDiff) / elapsed

	current.DroppedBytesDiff = current.DroppedBytes - last.DroppedBytes
	current.DroppedBandwidth = float64(current.DroppedBytesDiff) / elapsed

	current.BytesDiff = current.Bytes - last.Bytes
	current.Bandwidth = float64(current.BytesDiff) / elapsed

	for k, currentLink := range current.Links {
		if lastLink, ok := last.Links[k]; ok {
			currentLink.DroppedPacketsDiff = currentLink.DroppedPackets - lastLink.DroppedPackets
			currentLink.PacketsDiff = currentLink.Packets - lastLink.Packets
			currentLink.DroppedBytesDiff = currentLink.DroppedBytes - lastLink.DroppedBytes
			currentLink.BytesDiff = currentLink.Bytes - lastLink.Bytes
		} else {
			currentLink.DroppedPacketsDiff = currentLink.DroppedPackets
			currentLink.PacketsDiff = currentLink.Packets
			currentLink.DroppedBytesDiff = currentLink.DroppedBytes
			currentLink.BytesDiff = currentLink.Bytes
		}

		currentLink.DroppedPPS = float64(currentLink.DroppedPacketsDiff) / elapsed
		currentLink.PPS = float64(currentLink.PacketsDiff) / elapsed
		currentLink.DroppedBandwidth = float64(currentLink.DroppedBytesDiff) / elapsed
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
func New(log *common.Logger, cfg *common.Config, incomingStats []*common.Stats, outgoingStats []*common.Stats) *Agg {
	return &Agg{
		cfg:               cfg,
		start:             time.Now(),
		sinks:             []StatSink{&ConsoleSink{log: log}},
		ticker:            time.NewTicker(cfg.StatsWindow),
		stop:              make(chan struct{}),
		lastIncomingStats: common.NewStats(),
		lastOutgoingStats: common.NewStats(),
		incomingStats:     incomingStats,
		outgoingStats:     outgoingStats,
	}
}
