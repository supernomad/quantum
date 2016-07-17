package datastore

import (
	"errors"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/config"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/etcd"
	"path"
	"time"
)

type DatastoreType string
type DatastoreAction int

var (
	EtcdDatastore   DatastoreType = "etcd"
	ConsulDatastore DatastoreType = "consul"
)

var (
	UpdateAction DatastoreAction = 0
	RemoveAction DatastoreAction = 1
)

type Datastore struct {
	store  store.Store
	prefix string

	privateKey []byte
	privateIP  string
	mapping    *common.Mapping

	syncInterval time.Duration
	leaseTime    time.Duration
	retries      time.Duration

	Mappings map[uint32]*common.Mapping
}

func (datastore *Datastore) mappingHandler(action DatastoreAction, key string, value []byte) {
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

func (datastore *Datastore) sync() error {
	nodes, err := datastore.store.List(path.Join(datastore.prefix, "/mappings/"))
	if err != nil {
		return err
	}

	for _, keyVal := range nodes {
		_, key := path.Split(keyVal.Key)
		datastore.mappingHandler(UpdateAction, key, keyVal.Value)
	}
	return nil
}

func (datastore *Datastore) set() error {
	return datastore.store.Put(
		path.Join(datastore.prefix, "/mappings/", datastore.privateIP),
		datastore.mapping.Bytes(),
		&store.WriteOptions{TTL: datastore.leaseTime * time.Second})
}

func toDatastore(privateKey []byte, mapping *common.Mapping, store store.Store, cfg *config.Config) (*Datastore, error) {
	datastore := &Datastore{
		store:  store,
		prefix: cfg.Prefix,

		privateKey: privateKey,
		privateIP:  cfg.PrivateIP,
		mapping:    mapping,

		syncInterval: cfg.SyncInterval,
		leaseTime:    cfg.LeaseTime,
		retries:      cfg.Retries,

		Mappings: make(map[uint32]*common.Mapping),
	}

	err := datastore.set()
	return datastore, err
}

func (datastore *Datastore) watch() {
	events, err := datastore.store.WatchTree(path.Join(datastore.prefix, "/mappings/"), nil)
	if err != nil {
		return
	}

	for {
		select {
		case pairs := <-events:
			for _, keyVal := range pairs {
				_, key := path.Split(keyVal.Key)
				if len(keyVal.Value) == 0 {
					datastore.mappingHandler(RemoveAction, key, keyVal.Value)
				} else {
					datastore.mappingHandler(UpdateAction, key, keyVal.Value)
				}
			}
		}
	}
}

func (datastore *Datastore) Start() error {
	err := datastore.sync()
	if err != nil {
		return err
	}
	refresh := time.NewTicker(datastore.leaseTime / datastore.retries * time.Second)
	sync := time.NewTicker(datastore.syncInterval * time.Second)

	go datastore.watch()
	go func() {
		for {
			select {
			case <-refresh.C:
				datastore.set()
			case <-sync.C:
				datastore.sync()
			}
		}
	}()
	return nil
}

func New(privateKey []byte, mapping *common.Mapping, cfg *config.Config) (*Datastore, error) {
	options := &store.Config{
		//TODO:: ClientTLS: cliTlsCfg,
		//TODO:: TLS: tlsCfg,
		//TODO:: ConnectionTimeout: connTimeout,
		//TODO:: Bucket: "quantum",
		PersistConnection: true,
		Username:          cfg.Username,
		Password:          cfg.Password,
	}

	switch DatastoreType(cfg.Datastore) {
	case EtcdDatastore:
		store, err := libkv.NewStore(store.ETCD, cfg.Endpoints, options)
		if err != nil {
			return nil, err
		}
		return toDatastore(privateKey, mapping, store, cfg)
	case ConsulDatastore:
		store, err := libkv.NewStore(store.CONSUL, cfg.Endpoints, options)
		if err != nil {
			return nil, err
		}
		return toDatastore(privateKey, mapping, store, cfg)
	default:
		return nil, errors.New("The specified 'DatastoreType' is not supported.")
	}
}

func init() {
	consul.Register()
	etcd.Register()
}
