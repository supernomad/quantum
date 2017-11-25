// Copyright (c) 2016-2017 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package datastore

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net/http"
	"path"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/supernomad/quantum/common"
	"golang.org/x/net/context"
)

// Etcd datastore struct for interacting with the coreos etcd key/value datastore.
type Etcd struct {
	cfg                 *common.Config
	mappings            map[uint32]*common.Mapping
	gwMappings          map[uint32]*common.Mapping
	ctx                 context.Context
	cli                 client.Client
	kapi                client.KeysAPI
	watchIndex          uint64
	stopSyncing         chan struct{}
	stopRefreshingLock  chan struct{}
	stopRefreshingLease chan struct{}
	stopWatchingNodes   chan struct{}
}

func isError(err error, codes ...int) bool {
	if err == nil {
		return false
	}

	if etcdError, ok := err.(client.Error); ok {
		for i := 0; i < len(codes); i++ {
			if etcdError.Code == codes[i] {
				return true
			}
		}
	}

	return false
}

func (etcd *Etcd) key(strs ...string) string {
	strs = append([]string{etcd.cfg.DatastorePrefix}, strs...)
	return path.Join(strs...)
}

func (etcd *Etcd) handleNetworkConfig() error {
	key := etcd.key("config")
	resp, err := etcd.kapi.Get(etcd.ctx, key, &client.GetOptions{})

	if err != nil && !isError(err, client.ErrorCodeKeyNotFound) {
		return errors.New("error retrieving the network configuration from etcd: " + err.Error())
	} else if isError(err, client.ErrorCodeKeyNotFound) {
		_, err := etcd.kapi.Set(etcd.ctx, key, etcd.cfg.NetworkConfig.String(), &client.SetOptions{})
		if err != nil {
			return errors.New("error setting the default network configuration in etcd: " + err.Error())
		}
		return nil
	}

	networkCfg, err := common.ParseNetworkConfig([]byte(resp.Node.Value))
	if err != nil {
		return errors.New("error parsing the network configuration retrieved from etcd: " + err.Error())
	}

	etcd.cfg.NetworkConfig = networkCfg
	return nil
}

func (etcd *Etcd) handleLocalMapping() error {
	mapping, err := common.GenerateLocalMapping(etcd.cfg, etcd.mappings)
	if err != nil {
		return errors.New("error generating the local network mapping: " + err.Error())
	}

	key := etcd.key("nodes", etcd.cfg.PrivateIP.String())
	opts := &client.SetOptions{
		TTL: etcd.cfg.NetworkConfig.LeaseTime,
	}

	_, err = etcd.kapi.Set(etcd.ctx, key, mapping.String(), opts)
	if err != nil {
		return errors.New("error setting the local network mapping in etcd: " + err.Error())
	}

	go etcd.refresh(key, "", etcd.cfg.NetworkConfig.LeaseTime, etcd.cfg.DatastoreRefreshInterval, etcd.stopRefreshingLease)

	return nil
}

func (etcd *Etcd) lockFloatingIP(key, value string) {
	opts := &client.SetOptions{
		PrevExist: client.PrevNoExist,
		TTL:       etcd.cfg.DatastoreFloatingIPTTL,
	}

	stop := make(chan struct{})
	for {
		_, err := etcd.kapi.Set(etcd.ctx, key, value, opts)

		if err != nil && !isError(err, client.ErrorCodeNodeExist) {
			etcd.cfg.Log.Error.Println("[ETCD]", "Error attempting to set floating mapping in etcd: "+err.Error())
			time.Sleep(etcd.cfg.DatastoreFloatingIPTTL)
			continue
		} else if isError(err, client.ErrorCodeNodeExist) {
			time.Sleep(etcd.cfg.DatastoreFloatingIPTTL)
			continue
		}

		etcd.refresh(key, value, etcd.cfg.DatastoreFloatingIPTTL, etcd.cfg.DatastoreFloatingIPTTL/2, stop)
	}
}

func (etcd *Etcd) handleFloatingMappings() error {
	for i := 0; i < len(etcd.cfg.FloatingIPs); i++ {
		mapping, err := common.GenerateFloatingMapping(etcd.cfg, i, etcd.mappings)
		if err != nil {
			return err
		}

		go etcd.lockFloatingIP(etcd.key("nodes", mapping.PrivateIP.String()), mapping.String())
	}

	return nil
}

func (etcd *Etcd) refresh(key, value string, ttl, refreshInterval time.Duration, stop chan struct{}) {
	ticker := time.NewTicker(refreshInterval)

	opts := &client.SetOptions{
		PrevValue: value,
		PrevExist: client.PrevExist,
		TTL:       ttl,
		Refresh:   true,
	}

	stopRefreshing := false
	for !stopRefreshing {
		select {
		case <-stop:
			stopRefreshing = true
		case <-ticker.C:
			_, err := etcd.kapi.Set(etcd.ctx, key, "", opts)
			if err != nil {
				etcd.cfg.Log.Error.Println("[ETCD]", "Error refreshing key in etcd: "+err.Error())
				if isError(err, client.ErrorCodeKeyNotFound, client.ErrorCodePrevValueRequired, client.ErrorCodeTestFailed) {
					stopRefreshing = true
				}
			}
		}
	}

	ticker.Stop()
}

