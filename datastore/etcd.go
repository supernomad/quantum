// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package datastore

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net/http"
	"path"
	"time"

	"github.com/Supernomad/quantum/common"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

// Etcd datastore struct for interacting with the coreos etcd key/value datastore.
type Etcd struct {
	log                 *common.Logger
	cfg                 *common.Config
	localMapping        *common.Mapping
	mappings            map[uint32]*common.Mapping
	cli                 client.Client
	kapi                client.KeysAPI
	watchIndex          uint64
	stopSyncing         chan struct{}
	stopRefreshingLease chan struct{}
	stopWatchingNodes   chan struct{}
}

func isError(err error, code int) bool {
	if err == nil {
		return false
	}

	if cErr, ok := err.(client.Error); ok {
		return cErr.Code == code
	}
	return false
}

func (etcd *Etcd) handleNetworkConfig() error {
	resp, err := etcd.kapi.Get(context.Background(), etcd.key("config"), &client.GetOptions{})
	if err != nil {
		if !isError(err, client.ErrorCodeKeyNotFound) {
			return errors.New("error retrieving the network configuration from etcd: " + err.Error())
		}

		etcd.log.Warn.Println("[ETCD]", "Using default network configuration.")
		_, err := etcd.kapi.Set(context.Background(), etcd.key("config"), common.DefaultNetworkConfig.String(), &client.SetOptions{})
		if err != nil {
			return errors.New("error setting the default network configuration in etcd: " + err.Error())
		}

		etcd.cfg.NetworkConfig = common.DefaultNetworkConfig
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

	opts := &client.SetOptions{
		TTL: etcd.cfg.NetworkConfig.LeaseTime,
	}
	_, err = etcd.kapi.Set(context.Background(), etcd.key("nodes", etcd.cfg.PrivateIP.String()), mapping.String(), opts)
	if err != nil {
		return errors.New("error setting the local network mapping in etcd: " + err.Error())
	}
	etcd.localMapping = mapping
	etcd.stopRefreshingLease = etcd.refresh(etcd.key("nodes", etcd.cfg.PrivateIP.String()), "", etcd.cfg.NetworkConfig.LeaseTime, etcd.cfg.DatastoreRefreshInterval)
	return nil
}

func (etcd *Etcd) key(str ...string) string {
	strs := []string{etcd.cfg.DatastorePrefix}
	return path.Join(append(strs, str...)...)
}

func (etcd *Etcd) lock() (chan struct{}, error) {
	opts := &client.SetOptions{
		PrevExist: client.PrevNoExist,
		TTL:       lockTTL,
	}

	for {
		_, err := etcd.kapi.Set(context.Background(), etcd.key("lock"), etcd.cfg.MachineID, opts)

		if err != nil && !isError(err, client.ErrorCodeNodeExist) {
			return nil, errors.New("error retrieving the lock on etcd: " + err.Error())
		} else if isError(err, client.ErrorCodeNodeExist) {
			time.Sleep(lockTTL)
			continue
		}

		break
	}

	stopRefreshing := etcd.refresh(etcd.key("lock"), etcd.cfg.MachineID, lockTTL, 5*time.Second)
	return stopRefreshing, nil
}

func (etcd *Etcd) refresh(key, value string, ttl, refreshInterval time.Duration) chan struct{} {
	stop := make(chan struct{})

	go func() {
		ticker := time.NewTicker(refreshInterval)

		opts := &client.SetOptions{
			PrevValue: value,
			PrevExist: client.PrevExist,
			TTL:       ttl,
			Refresh:   true,
		}

	loop:
		for {
			select {
			case <-stop:
				break loop
			case <-ticker.C:
				_, err := etcd.kapi.Set(context.Background(), key, "", opts)
				if err != nil {
					etcd.log.Error.Println("[ETCD]", "Error refreshing key in etcd: "+err.Error())
					if isError(err, client.ErrorCodeKeyNotFound) ||
						isError(err, client.ErrorCodePrevValueRequired) ||
						isError(err, client.ErrorCodeTestFailed) {
						break loop
					}
				}
			}
		}

		close(stop)
		ticker.Stop()
	}()

	return stop
}

func (etcd *Etcd) sync() error {
	var nodes client.Nodes
	resp, err := etcd.kapi.Get(context.Background(), etcd.key("nodes"), &client.GetOptions{Recursive: true})

	if err != nil {
		if !isError(err, client.ErrorCodeKeyNotFound) {
			return errors.New("error retrieving the mapping list from etcd: " + err.Error())
		}
		nodes = make(client.Nodes, 0)
	} else {
		nodes = resp.Node.Nodes
	}

	mappings := make(map[uint32]*common.Mapping)
	for _, node := range nodes {
		mapping, err := common.ParseMapping(node.Value, etcd.cfg)
		if err != nil {
			return errors.New("error parsing a mapping retrieved from etcd: " + err.Error())
		}
		mappings[common.IPtoInt(mapping.PrivateIP)] = mapping
	}
	etcd.mappings = mappings
	return nil
}

func (etcd *Etcd) unlock(stopRefreshing chan struct{}) error {
	stopRefreshing <- struct{}{}

	opts := &client.DeleteOptions{
		PrevValue: etcd.cfg.MachineID,
	}

	_, err := etcd.kapi.Delete(context.Background(), etcd.key("lock"), opts)

	if err != nil &&
		!isError(err, client.ErrorCodeKeyNotFound) &&
		!isError(err, client.ErrorCodePrevValueRequired) &&
		!isError(err, client.ErrorCodeTestFailed) {
		return errors.New("error releasing the etcd lock: " + err.Error())
	}
	return nil
}

func (etcd *Etcd) watch() {
	go func() {
	watch:
		opts := &client.WatcherOptions{
			AfterIndex: etcd.watchIndex,
			Recursive:  true,
		}
		watcher := etcd.kapi.Watcher(etcd.key("nodes"), opts)

	loop:
		for {
			select {
			case <-etcd.stopWatchingNodes:
				break loop
			default:
				resp, err := watcher.Next(context.Background())
				if err != nil {
					etcd.log.Error.Println("[ETCD]", "Error during watch on the etcd cluster: "+err.Error())
					time.Sleep(5 * time.Second)
					goto watch
				}

				etcd.watchIndex = resp.Index
				nodes := resp.Node.Nodes

				switch resp.Action {
				case "set", "update", "create":
					for _, node := range nodes {
						mapping, err := common.ParseMapping(node.Value, etcd.cfg)
						if err != nil {
							etcd.log.Error.Println("[ETCD]", "Error parsing mapping: "+err.Error())
							continue
						}
						etcd.mappings[common.IPtoInt(mapping.PrivateIP)] = mapping
					}
				case "delete", "expire":
					for _, node := range nodes {
						mapping, err := common.ParseMapping(node.Value, etcd.cfg)
						if err != nil {
							etcd.log.Error.Println("[ETCD]", "Error parsing mapping: "+err.Error())
							continue
						}
						delete(etcd.mappings, common.IPtoInt(mapping.PrivateIP))
					}
				}
			}
		}
		close(etcd.stopWatchingNodes)
	}()
}

// Mapping returns a mapping and true based on the supplied uint32 representation of an ipv4 address if it exists within the datastore, otherwise it returns nil for the mapping and false.
func (etcd *Etcd) Mapping(ip uint32) (*common.Mapping, bool) {
	mapping, exists := etcd.mappings[ip]
	return mapping, exists
}

// Init the Etcd datastore which will open any necessary connections, preform an initial sync of the datastore, and define the local mapping in the datastore.
func (etcd *Etcd) Init() error {
	stopRefreshing, err := etcd.lock()
	if err != nil {
		return err
	}

	err = etcd.handleNetworkConfig()
	if err != nil {
		etcd.unlock(stopRefreshing)
		return err
	}

	err = etcd.sync()
	if err != nil {
		etcd.unlock(stopRefreshing)
		return err
	}

	err = etcd.handleLocalMapping()
	if err != nil {
		etcd.unlock(stopRefreshing)
		return err
	}

	return etcd.unlock(stopRefreshing)
}

// Start periodic synchronization, and DHCP lease refresh with the datastore, as well as start watching for changes in network topology.
func (etcd *Etcd) Start() {
	etcd.watch()

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
					etcd.log.Error.Println("[ETCD]", "Error synchronizing mappings with the backend: "+err.Error())
				}
			}
		}
		close(etcd.stopSyncing)
		ticker.Stop()
	}()
}

// Stop synchronizing with the backend and shutdown open connections.
func (etcd *Etcd) Stop() {
	go func() {
		etcd.stopSyncing <- struct{}{}
		etcd.stopWatchingNodes <- struct{}{}
		etcd.stopRefreshingLease <- struct{}{}
	}()
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

func newEtcd(log *common.Logger, cfg *common.Config) (Datastore, error) {
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
		log:      log,
		cfg:      cfg,
		mappings: make(map[uint32]*common.Mapping),
		cli:      cli,
		kapi:     kapi,
	}, nil
}
