package etcd

import (
	"encoding/binary"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/logger"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"net"
	"path"
	"time"
)

type Etcd struct {
	log        *logger.Logger
	cli        client.Client
	key        string
	ttl        time.Duration
	retries    time.Duration
	privateKey []byte
	Mappings   map[uint32]*common.Mapping
}

func IP4toInt(IP string) uint32 {
	buf := net.ParseIP(IP).To4()
	return binary.LittleEndian.Uint32(buf)
}

func (e *Etcd) Watch() {
	go func() {
		kapi := client.NewKeysAPI(e.cli)
		watch := kapi.Watcher(e.key+"/mappings", &client.WatcherOptions{Recursive: true})

		for {
			resp, err := watch.Next(context.Background())
			if err != nil {
				e.log.Error("[ETCD]", "Error during watch:", err)
				time.Sleep(e.ttl / e.retries * time.Second)
				continue
			}

			_, key := path.Split(resp.Node.Key)
			switch resp.Action {
			case "set", "update":
				mapping, err := common.ParseMapping(resp.Node.Value, e.privateKey)
				if err != nil {
					e.log.Error("[ETCD]", "Error during watch:", err)
					time.Sleep(e.ttl / e.retries * time.Second)
					continue
				}
				e.Mappings[IP4toInt(key)] = mapping
			case "delete", "expire":
				delete(e.Mappings, IP4toInt(key))
			}
		}
	}()
}

func (e *Etcd) Heartbeat(privateIP string, mapping *common.Mapping) {
	go func() {
		kapi := client.NewKeysAPI(e.cli)
		key := path.Join("/", e.key, "mappings", privateIP)

		refreshOptions := &client.SetOptions{TTL: e.ttl * time.Second, Refresh: true}
		for {
			time.Sleep(e.ttl / e.retries * time.Second)
			_, err := kapi.Set(context.Background(), key, "", refreshOptions)

			if err != nil {
				if client.IsKeyNotFound(err) {
					err := e.SetMapping(privateIP, mapping)
					if err != nil {
						e.log.Error("[ETCD]", "Error during re-registration:", err)
						continue
					}
					err = e.SyncMappings()
					if err != nil {
						e.log.Error("[ETCD]", "Error during sync of cluster:", err)
						continue
					}
					continue
				}
				e.log.Error("[ETCD]", "Error during heartbeat:", err)
			}
		}
	}()
}

func (e *Etcd) SetMapping(privateIP string, mapping *common.Mapping) error {
	kapi := client.NewKeysAPI(e.cli)
	mapping.SecretKey = nil
	_, err := kapi.Set(context.Background(),
		e.key+"/mappings/"+privateIP,
		mapping.String(),
		&client.SetOptions{TTL: e.ttl * time.Second})

	return err
}

func (e *Etcd) SyncMappings() error {
	kapi := client.NewKeysAPI(e.cli)
	mappingsNode, err := kapi.Get(context.Background(),
		e.key+"/mappings",
		&client.GetOptions{Recursive: true})

	if err != nil {
		return err
	}

	for _, v := range mappingsNode.Node.Nodes {
		_, key := path.Split(v.Key)

		mapping, err := common.ParseMapping(v.Value, e.privateKey)
		if err != nil {
			return err
		}
		e.Mappings[IP4toInt(key)] = mapping
	}

	return nil
}

func New(host, key string, privkey []byte, log *logger.Logger) (*Etcd, error) {
	etcdCfg := client.Config{
		Endpoints: []string{host},
	}

	c, err := client.New(etcdCfg)
	if err != nil {
		return nil, err
	}

	mappings := make(map[uint32]*common.Mapping)
	return &Etcd{
		cli:        c,
		key:        key,
		log:        log,
		ttl:        15,
		retries:    3,
		privateKey: privkey,
		Mappings:   mappings,
	}, nil
}
