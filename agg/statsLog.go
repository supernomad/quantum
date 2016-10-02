package agg

import (
	"encoding/json"
	"github.com/Supernomad/quantum/common"
)

// StatsLog object to hold statistics information for quantum
type StatsLog struct {
	TxStats      *common.Stats
	RxStats      *common.Stats
	TxQueueStats []*common.Stats
	RxQueueStats []*common.Stats
}

// String the StatsLog object
func (statsl *StatsLog) String() string {
	data, _ := json.Marshal(statsl)
	return string(data)
}