func (etcd *Etcd) lock() error {
	key := etcd.key("lock")
	opts := &client.SetOptions{
		PrevExist: client.PrevNoExist,
		TTL:       lockTTL,
	}

	for {
		_, err := etcd.kapi.Set(etcd.ctx, key, etcd.cfg.MachineID, opts)

		if err != nil && !isError(err, client.ErrorCodeNodeExist) {
			return errors.New("error retrieving the lock on etcd: " + err.Error())
		} else if isError(err, client.ErrorCodeNodeExist) {
			time.Sleep(lockTTL)
			continue
		}

		break
	}

	go etcd.refresh(key, etcd.cfg.MachineID, lockTTL, lockTTL/2, etcd.stopRefreshingLock)
	return nil
}

func (etcd *Etcd) unlock() error {
	etcd.stopRefreshingLock <- struct{}{}

	key := etcd.key("lock")
	opts := &client.DeleteOptions{
		PrevValue: etcd.cfg.MachineID,
	}

	_, err := etcd.kapi.Delete(etcd.ctx, key, opts)

	if err != nil && !isError(err, client.ErrorCodeKeyNotFound, client.ErrorCodePrevValueRequired, client.ErrorCodeTestFailed) {
		return errors.New("error releasing the etcd lock: " + err.Error())
	}

	return nil
}

func (etcd *Etcd) sync() error {
	var nodes client.Nodes
	resp, err := etcd.kapi.Get(etcd.ctx, etcd.key("nodes"), &client.GetOptions{Recursive: true})

	if err != nil {
		if !isError(err, client.ErrorCodeKeyNotFound) {
			return errors.New("error retrieving the mapping list from etcd: " + err.Error())
		}
		nodes = make(client.Nodes, 0)
	} else {
		nodes = resp.Node.Nodes
	}

	mappings := make(map[uint32]*common.Mapping)
	gwMappings := make(map[uint32]*common.Mapping)
	for _, node := range nodes {
		mapping, err := common.ParseMapping(node.Value, etcd.cfg)
		if err != nil {
			return errors.New("error parsing a mapping retrieved from etcd: " + err.Error())
		}
		mappings[common.IPtoInt(mapping.PrivateIP)] = mapping
		if mapping.Gateway {
			etcd.cfg.Log.Info.Println("[ETCD]", "Got gateway mapping.")
			gwMappings[common.IPtoInt(mapping.PrivateIP)] = mapping
		}
	}

	etcd.mappings = mappings
	etcd.gwMappings = gwMappings
	return nil
}

func (etcd *Etcd) watch() {
	opts := &client.WatcherOptions{
		AfterIndex: etcd.watchIndex,
		Recursive:  true,
	}
	watcher := etcd.kapi.Watcher(etcd.key("nodes"), opts)

	stopWatching := false
	for !stopWatching {
		select {
		case <-etcd.stopWatchingNodes:
			stopWatching = true
		default:
			resp, err := watcher.Next(etcd.ctx)
			if err != nil {
				etcd.cfg.Log.Error.Println("[ETCD]", "Error during watch on the etcd cluster: "+err.Error())
				time.Sleep(5 * time.Second)

				go etcd.watch()
				return
			}

			etcd.watchIndex = resp.Index
			nodes := resp.Node.Nodes

			switch resp.Action {
			case "set", "update", "create":
				for _, node := range nodes {
					mapping, err := common.ParseMapping(node.Value, etcd.cfg)
					if err != nil {
						etcd.cfg.Log.Error.Println("[ETCD]", "Error parsing mapping: "+err.Error())
						continue
					}
					etcd.mappings[common.IPtoInt(mapping.PrivateIP)] = mapping
					if mapping.Gateway {
						etcd.gwMappings[common.IPtoInt(mapping.PrivateIP)] = mapping
					}
				}
			case "delete", "expire":
				for _, node := range nodes {
					mapping, err := common.ParseMapping(node.Value, etcd.cfg)
					if err != nil {
						etcd.cfg.Log.Error.Println("[ETCD]", "Error parsing mapping: "+err.Error())
						continue
					}
					delete(etcd.mappings, common.IPtoInt(mapping.PrivateIP))
					if mapping.Gateway {
						delete(etcd.gwMappings, common.IPtoInt(mapping.PrivateIP))
					}
				}
			}
		}
	}
}

// Mapping returns a mapping and true based on the supplied uint32 representation of an ipv4 address if it exists within the datastore, otherwise it returns nil for the mapping and false.
func (etcd *Etcd) Mapping(ip uint32) (*common.Mapping, bool) {
	mapping, exists := etcd.mappings[ip]
	return mapping, exists
}

