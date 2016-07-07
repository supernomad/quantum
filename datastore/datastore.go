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
	ttl          time.Duration
	retries      time.Duration

	Mappings map[uint32]*common.Mapping
}

func (datastore *Datastore) Start() error {
	err := datastore.backend.Set(datastore.prefix, datastore.privateIp, datastore.ttl, datastore.mapping)
	if err != nil {
		return err
	}
	err = datastore.backend.Sync(datastore.prefix, datastore.mappingHandler)
	if err != nil {
		return err
	}
	refresh := time.NewTicker(datastore.ttl / datastore.retries * time.Second)
	sync := time.NewTicker(datastore.syncInterval * time.Second)

	go datastore.backend.Watch(datastore.prefix, datastore.mappingHandler)
	go func() {
		for {
			select {
			case <-refresh.C:
				datastore.backend.ResetTtl(datastore.prefix, datastore.privateIp, datastore.ttl)
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

func New(datastoreType DatastoreType, privateKey []byte, mapping *common.Mapping, cfg *config.Config) (*Datastore, error) {
	switch datastoreType {
	case EtcdDatastore:
		backend, err := newEtcd(cfg)
		if err != nil {
			return nil, err
		}
		return &Datastore{
			backend: backend,
			prefix:  cfg.Prefix,

			privateKey: privateKey,
			privateIp:  cfg.PrivateIP,
			mapping:    mapping,

			syncInterval: cfg.SyncInterval,
			ttl:          cfg.Ttl,
			retries:      cfg.Retries,

			Mappings: make(map[uint32]*common.Mapping),
		}, nil
	default:
		return nil, errors.New("The specified 'DatastoreType' is not supported.")
	}
}
