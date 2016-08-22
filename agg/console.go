package agg

import (
	"github.com/Supernomad/quantum/common"
)

// ConsoleSink stats sink
type ConsoleSink struct {
	log *common.Logger
}

// SendStats to the console stat sink
func (con *ConsoleSink) SendStats(statsl *StatsLog) error {
	con.log.Info.Println("[STATS]", statsl)
	return nil
}
