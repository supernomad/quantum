package etcd

import (
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/logger"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"path"
	"time"
)

type Etcd struct {
	log     *logger.Logger
	cli     client.Client
	key     string
	ttl     time.Duration
	retries time.Duration
}

func (e *Etcd) Watch(mappings map[string]common.Mapping) {
	go func() {
		kapi := client.NewKeysAPI(e.cli)
		watch := kapi.Watcher(e.key+"/mappings", &client.WatcherOptions{Recursive: true})

		for {
			resp, err := watch.Next(context.Background())
			if err != nil {
				e.log.Error("[ETCD]", "Error during watch:", err)
				continue
			}

			_, key := path.Split(resp.Node.Key)
			switch resp.Action {
			case "set", "update":
				mapping := common.ParseMapping(resp.Node.Value)
				mappings[key] = mapping
			case "delete", "expire":
				delete(mappings, key)
			}
		}
	}()
}

func (e *Etcd) Heartbeat(privateIP string) {
	go func() {
		kapi := client.NewKeysAPI(e.cli)
		setOptions := &client.SetOptions{TTL: e.ttl * time.Second, Refresh: true}
		for {
			time.Sleep(e.ttl / e.retries * time.Second)
			_, err := kapi.Set(context.Background(), e.key+"/mappings/"+privateIP, "", setOptions)

			if err != nil {
				e.log.Error("[ETCD]", "Error during heartbeat:", err)
				continue
			}
		}
	}()
}

func (e *Etcd) SetMapping(privateIP string, mapping common.Mapping) error {
	kapi := client.NewKeysAPI(e.cli)
	_, err := kapi.Set(context.Background(),
		e.key+"/mappings/"+privateIP,
		mapping.String(),
		&client.SetOptions{TTL: e.ttl * time.Second})

	return err
}

func (e *Etcd) GetMappings() (map[string]common.Mapping, error) {
	kapi := client.NewKeysAPI(e.cli)
	mappingsNode, err := kapi.Get(context.Background(),
		e.key+"/mappings",
		&client.GetOptions{Recursive: true})

	if err != nil {
		return nil, err
	}

	mappings := map[string]common.Mapping{}
	for _, v := range mappingsNode.Node.Nodes {
		_, key := path.Split(v.Key)

		mapping := common.ParseMapping(v.Value)
		mappings[key] = mapping
	}

	return mappings, nil
}

func New(host string, key string, log *logger.Logger) (*Etcd, error) {
	etcdCfg := client.Config{
		Endpoints: []string{host},
	}

	c, err := client.New(etcdCfg)
	if err != nil {
		return nil, err
	}

	return &Etcd{
		cli:     c,
		key:     key,
		log:     log,
		ttl:     15,
		retries: 3,
	}, nil
}
