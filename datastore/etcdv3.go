// Copyright (c) 2016-2018 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package datastore

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"io/ioutil"
	"path"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"github.com/supernomad/quantum/common"
	"golang.org/x/net/context"
)

// EtcdV3 datastore struct for interacting with the coreos etcd key/value datastore using the v3 api.
type EtcdV3 struct {
	cfg         *common.Config
	etcdCfg     clientv3.Config
	mappings    map[uint32]*common.Mapping
	gateway     uint32
	stopSyncing chan struct{}
	cli         *clientv3.Client
	cliCtx      context.Context
	cliCancel   context.CancelFunc
}

func (etcd *EtcdV3) key(strs ...string) string {
	strs = append([]string{etcd.cfg.DatastorePrefix}, strs...)
	return path.Join(strs...)
}

func (etcd *EtcdV3) handleNetworkConfig() error {
	key := etcd.key("config")

	resp, err := etcd.cli.Txn(etcd.cliCtx).
		If(clientv3util.KeyMissing(key)).
		Then(clientv3.OpPut(key, etcd.cfg.NetworkConfig.String())).
		Else(clientv3.OpGet(key, clientv3.WithLimit(1))).
		Commit()

	if err != nil {
		return errors.New("error retrieving the network configuration from etcd: " + err.Error())
	}

	if resp.Succeeded {
		return nil
	}

	get := resp.Responses[0].GetResponseRange()
	networkCfg, err := common.ParseNetworkConfig(get.Kvs[0].Value)
	if err != nil {
		return errors.New("error parsing the network configuration retrieved from etcd: " + err.Error())
	}

	etcd.cfg.NetworkConfig = networkCfg
	return nil
}

func (etcd *EtcdV3) sync() error {
	key := etcd.key("nodes")

	resp, err := etcd.cli.Get(etcd.cliCtx, key, clientv3.WithPrefix())
	if err != nil {
		return errors.New("error retrieving the mapping list from etcd: " + err.Error())
	}

	mappings := make(map[uint32]*common.Mapping)

	for _, kv := range resp.Kvs {
		mapping, err := common.ParseMapping(string(kv.Value), etcd.cfg)
		if err != nil {
			return errors.New("error parsing a mapping retrieved from etcd: " + err.Error())
		}
		mappings[common.IPtoInt(mapping.PrivateIP)] = mapping
	}

	etcd.mappings = mappings

	return nil
}

func (etcd *EtcdV3) handleLocalMapping() error {
	mapping, err := common.GenerateLocalMapping(etcd.cfg, etcd.mappings)
	if err != nil {
		return errors.New("could not generate the local network mapping: " + err.Error())
	}

	key := etcd.key("nodes", etcd.cfg.PrivateIP.String())

	lease, err := etcd.lease(etcd.cfg.NetworkConfig.LeaseTime.Seconds() / time.Second.Seconds())
	if err != nil {
		return errors.New("could not lock private ip in etcd: " + err.Error())
	}

	_, err = etcd.cli.Put(etcd.cliCtx, key, mapping.String(), clientv3.WithLease(lease))
	if err != nil {
		return errors.New("could not update etcd with the local network mapping: " + err.Error())
	}

	err = etcd.keepalive(lease)
	if err != nil {
		return errors.New("coult not refresh local mapping in etcd: " + err.Error())
	}

	if mapping.Gateway != nil {
		etcd.gateway = binary.LittleEndian.Uint32(mapping.Gateway.To4())
	}

	return nil
}

func (etcd *EtcdV3) lockFloatingIP(key, value string) {
	first := true
	for {
		if !first {
			time.Sleep(etcd.cfg.DatastoreFloatingIPTTL)
		}
		first = false

		lease, err := etcd.lease(etcd.cfg.DatastoreFloatingIPTTL.Seconds() / time.Second.Seconds())
		if err != nil {
			etcd.cfg.Log.Error.Println("[ETCD]", "Error attempting to lock floating mapping in etcd: "+err.Error())
			continue
		}

		_, err = etcd.cli.Put(etcd.cliCtx, key, value)
		if err != nil {
			etcd.cfg.Log.Error.Println("[ETCD]", "Error attempting to set floating mapping in etcd: "+err.Error())
			continue
		}

		err = etcd.keepalive(lease)
		if err != nil {
			etcd.cfg.Log.Error.Println("[ETCD]", "Error attempting to refresh floating mapping in etcd: "+err.Error())
			continue
		}
	}
}

func (etcd *EtcdV3) handleFloatingMappings() error {
	for i := 0; i < len(etcd.cfg.FloatingIPs); i++ {
		mapping, err := common.GenerateFloatingMapping(etcd.cfg, i, etcd.mappings)
		if err != nil {
			return err
		}

		go etcd.lockFloatingIP(etcd.key("nodes", mapping.PrivateIP.String()), mapping.String())
	}

	return nil
}

func (etcd *EtcdV3) lease(ttl float64) (clientv3.LeaseID, error) {
	resp, err := etcd.cli.Grant(etcd.cliCtx, int64(ttl))
	if err != nil {
		return -1, errors.New("failed creating lease: " + err.Error())
	}

	return resp.ID, nil
}

func (etcd *EtcdV3) keepalive(lease clientv3.LeaseID) error {
	keepalives, err := etcd.cli.KeepAlive(etcd.cliCtx, lease)

	if err != nil {
		return errors.New("failed setting lease keepalive: " + err.Error())
	}

	// This is here because of an oddity in the clientv3 keep alive implementation.
	// See https://github.com/coreos/etcd/issues/7446
	go func(keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse) {
		for range keepalives {
		}
	}(keepalives)

	return nil
}

