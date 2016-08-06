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
	"net"
	"path"
	"time"
)

// Need to register the backends with libkv
func init() {
	consul.Register()
	etcd.Register()
}

// Type of datastore backend to utilize
type Type string

const (
	// ETCD datastore backend
	ETCD Type = "etcd"
	// CONSUL datastore backend
	CONSUL Type = "consul"

	lockTTL time.Duration = time.Duration(5)

	update int = 0
	remove int = 1
)

// Datastore object which handles syncronization of mapping and network configuration
type Datastore struct {
	store  store.Store
	prefix string

	privateKey []byte
	privateIP  string
	mapping    *common.Mapping

	syncInterval    time.Duration
	leaseTime       time.Duration
	refreshInterval time.Duration

	Mappings map[uint32]*common.Mapping
	Network  string

	stop chan struct{}
}

func (datastore *Datastore) mappingHandler(action int, key string, value []byte) {
	switch action {
	case update:
		mapping, err := common.ParseMapping(value, datastore.privateKey)
		if err != nil {
			return
		}
		datastore.Mappings[common.IPtoInt(key)] = mapping
	case remove:
		delete(datastore.Mappings, common.IPtoInt(key))
	}
}

func (datastore *Datastore) locker() (store.Locker, error) {
	lockOps := &store.LockOptions{TTL: lockTTL * time.Second}
	locker, err := datastore.store.NewLock(path.Join(datastore.prefix, "/lock"), lockOps)
	if err != nil {
		return nil, err
	}
	return locker, nil
}

func (datastore *Datastore) set() error {
	return datastore.store.Put(
		path.Join(datastore.prefix, "/mappings/", datastore.privateIP),
		datastore.mapping.Bytes(),
		&store.WriteOptions{TTL: datastore.leaseTime * time.Second})
}

func (datastore *Datastore) sync() error {
	nodes, err := datastore.store.List(path.Join(datastore.prefix, "/mappings/"))
	if err != nil {
		return err
	}

	for _, keyVal := range nodes {
		_, key := path.Split(keyVal.Key)
		datastore.mappingHandler(update, key, keyVal.Value)
	}
	return nil
}

func (datastore *Datastore) syncNetwork() error {
	config, err := datastore.store.Get(path.Join(datastore.prefix, "config"))
	if err != nil {
		return err
	}

	datastore.Network = string(config.Value)
	return nil
}

func (datastore *Datastore) watch() {
	events, err := datastore.store.WatchTree(path.Join(datastore.prefix, "/mappings/"), datastore.stop)
	if err != nil {
		return
	}

	for {
		select {
		case pairs := <-events:
			for _, keyVal := range pairs {
				_, key := path.Split(keyVal.Key)
				if len(keyVal.Value) == 0 {
					datastore.mappingHandler(remove, key, keyVal.Value)
				} else {
					datastore.mappingHandler(update, key, keyVal.Value)
				}
			}
		}
	}
}

// Start handling the datastore backend sync and watch
func (datastore *Datastore) Start() {
	refresh := time.NewTicker(datastore.refreshInterval * time.Second)
	sync := time.NewTicker(datastore.syncInterval * time.Second)

	go datastore.watch()
	go func() {
		for {
			select {
			case <-datastore.stop:
				break
			case <-refresh.C:
				datastore.set()
			case <-sync.C:
				datastore.sync()
			}
		}
	}()
}

// Stop handling the datastore backend sync and watch
func (datastore *Datastore) Stop() {
	go func() {
		datastore.stop <- struct{}{}
	}()
}

func nextIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] > 0 {
			break
		}
	}
}

func (datastore *Datastore) getFreeIP() (string, error) {
	ip, ipnet, err := net.ParseCIDR(datastore.Network)
	if err != nil {
		return "", err
	}
	for ip = ip.Mask(ipnet.Mask); ipnet.Contains(ip); nextIP(ip) {
		if _, exists := datastore.Mappings[common.IPtoInt(ip.String())]; !exists {
			break
		}
	}
	return ip.String(), nil
}

func newDatastore(store store.Store, cfg *config.Config) (*Datastore, error) {
	datastore := &Datastore{
		store:           store,
		prefix:          cfg.Prefix,
		privateKey:      cfg.PrivateKey,
		syncInterval:    cfg.SyncInterval,
		leaseTime:       cfg.LeaseTime,
		refreshInterval: cfg.RefreshInterval,
		Mappings:        make(map[uint32]*common.Mapping),
		stop:            make(chan struct{}),
	}

	locker, err := datastore.locker()
	if err != nil {
		return nil, err
	}

	_, err = locker.Lock(nil)
	if err != nil {
		return nil, err
	}

	err = datastore.sync()
	if err != nil {
		return nil, err
	}

	err = datastore.syncNetwork()
	if err != nil {
		return nil, err
	}

	if cfg.PrivateIP == "" {
		free, err := datastore.getFreeIP()
		if err != nil {
			return nil, err
		}
		cfg.PrivateIP = free
	}

	datastore.privateIP = cfg.PrivateIP
	datastore.mapping = common.NewMapping(cfg.PublicAddress, cfg.PublicKey)

	err = datastore.set()
	if err != nil {
		return nil, err
	}

	err = locker.Unlock()
	if err != nil {
		return nil, err
	}

	return datastore, nil
}

func newStoreConfig(cfg *config.Config) (*store.Config, error) {
	storeCfg := &store.Config{PersistConnection: true}
	if cfg.Username != "" && cfg.Password != "" {
		storeCfg.Username = cfg.Username
		storeCfg.Password = cfg.Password
	}

	if cfg.TLSEnabled {
		storeCfg.TLS = &tls.Config{}

		if cfg.TLSKey != "" && cfg.TLSCert != "" {
			cert, err := tls.LoadX509KeyPair(cfg.TLSCert, cfg.TLSKey)

			if err != nil {
				return nil, err
			}

			storeCfg.TLS.Certificates = []tls.Certificate{cert}
		}

		if cfg.TLSCA != "" {
			cert, err := ioutil.ReadFile(cfg.TLSCA)

			if err != nil {
				return nil, err
			}

			storeCfg.TLS.RootCAs = x509.NewCertPool()
			storeCfg.TLS.RootCAs.AppendCertsFromPEM(cert)
		}

		storeCfg.TLS.BuildNameToCertificate()
	}
	return storeCfg, nil
}

// New datastore object
func New(cfg *config.Config) (*Datastore, error) {
	storeCfg, err := newStoreConfig(cfg)
	if err != nil {
		return nil, err
	}

	switch Type(cfg.Datastore) {
	case ETCD:
		store, err := libkv.NewStore(store.ETCD, cfg.Endpoints, storeCfg)
		if err != nil {
			return nil, err
		}
		return newDatastore(store, cfg)
	case CONSUL:
		store, err := libkv.NewStore(store.CONSUL, cfg.Endpoints, storeCfg)
		if err != nil {
			return nil, err
		}
		return newDatastore(store, cfg)
	default:
		return nil, errors.New("The specified datastore backend is not supported.")
	}
}
