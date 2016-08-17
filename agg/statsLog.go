package agg

import (
	"encoding/json"
	"github.com/Supernomad/quantum/common"
)

// StatsLog object to hold statistics information for quantum
type StatsLog struct {
	TimeSpan float64

	TxStats *common.Stats
	RxStats *common.Stats
}

// String the StatsLog object
func (statsl *StatsLog) String() string {
	data, _ := json.MarshalIndent(statsl, "", "    ")
	return string(data)
}
