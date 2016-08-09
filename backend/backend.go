package backend

import (
	"errors"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/config"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/etcd"
	"github.com/go-playground/log"
	"path"
	"time"
)

func init() {
	consul.Register()
	etcd.Register()
}

const (
	consulStore string        = "consul"
	etcdStore   string        = "etcd"
	lockTTL     time.Duration = 10
)

// Backend datastore object which is responsible for managing state between the local node and the real backend datastore.
type Backend struct {
	store  store.Store
	locker store.Locker

	cfg        *config.Config
	NetworkCfg *common.NetworkConfig

	localMapping *common.Mapping
	mappings     map[uint32]*common.Mapping

	stop chan struct{}
}

func (backend *Backend) getKey(key string) string {
	return path.Join(backend.cfg.Prefix, key)
}

func (backend *Backend) lock() error {
	lockOps := &store.LockOptions{TTL: lockTTL * time.Second}
	locker, err := backend.store.NewLock(backend.getKey("/lock"), lockOps)
	if err != nil {
		return err
	}

	_, err = locker.Lock(nil)
	if err != nil {
		return err
	}
	backend.locker = locker
	return nil
}

func (backend *Backend) unlock() error {
	err := backend.locker.Unlock()
	if err != nil {
		return err
	}
	backend.locker = nil
	return nil
}

func (backend *Backend) getMappingIfExists() (*common.Mapping, bool) {
	node, err := backend.store.Get(backend.getKey("/nodes/" + backend.cfg.MachineID))
	if err != nil {
		return nil, false
	}
	mapping, err := common.ParseMapping(node.Value, backend.cfg.PrivateKey)
	if err != nil {
		return nil, false
	}
	return mapping, true
}

func (backend *Backend) getFreeIP() (string, error) {
	for ip := backend.NetworkCfg.BaseIP.Mask(backend.NetworkCfg.IPNet.Mask); backend.NetworkCfg.IPNet.Contains(ip); common.IncrementIP(ip) {
		str := ip.String()
		if _, exists := backend.mappings[common.IPtoInt(str)]; !exists {
			return str, nil
		}
	}
	return "", errors.New("There are no available ip addresses in the configured network.")
}

func (backend *Backend) handleLocalMapping() error {
	if backend.cfg.PrivateIP == "" {
		if mapping, ok := backend.getMappingIfExists(); ok {
			backend.cfg.PrivateIP = mapping.PrivateIP
		} else {
			ip, err := backend.getFreeIP()
			if err != nil {
				return err
			}
			backend.cfg.PrivateIP = ip
		}
	}

	mapping := common.NewMapping(backend.cfg.PrivateIP, backend.cfg.PublicAddress, backend.cfg.MachineID, backend.cfg.PublicKey)
	key := path.Join("/nodes/", backend.cfg.MachineID)

	err := backend.set(key, mapping.Bytes(), backend.NetworkCfg.LeaseTime)
	if err != nil {
		return err
	}
	backend.localMapping = mapping
	return nil
}

func (backend *Backend) syncMappings() error {
	nodes, err := backend.store.List(backend.getKey("/nodes/"))
	if err != nil {
		if err != store.ErrKeyNotFound {
			return err
		}
		nodes = make([]*store.KVPair, 0)
	}

	mappings := make(map[uint32]*common.Mapping)
	for _, node := range nodes {
		mapping, err := common.ParseMapping(node.Value, backend.cfg.PrivateKey)
		if err != nil {
			return err
		}
		mappings[common.IPtoInt(mapping.PrivateIP)] = mapping
	}
	backend.mappings = mappings
	return nil
}

func (backend *Backend) fetchNetworkConfig() error {
	netCfg, err := backend.store.Get(backend.getKey("config"))
	if err != nil {
		if err != store.ErrKeyNotFound {
			return err
		}
		backend.set("config", common.DefaultNetworkConfig.Bytes(), 0)
		backend.NetworkCfg = common.DefaultNetworkConfig
		return nil
	}
	networkCfg, err := common.ParseNetworkConfig(netCfg.Value)
	if err != nil {
		return err
	}
	backend.NetworkCfg = networkCfg
	return nil
}

