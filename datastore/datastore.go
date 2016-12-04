package datastore

import (
	"github.com/Supernomad/quantum/common"
	"sync"
)

const (
	ETCDDatastore = iota
	MOCKDatastore
)

type Datastore interface {
	Init() error
	Mapping(ip uint32) (*common.Mapping, bool)
	Start(wg *sync.WaitGroup)
	Stop()
}

func New(kind int, log *common.Logger, cfg *common.Config) Datastore {
	switch kind {
	case ETCDDatastore:
		return newEtcd(log, cfg)
	case MOCKDatastore:
		return newMock(log, cfg)
	}
	return nil
}
