package datastore

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/config"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/etcd"
	"io/ioutil"
	"path"
	"time"
)

// Backend - The type of datastore to use
type Backend string

const (
	etcdBackend   Backend = "etcd"
	consulBackend Backend = "consul"

	updateAction int = 0
	removeAction int = 1
)

// Datastore - The datastore object handles syncing the mappings from the configured backend
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

func (datastore *Datastore) mappingHandler(action int, key string, value []byte) {
	switch action {
	case updateAction:
		mapping, err := common.ParseMapping(value, datastore.privateKey)
		if err != nil {
			return
		}
		datastore.Mappings[common.IPtoInt(key)] = mapping
	case removeAction:
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
		datastore.mappingHandler(updateAction, key, keyVal.Value)
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
					datastore.mappingHandler(removeAction, key, keyVal.Value)
				} else {
					datastore.mappingHandler(updateAction, key, keyVal.Value)
				}
			}
		}
	}
}

// Start handling the datastore backend sync and watch
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

func handleTLS(options *store.Config, cfg *config.Config) error {
	if cfg.TLSKey != "" && cfg.TLSCert != "" {
		cert, err := tls.LoadX509KeyPair(cfg.TLSCert, cfg.TLSKey)
		if err != nil {
			return err
		}

		config := &tls.Config{Certificates: []tls.Certificate{cert}}
		if cfg.TLSCA != "" {
			// Load CA cert
			ca, err := ioutil.ReadFile(cfg.TLSCA)
			if err != nil {
				return err
			}
			caPool := x509.NewCertPool()
			caPool.AppendCertsFromPEM(ca)
			config.RootCAs = caPool
		}

		config.BuildNameToCertificate()
		options.TLS = config
	}
	return nil
}

// New datastore
func New(privateKey []byte, mapping *common.Mapping, cfg *config.Config) (*Datastore, error) {
	options := &store.Config{
		//TODO:: ConnectionTimeout: connTimeout,
		//TODO:: Bucket: "quantum",
		PersistConnection: true,
		Username:          cfg.Username,
		Password:          cfg.Password,
	}

	err := handleTLS(options, cfg)
	if err != nil {
		return nil, err
	}

	switch Backend(cfg.Datastore) {
	case etcdBackend:
		store, err := libkv.NewStore(store.ETCD, cfg.Endpoints, options)
		if err != nil {
			return nil, err
		}
		return toDatastore(privateKey, mapping, store, cfg)
	case consulBackend:
		store, err := libkv.NewStore(store.CONSUL, cfg.Endpoints, options)
		if err != nil {
			return nil, err
		}
		return toDatastore(privateKey, mapping, store, cfg)
	default:
		return nil, errors.New("The specified 'Backend' is not supported.")
	}
}

func init() {
	consul.Register()
	etcd.Register()
}
