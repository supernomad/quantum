package datastore

import (
	"errors"
	"sync"
	"time"

	"github.com/Supernomad/quantum/common"
)

const (
	// ETCDDatastore will tell quantum to use etcd as the backend datastore
	ETCDDatastore = iota
	// MOCKDatastore will tell quantum to use a moked out backend datastore for testing
	MOCKDatastore

	lockTTL = 10 * time.Second
)

// Datastore interface for quantum to get mapping data from the backend datastore
type Datastore interface {
	Init() error
	Mapping(ip uint32) (*common.Mapping, bool)
	Start(wg *sync.WaitGroup)
	Stop()
}

// New datastore object
func New(kind int, log *common.Logger, cfg *common.Config) (Datastore, error) {
	switch kind {
	case ETCDDatastore:
		return newEtcd(log, cfg)
	case MOCKDatastore:
		return newMock(log, cfg)
	default:
		return nil, errors.New("specified backend doesn't exist")
	}
}