func (etcd *EtcdV3) lock() (clientv3.LeaseID, error) {
	key := etcd.key("lock")

	var err error
	var lease clientv3.LeaseID

	for {
		lease, err = etcd.lease(lockTTL.Seconds() / time.Second.Seconds())
		if err != nil {
			return -1, errors.New("could not lock etcd: " + err.Error())
		}

		txResp, err := etcd.cli.Txn(etcd.cliCtx).
			If(clientv3util.KeyMissing(key)).
			Then(clientv3.OpPut(key, etcd.cfg.MachineID, clientv3.WithLease(lease))).
			Commit()

		if err != nil {
			return -1, errors.New("error retrieving the lock on etcd: " + err.Error())
		}

		if txResp.Succeeded {
			break
		}

		time.Sleep(lockTTL + time.Second)
	}

	err = etcd.keepalive(lease)
	if err != nil {
		return -1, errors.New("could not refresh etcd lock: " + err.Error())
	}

	return lease, nil
}

func (etcd *EtcdV3) unlock(lease clientv3.LeaseID) error {
	_, err := etcd.cli.Revoke(etcd.cliCtx, lease)

	if err != nil {
		return errors.New("error revoking etcd lock: " + err.Error())
	}

	return nil
}

func (etcd *EtcdV3) watch() {
	key := etcd.key("nodes")
outer:
	for {
		watch := etcd.cli.Watch(etcd.cliCtx, key, clientv3.WithPrefix())
		for resp := range watch {
			if resp.Canceled {
				break outer
			}
			for _, ev := range resp.Events {
				switch ev.Type.String() {
				case "PUT":
					mapping, err := common.ParseMapping(string(ev.Kv.Value), etcd.cfg)
					if err != nil {
						etcd.cfg.Log.Error.Println("[ETCD]", "Error parsing mapping: "+err.Error())
						continue
					}
					etcd.mappings[common.IPtoInt(mapping.PrivateIP)] = mapping
				case "DELETE":
					mapping, err := common.ParseMapping(string(ev.Kv.Value), etcd.cfg)
					if err != nil {
						etcd.cfg.Log.Error.Println("[ETCD]", "Error parsing mapping: "+err.Error())
						continue
					}
					delete(etcd.mappings, common.IPtoInt(mapping.PrivateIP))
				}
			}
		}
	}
}

// Mapping returns a mapping and true based on the supplied uint32 representation of an ipv4 address if it exists within the datastore, otherwise it returns nil for the mapping and false.
func (etcd *EtcdV3) Mapping(ip uint32) (*common.Mapping, bool) {
	mapping, exists := etcd.mappings[ip]
	return mapping, exists
}

// GatewayMapping should retun the mapping and true if it exists specifically for destinations outside of the quantum network, if the mapping doesn't exist it will return nil and false.
func (etcd *EtcdV3) GatewayMapping() (*common.Mapping, bool) {
	mapping, exists := etcd.mappings[etcd.gateway]
	return mapping, exists
}

// Init the Etcd datastore which will open any necessary connections, preform an initial sync of the datastore, and define the local mapping in the datastore.
func (etcd *EtcdV3) Init() error {
	lease, err := etcd.lock()
	if err != nil {
		return err
	}

	err = etcd.handleNetworkConfig()
	if err != nil {
		etcd.unlock(lease)
		return err
	}

	err = etcd.sync()
	if err != nil {
		etcd.unlock(lease)
		return err
	}

	err = etcd.handleLocalMapping()
	if err != nil {
		etcd.unlock(lease)
		return err
	}

	err = etcd.handleFloatingMappings()
	if err != nil {
		etcd.unlock(lease)
		return err
	}

	return etcd.unlock(lease)
}

// Start periodic synchronization, and DHCP lease refresh with the datastore, as well as start watching for changes in network topology.
func (etcd *EtcdV3) Start() {
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
func (etcd *EtcdV3) Stop() {
	etcd.stopSyncing <- struct{}{}

	// Cancel all outstanding contexts and close the main client.
	etcd.cliCancel()
	etcd.cli.Close()

	close(etcd.stopSyncing)
}

func generateV3Config(ctx context.Context, cfg *common.Config) (clientv3.Config, error) {
	etcdCfg := clientv3.Config{
		DialTimeout:          5 * time.Second,
		DialKeepAliveTime:    10 * time.Second,
		DialKeepAliveTimeout: 5 * time.Second,
		Context:              ctx,
		RejectOldCluster:     true,
	}

	if cfg.AuthEnabled {
		etcdCfg.Username = cfg.DatastoreUsername
		etcdCfg.Password = cfg.DatastorePassword
	}

	if cfg.TLSEnabled {
		tlsCfg := &tls.Config{}

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

		etcdCfg.TLS = tlsCfg
	}

	etcdCfg.Endpoints = make([]string, len(cfg.DatastoreEndpoints))
	for i := 0; i < len(cfg.DatastoreEndpoints); i++ {
		etcdCfg.Endpoints[i] = cfg.DatastoreEndpoints[i]
	}

	return etcdCfg, nil
}

func newEtcdV3(cfg *common.Config) (Datastore, error) {
	ctx, cancel := context.WithCancel(context.Background())

	etcdCfg, err := generateV3Config(ctx, cfg)
	if err != nil {
		return nil, err
	}

	cli, err := clientv3.New(etcdCfg)
	if err != nil {
		return nil, errors.New("error connecting to etcd: " + err.Error())
	}

	return &EtcdV3{
		cfg:         cfg,
		etcdCfg:     etcdCfg,
		mappings:    make(map[uint32]*common.Mapping),
		stopSyncing: make(chan struct{}),
		cli:         cli,
		cliCtx:      ctx,
		cliCancel:   cancel,
	}, nil
}