func (backend *Backend) set(key string, value []byte, leaseTime time.Duration) error {
	var writeOps *store.WriteOptions
	if leaseTime > 0 {
		writeOps = &store.WriteOptions{TTL: leaseTime * time.Second}
	}
	return backend.store.Put(
		backend.getKey(key),
		value,
		writeOps)
}

func (backend *Backend) watch() {
	events, err := backend.store.WatchTree(backend.getKey("/nodes/"), backend.stop)
	if err != nil {
		log.Error("Error during watch:", err)
		return
	}

	for {
		select {
		case <-backend.stop:
			break
		case <-events:
			err := backend.syncMappings()
			if err != nil {
				log.Error("Error during watch:", err)
			}
		}
	}
}

// GetMapping which will either be a mapping in the current network or nil and the bool will be false, indicating that the mapping does not exist.
func (backend *Backend) GetMapping(ip uint32) (*common.Mapping, bool) {
	mapping, ok := backend.mappings[ip]
	return mapping, ok
}

// Init the backend datastore.
func (backend *Backend) Init() error {
	err := backend.lock()
	if err != nil {
		log.Error("Error acquiring backend store lock:", err)
		return err
	}
	err = backend.fetchNetworkConfig()
	if err != nil {
		log.Error("Error getting network configuration:", err)
		return err
	}
	err = backend.syncMappings()
	if err != nil {
		log.Error("Error synchronizing the network mappings with the backend store:", err)
		return err
	}
	err = backend.handleLocalMapping()
	if err != nil {
		log.Error("Error handling the local node mapping:", err)
		return err
	}
	err = backend.unlock()
	if err != nil {
		log.Error("Error releasing backend store lock:", err)
	}
	return err
}

// Start watching the backend and updating mappings.
func (backend *Backend) Start() {
	refresh := time.NewTicker(backend.cfg.RefreshInterval * time.Second)
	sync := time.NewTicker(backend.cfg.SyncInterval * time.Second)
	key := path.Join("/nodes/", backend.cfg.MachineID)

	go backend.watch()
	go func() {
		for {
			select {
			case <-backend.stop:
				break
			case <-refresh.C:
				err := backend.set(key, backend.localMapping.Bytes(), backend.NetworkCfg.LeaseTime)
				if err != nil {
					log.Error("Error during refresh of the ip address lease:", err)
				}
			case <-sync.C:
				err := backend.syncMappings()
				if err != nil {
					log.Error("Error during resync of the ip address mappings:", err)
				}
			}
		}
	}()
}

// Stop watching the backend and updating mappings
func (backend *Backend) Stop() {
	go func() {
		backend.stop <- struct{}{}
	}()
}

// New Backend object
func New(cfg *config.Config) (*Backend, error) {
	var backendStore store.Store

	storeCfg, err := generateStoreConfig(cfg)
	if err != nil {
		log.Error("Error generating the backend store configuration:", err)
		return nil, err
	}

	switch cfg.Datastore {
	case consulStore:
		backendStore, err = libkv.NewStore(store.CONSUL, cfg.Endpoints, storeCfg)
	case etcdStore:
		backendStore, err = libkv.NewStore(store.ETCD, cfg.Endpoints, storeCfg)
	default:
		log.Error("Configured 'Datastore' is not supported by quantum.")
		return nil, errors.New("Configured 'Datastore' is not supported by quantum.")
	}

	if err != nil {
		log.Error("Error connecting to the backend store:", err)
		return nil, err
	}

	return &Backend{
		store:    backendStore,
		cfg:      cfg,
		mappings: make(map[uint32]*common.Mapping),
		stop:     make(chan struct{}),
	}, nil
}
