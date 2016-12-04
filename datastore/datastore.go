package datastore

import (
	"sync"
	"time"

	"github.com/Supernomad/quantum/common"
)

const (
	ETCDDatastore = iota
	MOCKDatastore

	lockTTL = 10 * time.Second
)

type Datastore interface {
	Init() error
	Mapping(ip uint32) (*common.Mapping, bool)
	Start(wg *sync.WaitGroup)
	Stop()
}

func New(kind int, log *common.Logger, cfg *common.Config) (Datastore, error) {
	switch kind {
	case ETCDDatastore:
		return newEtcd(log, cfg)
	case MOCKDatastore:
		return newMock(log, cfg)
	}
	return nil
}
