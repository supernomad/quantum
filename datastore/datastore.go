package datastore

import (
	"errors"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/config"
	"time"
)

type DatastoreType int
type DatastoreAction int

var (
	EtcdDatastore   DatastoreType = 0
	ConsulDatastore DatastoreType = 1
)

var (
	UpdateAction DatastoreAction = 0
	RemoveAction DatastoreAction = 1
)

type MappingHandler func(DatastoreAction, string, string)

type DatastoreBackend interface {
	Watch(string, MappingHandler)
	Sync(string, MappingHandler) error
	ResetTtl(string, string, time.Duration) error
	Set(string, string, time.Duration, *common.Mapping) error
}

type Datastore struct {
	backend DatastoreBackend
	prefix  string

	privateKey []byte
	privateIp  string
	mapping    *common.Mapping

	syncInterval time.Duration
	leaseTime    time.Duration
	retries      time.Duration

	Mappings map[uint32]*common.Mapping
}

func (datastore *Datastore) Start() error {
	err := datastore.backend.Sync(datastore.prefix, datastore.mappingHandler)
	if err != nil {
		return err
	}
	refresh := time.NewTicker(datastore.leaseTime / datastore.retries * time.Second)
	sync := time.NewTicker(datastore.syncInterval * time.Second)

	go datastore.backend.Watch(datastore.prefix, datastore.mappingHandler)
	go func() {
		for {
			select {
			case <-refresh.C:
				datastore.backend.ResetTtl(datastore.prefix, datastore.privateIp, datastore.leaseTime)
			case <-sync.C:
				datastore.backend.Sync(datastore.prefix, datastore.mappingHandler)
			}
		}
	}()
	return nil
}

func (datastore *Datastore) mappingHandler(action DatastoreAction, key, value string) {
	switch action {
	case UpdateAction:
		mapping, err := common.ParseMapping(value, datastore.privateKey)
		if err != nil {
			return
		}
		datastore.Mappings[common.IPtoInt(key)] = mapping
	case RemoveAction:
		delete(datastore.Mappings, common.IPtoInt(key))
	}
}

func toDatastore(privateKey []byte, mapping *common.Mapping, backend DatastoreBackend, cfg *config.Config) (*Datastore, error) {
	datastore := &Datastore{
		backend: backend,
		prefix:  cfg.Prefix,

		privateKey: privateKey,
		privateIp:  cfg.PrivateIP,
		mapping:    mapping,

		syncInterval: cfg.SyncInterval,
		leaseTime:    cfg.LeaseTime,
		retries:      cfg.Retries,

		Mappings: make(map[uint32]*common.Mapping),
	}

	err := datastore.backend.Set(datastore.prefix, datastore.privateIp, datastore.leaseTime, datastore.mapping)
	return datastore, err
}

func New(datastoreType DatastoreType, privateKey []byte, mapping *common.Mapping, cfg *config.Config) (*Datastore, error) {
	switch datastoreType {
	case EtcdDatastore:
		backend, err := newEtcd(cfg)
		if err != nil {
			return nil, err
		}
		return toDatastore(privateKey, mapping, backend, cfg)
	default:
		return nil, errors.New("The specified 'DatastoreType' is not supported.")
	}
}
