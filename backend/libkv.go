package backend

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"github.com/Supernomad/quantum/common"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/etcd"
	"io/ioutil"
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
	mockStore   string        = "mock"
	lockTTL     time.Duration = 20 * time.Second
)

// Libkv datastore object which is responsible for managing state between the local node and the real libkv datastore.
type Libkv struct {
	log *common.Logger

	store  store.Store
	locker store.Locker

	cfg        *common.Config
	NetworkCfg *common.NetworkConfig

	localMapping *common.Mapping
	mappings     map[uint32]*common.Mapping

	stop chan struct{}
}

func generateStoreConfig(cfg *common.Config) (*store.Config, error) {
	storeCfg := &store.Config{PersistConnection: true}

	if cfg.AuthEnabled {
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
			storeCfg.TLS.BuildNameToCertificate()
		}
		if cfg.TLSSkipVerify {
			storeCfg.TLS.InsecureSkipVerify = true
		} else if cfg.TLSCA != "" {
			cert, err := ioutil.ReadFile(cfg.TLSCA)
			if err != nil {
				return nil, err
			}
			storeCfg.TLS.RootCAs = x509.NewCertPool()
			storeCfg.TLS.RootCAs.AppendCertsFromPEM(cert)
			storeCfg.TLS.BuildNameToCertificate()
		}
	}

	return storeCfg, nil
}

func (libkv *Libkv) getKey(key string) string {
	return path.Join(libkv.cfg.Prefix, key)
}

func (libkv *Libkv) lock() error {
	stopWaiting := make(chan struct{})
	ticker := time.NewTicker(lockTTL)
	locker, err := libkv.store.NewLock(libkv.getKey("/lock"), &store.LockOptions{Value: []byte(libkv.cfg.MachineID), TTL: lockTTL})
	if err != nil {
		return err
	}

	go func() {
		<-ticker.C
		stopWaiting <- struct{}{}
		ticker.Stop()
		close(stopWaiting)
	}()

	_, err = locker.Lock(stopWaiting)
	if err != nil {
		return err
	}
	libkv.locker = locker
	return nil
}

func (libkv *Libkv) unlock() error {
	err := libkv.locker.Unlock()
	if err != nil {
		return err
	}
	libkv.locker = nil
	return nil
}

func (libkv *Libkv) getMappingIfExists() (*common.Mapping, bool) {
	node, err := libkv.store.Get(libkv.getKey("/nodes/" + libkv.cfg.MachineID))
	if err != nil {
		return nil, false
	}
	mapping, err := common.ParseMapping(node.Value, libkv.cfg.PrivateKey)
	if err != nil {
		return nil, false
	}
	return mapping, true
}

func (libkv *Libkv) getFreeIP() (string, error) {
	for ip := libkv.NetworkCfg.BaseIP.Mask(libkv.NetworkCfg.IPNet.Mask); libkv.NetworkCfg.IPNet.Contains(ip); common.IncrementIP(ip) {
		if ip[3] == 0 || libkv.NetworkCfg.ReservedIPNet.Contains(ip) {
			continue
		}

		str := ip.String()
		if _, exists := libkv.mappings[common.IPtoInt(str)]; !exists {
			return str, nil
		}
	}
	return "", errors.New("There are no available ip addresses in the configured network.")
}

func (libkv *Libkv) handleLocalMapping() error {
	if libkv.cfg.PrivateIP == "" {
		if mapping, ok := libkv.getMappingIfExists(); ok {
			libkv.cfg.PrivateIP = mapping.PrivateIP
		} else {
			ip, err := libkv.getFreeIP()
			if err != nil {
				return err
			}
			libkv.cfg.PrivateIP = ip
		}
	}

	mapping := common.NewMapping(libkv.cfg.PrivateIP, libkv.cfg.PublicAddress, libkv.cfg.MachineID, libkv.cfg.PublicKey)
	key := path.Join("/nodes/", libkv.cfg.MachineID)

	err := libkv.set(key, mapping.Bytes(), libkv.NetworkCfg.LeaseTime)
	if err != nil {
		return err
	}
	libkv.localMapping = mapping
	return nil
}

