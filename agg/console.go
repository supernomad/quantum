package agg

import (
	"github.com/go-playground/log"
)

// ConsoleSink stats sink
type ConsoleSink struct {
}

// SendStats to the console stat sink
func (con *ConsoleSink) SendStats(statsl *StatsLog) error {
	log.Infof("Stats:\n%s", statsl)
	return nil
}