// GatewayMapping should retun the mapping and true if it exists specifically for destinations outside of the quantum network, if the mapping doesn't exist it will return nil and false.
func (etcd *Etcd) GatewayMapping(ip uint32) (*common.Mapping, bool) {
	for _, mapping := range etcd.gwMappings {
		return mapping, true
	}
	return nil, false
}

// Init the Etcd datastore which will open any necessary connections, preform an initial sync of the datastore, and define the local mapping in the datastore.
func (etcd *Etcd) Init() error {
	err := etcd.lock()
	if err != nil {
		return err
	}

	err = etcd.handleNetworkConfig()
	if err != nil {
		etcd.unlock()
		return err
	}

	err = etcd.sync()
	if err != nil {
		etcd.unlock()
		return err
	}

	err = etcd.handleLocalMapping()
	if err != nil {
		etcd.unlock()
		return err
	}

	err = etcd.handleFloatingMappings()
	if err != nil {
		etcd.unlock()
		return err
	}

	return etcd.unlock()
}

// Start periodic synchronization, and DHCP lease refresh with the datastore, as well as start watching for changes in network topology.
func (etcd *Etcd) Start() {
	go etcd.watch()

	ticker := time.NewTicker(etcd.cfg.DatastoreSyncInterval)
	go func() {
	loop:
		for {
			select {
			case <-etcd.stopSyncing:
				break loop
			case <-ticker.C:
				err := etcd.sync()
				if err != nil {
					etcd.cfg.Log.Error.Println("[ETCD]", "Error synchronizing mappings with the backend: "+err.Error())
				}
			}
		}

		ticker.Stop()
	}()
}

// Stop synchronizing with the backend and shutdown open connections.
func (etcd *Etcd) Stop() {
	etcd.stopSyncing <- struct{}{}
	etcd.stopRefreshingLease <- struct{}{}
	etcd.stopWatchingNodes <- struct{}{}

	close(etcd.stopSyncing)
	close(etcd.stopRefreshingLock)
	close(etcd.stopRefreshingLease)
	close(etcd.stopWatchingNodes)
}

func generateConfig(cfg *common.Config) (client.Config, error) {
	endpointPrefix := "http://"
	etcdCfg := client.Config{}

	if cfg.AuthEnabled {
		etcdCfg.Username = cfg.DatastoreUsername
		etcdCfg.Password = cfg.DatastorePassword
	}

	if cfg.TLSEnabled {
		tlsCfg := &tls.Config{}
		endpointPrefix = "https://"

		if cfg.DatastoreTLSKey != "" && cfg.DatastoreTLSCert != "" {
			cert, err := tls.LoadX509KeyPair(cfg.DatastoreTLSCert, cfg.DatastoreTLSKey)
			if err != nil {
				return etcdCfg, errors.New("error reading the supplied tls certificate and/or key: " + err.Error())
			}
			tlsCfg.Certificates = []tls.Certificate{cert}
			tlsCfg.BuildNameToCertificate()
		}

		tlsCfg.InsecureSkipVerify = cfg.DatastoreTLSSkipVerify

		if cfg.DatastoreTLSCA != "" {
			cert, err := ioutil.ReadFile(cfg.DatastoreTLSCA)
			if err != nil {
				return etcdCfg, errors.New("error reading the supplied tls ca certificate: " + err.Error())
			}
			tlsCfg.RootCAs = x509.NewCertPool()
			tlsCfg.RootCAs.AppendCertsFromPEM(cert)
			tlsCfg.BuildNameToCertificate()
		}

		etcdCfg.Transport = &http.Transport{
			TLSClientConfig: tlsCfg,
		}
	}

	etcdCfg.Endpoints = make([]string, len(cfg.DatastoreEndpoints))
	for i := 0; i < len(cfg.DatastoreEndpoints); i++ {
		etcdCfg.Endpoints[i] = endpointPrefix + cfg.DatastoreEndpoints[i]
	}

	return etcdCfg, nil
}

func newEtcd(cfg *common.Config) (Datastore, error) {
	etcdCfg, err := generateConfig(cfg)
	if err != nil {
		return nil, err
	}

	cli, err := client.New(etcdCfg)
	if err != nil {
		return nil, errors.New("error creating client connection to etcd: " + err.Error())
	}

	kapi := client.NewKeysAPI(cli)
	return &Etcd{
		ctx:                 context.TODO(),
		cfg:                 cfg,
		mappings:            make(map[uint32]*common.Mapping),
		gwMappings:          make(map[uint32]*common.Mapping),
		cli:                 cli,
		kapi:                kapi,
		stopSyncing:         make(chan struct{}),
		stopRefreshingLock:  make(chan struct{}),
		stopRefreshingLease: make(chan struct{}),
		stopWatchingNodes:   make(chan struct{}),
	}, nil
}
