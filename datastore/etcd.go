package datastore

import (
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/Supernomad/quantum/common"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

// Etcd datastore
type Etcd struct {
	log                 *common.Logger
	cfg                 *common.Config
	localMapping        *common.Mapping
	mappings            map[uint32]*common.Mapping
	cli                 client.Client
	kapi                client.KeysAPI
	watchIndex          uint64
	stopRefreshingLease chan struct{}
	stopWatchingNodes   chan struct{}
	wg                  *sync.WaitGroup
}

func isError(err error, code int) bool {
	if err == nil {
		return false
	}

	if cErr, ok := err.(Error); ok {
		return cErr.Code == code
	}
	return false
}

func (etcd *Etcd) handleNetworkConfig() error {
	resp, err := etcd.kapi.Get(context.Background(), etcd.key("config"), &client.GetOptions{})
	if err != nil {
		if !isError(err, client.ErrorCodeKeyNotFound) {
			return err
		}
		_, err := etcd.kapi.Set(context.Background(), etcd.key("config"), etcd.cfg.MachineID, &client.SetOptions{})
		if err != nil {
			return err
		}

		etcd.cfg.NetworkConfig = common.DefaultNetworkConfig
		return nil
	}

	networkCfg, err := common.ParseNetworkConfig([]byte(resp.Node.Value))
	if err != nil {
		return err
	}

	etcd.cfg.NetworkConfig = networkCfg
	return nil
}

func (etcd *Etcd) handleLocalMapping() error {
	mapping := common.GenerateLocalMapping(etcd.cfg, etcd.mappings)

	opts := &client.SetOptions{
		TTL:              etcd.cfg.NetworkConfig.LeaseTime,
		NoValueOnSuccess: true,
	}
	_, err := etcd.kapi.Set(context.Background(), etcd.key("nodes", etcd.cfg.MachineID), mapping.String(), opts)
	if err != nil {
		return err
	}
	etcd.localMapping = mapping
	etcd.stopRefreshingLease = etcd.refresh("nodes/"+etcd.cfg.MachineID, value, ttl, refreshInterval)
	return nil
}

func (etcd *Etcd) key(str ...string) string {
	return path.Join(etcd.cfg.Prefix, str...)
}

func (etcd *Etcd) lock() (chan struct{}, error) {
	opts := &client.SetOptions{
		PrevExist:        client.PrevNoExist,
		TTL:              lockTTL,
		NoValueOnSuccess: true,
	}

	for {
		_, err := etcd.kapi.Set(context.Background(), etcd.key("lock"), etcd.cfg.MachineID, opts)

		if err != nil && !isError(err, client.ErrorCodeNodeExist) {
			return nil, err
		} else if isError(err, client.ErrorCodeNodeExist) {
			time.Sleep(lockTTL)
			continue
		}

		break
	}

	stopRefreshing := etcd.refresh("lock", lockTTL, 5*time.Second)
	return stopRefreshing, nil
}

func (etcd *Etcd) refresh(key, value string, ttl, refreshInterval time.Duration) chan struct{} {
	stop := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)

		opts := &client.SetOptions{
			PrevValue:        value,
			PrevExist:        client.PrevExist,
			TTL:              ttl,
			NoValueOnSuccess: true,
			Refresh:          true,
		}

	loop:
		for {
			select {
			case <-stop:
				break loop
			case <-ticker.C:
				_, err := etcd.kapi.Set(context.Background(), etcd.key(key), "", opts)
				if err != nil {
					if isError(err, client.ErrKeyNotFound) ||
						isError(err, client.ErrorCodePrevValueRequired) ||
						isError(err, client.ErrorCodeTestFailed) {
						break loop
					}
					etcd.log.Error.Println("[ETCD]", "Error refreshing key in etcd:", err.Error())
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
			return err
		}
		nodes = make(client.Nodes, 0)
	} else {
		nodes = resp.Node.Nodes
	}

	mappings := make(map[uint32]*common.Mapping)
	for _, node := range nodes {
		mapping, err := common.ParseMapping(node.Value, etcd.cfg.PrivateKey)
		if err != nil {
			return err
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
		PrevExist: client.PrevExist,
	}

	_, err := etcd.kapi.Delete(context.Background(), etcd.key("lock"), opts)

	if err != nil &&
		!isError(err, client.ErrorCodeKeyNotFound) &&
		!isError(err, client.ErrorCodePrevValueRequired) &&
		!isError(err, client.ErrorCodeTestFailed) {
		return err
	}
	return nil
}

func (etcd *Etcd) watch() error {
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
					etcd.log.Error.Println("[ETCD]", "Error during watch on the etcd cluster:", err.Error())
					time.Sleep(10 * time.Second)
					goto watch
				}

				nodes := resp.Node.Nodes
				switch resp.Action {
				case "set", "update", "create":
					for _, node := range nodes {
						mapping, err := common.ParseMapping(node.Value, etcd.cfg.PrivateKey)
						if err != nil {
							etcd.log.Error.Println("[ETCD]", "Error deserializing mapping:", err.Error())
						}
						etcd.mappings[common.IPtoInt(mapping.PrivateIP)] = mapping
					}
				case "delete", "expire":
					for _, node := range nodes {
						mapping, err := common.ParseMapping(node.Value, etcd.cfg.PrivateKey)
						if err != nil {
							etcd.log.Error.Println("[ETCD]", "Error deserializing mapping:", err.Error())
						}
						delete(etcd.mappings, common.IPtoInt(mappings.PrivateIP))
					}
				}
			}
		}
		close(etcd.stopWatchingNodes)
	}()
}