func (libkv *Libkv) syncMappings(nodes []*store.KVPair) error {
	var err error
	if nodes == nil {
		nodes, err = libkv.store.List(libkv.getKey("/nodes/"))
		if err != nil {
			if err != store.ErrKeyNotFound {
				return err
			}
			nodes = make([]*store.KVPair, 0)
		}
	}

	mappings := make(map[uint32]*common.Mapping)
	for _, node := range nodes {
		mapping, err := common.ParseMapping(node.Value, libkv.cfg.PrivateKey)
		if err != nil {
			return err
		}
		mappings[common.IPtoInt(mapping.PrivateIP)] = mapping
	}
	libkv.mappings = mappings
	return nil
}

func (libkv *Libkv) fetchNetworkConfig() error {
	netCfg, err := libkv.store.Get(libkv.getKey("config"))
	if err != nil {
		if err != store.ErrKeyNotFound {
			return err
		}
		libkv.set("config", common.DefaultNetworkConfig.Bytes(), 0)
		libkv.NetworkCfg = common.DefaultNetworkConfig
		libkv.cfg.NetworkConfig = common.DefaultNetworkConfig
		return nil
	}
	networkCfg, err := common.ParseNetworkConfig(netCfg.Value)
	if err != nil {
		return err
	}
	libkv.NetworkCfg = networkCfg
	libkv.cfg.NetworkConfig = networkCfg
	return nil
}

func (libkv *Libkv) set(key string, value []byte, leaseTime time.Duration) error {
	var writeOps *store.WriteOptions
	if leaseTime > 0 {
		writeOps = &store.WriteOptions{TTL: leaseTime * time.Second}
	}
	return libkv.store.Put(
		libkv.getKey(key),
		value,
		writeOps)
}

// GetMapping which will either be a mapping in the current network or nil and the bool will be false, indicating that the mapping does not exist.
func (libkv *Libkv) GetMapping(ip uint32) (*common.Mapping, bool) {
	mapping, ok := libkv.mappings[ip]
	return mapping, ok
}

// Init the libkv datastore.
func (libkv *Libkv) Init() error {
	err := libkv.lock()
	if err != nil {
		return err
	}
	err = libkv.fetchNetworkConfig()
	if err != nil {
		return err
	}
	err = libkv.syncMappings(nil)
	if err != nil {
		return err
	}
	err = libkv.handleLocalMapping()
	if err != nil {
		return err
	}
	return libkv.unlock()
}

// Start watching the libkv and updating mappings.
func (libkv *Libkv) Start() {
	refresh := time.NewTicker(libkv.cfg.RefreshInterval)
	sync := time.NewTicker(libkv.cfg.SyncInterval)
	key := path.Join("/nodes/", libkv.cfg.MachineID)

	stopWatching := make(chan struct{})
	events, err := libkv.store.WatchTree(libkv.getKey("/nodes/"), stopWatching)
	if err != nil {
		libkv.log.Error.Println("[BACKEND]", "Error setting up watch:", err)
	}

	go func() {
		for {
			select {
			case stop := <-libkv.stop:
				stopWatching <- stop
				break
			case <-refresh.C:
				err := libkv.set(key, libkv.localMapping.Bytes(), libkv.NetworkCfg.LeaseTime)
				if err != nil {
					libkv.log.Error.Println("[BACKEND]", "Error during refresh of the ip address lease:", err)
				}
			case <-sync.C:
				err := libkv.syncMappings(nil)
				if err != nil {
					libkv.log.Error.Println("[BACKEND]", "Error during resync of the ip address mappings:", err)
				}
			case nodes := <-events:
				err := libkv.syncMappings(nodes)
				if err != nil {
					libkv.log.Error.Println("[BACKEND]", "Error while watching of the ip address mappings for changes:", err)
				}
			}
		}
	}()
}

// Stop watching the libkv and updating mappings
func (libkv *Libkv) Stop() {
	libkv.stop <- struct{}{}
}

// New Libkv object
func newLibkv(log *common.Logger, cfg *common.Config) (Backend, error) {
	var libkvStore store.Store

	storeCfg, err := generateStoreConfig(cfg)
	if err != nil {
		return nil, err
	}

	switch cfg.Datastore {
	case consulStore:
		libkvStore, err = libkv.NewStore(store.CONSUL, cfg.Endpoints, storeCfg)
	case etcdStore:
		libkvStore, err = libkv.NewStore(store.ETCD, cfg.Endpoints, storeCfg)
	default:
		err = errors.New("Configured 'Datastore' is not supported by quantum.")
	}

	if err != nil {
		return nil, err
	}

	return &Libkv{
		log:      log,
		store:    libkvStore,
		cfg:      cfg,
		mappings: make(map[uint32]*common.Mapping),
		stop:     make(chan struct{}),
	}, nil
}