// GetMapping from the Etcd datastore
func (etcd *Etcd) Mapping(ip uint32) (*common.Mapping, bool) {
	mapping, exists := etcd.mappings[ip]
	return mapping, exists
}

// Init the Etcd datastore
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

// Start the Etcd datastore
func (etcd *Etcd) Start(wg *sync.WaitGroup) {
	etcd.wg = wg
}

// Stop the Etcd datastore
func (etcd *Etcd) Stop() {
	go func() {
		etcd.stopWatchingNodes <- struct{}{}
		etcd.stopRefreshingLease <- struct{}{}
		etcd.wg.Done()
	}()
}

func generateConfig(cfg *common.Config) client.Config {
	endpointPrefix := "http://"
	etcdCfg := client.Config{}

	if cfg.AuthEnabled {
		etcdCfg.Username = cfg.Username
		etcdCfg.Password = cfg.Password
	}

	if cfg.TLSEnabled {
		tlsCfg = &tls.Config{}
		endpointPrefix = "https://"

		if cfg.TLSKey != "" && cfg.TLSCert != "" {
			cert, err := tls.LoadX509KeyPair(cfg.TLSCert, cfg.TLSKey)
			if err != nil {
				return etcdCfg, err
			}
			tlsCfg.Certificates = []tls.Certificate{cert}
			tlsCfg.BuildNameToCertificate()
		}

		tlsCfg.InsecureSkipVerify = cfg.TLSSkipVerify

		if cfg.TLSCA != "" {
			cert, err := ioutil.ReadFile(cfg.TLSCA)
			if err != nil {
				return etcdCfg, err
			}
			tlsCfg.RootCAs = x509.NewCertPool()
			tlsCfg.RootCAs.AppendCertsFromPEM(cert)
			tlsCfg.BuildNameToCertificate()
		}

		etcdCfg.Transport = &http.Transport{
			TLSClientConfig: tlsCfg,
		}
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
		return nil, err
	}

	kapi := client.NewKeysAPI(cli)
	return &Etcd{
		log:      log,
		cfg:      cfg,
		mappings: make(map[uint32]*common.Mapping),
		cli:      cli,
		kapi:     kapi,
	}
}
