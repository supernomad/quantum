package datastore

import (
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/config"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"path"
	"time"
)

type Etcd struct {
	cli client.Client
}

func (etcd *Etcd) Watch(prefix string, handler MappingHandler) {
	kapi := client.NewKeysAPI(etcd.cli)
	watch := kapi.Watcher(prefix+"/mappings", &client.WatcherOptions{Recursive: true})

	for {
		resp, err := watch.Next(context.Background())
		if err != nil {
			continue
		}

		_, key := path.Split(resp.Node.Key)

		var action DatastoreAction
		switch resp.Action {
		case "set", "update":
			action = UpdateAction
		case "delete", "expire":
			action = RemoveAction
		}

		handler(action, key, resp.Node.Value)
	}
}

func (etcd *Etcd) ResetTtl(prefix, privateIp string, ttl time.Duration) error {
	kapi := client.NewKeysAPI(etcd.cli)

	_, err := kapi.Set(context.Background(),
		path.Join(prefix, "mappings", privateIp),
		"",
		&client.SetOptions{TTL: ttl * time.Second, Refresh: true})

	return err
}

func (etcd *Etcd) Sync(prefix string, handler MappingHandler) error {
	kapi := client.NewKeysAPI(etcd.cli)
	mappingsNode, err := kapi.Get(context.Background(),
		path.Join(prefix, "/mappings/"),
		&client.GetOptions{Recursive: true})

	if err != nil {
		return err
	}

	for _, v := range mappingsNode.Node.Nodes {
		_, key := path.Split(v.Key)
		handler(UpdateAction, key, v.Value)
	}

	return nil
}

func (etcd *Etcd) Set(prefix, privateIp string, ttl time.Duration, mapping *common.Mapping) error {
	kapi := client.NewKeysAPI(etcd.cli)

	_, err := kapi.Set(context.Background(),
		path.Join(prefix, "/mappings/", privateIp),
		mapping.String(),
		&client.SetOptions{TTL: ttl * time.Second})

	return err
}

func newEtcd(cfg *config.Config) (DatastoreBackend, error) {
	etcdCfg := client.Config{
		Endpoints: cfg.EtcdEndpoints,
		Username:  cfg.EtcdUsername,
		Password:  cfg.EtcdPassword,
	}

	c, err := client.New(etcdCfg)
	if err != nil {
		return nil, err
	}

	return &Etcd{
		cli: c,
	}, nil
}
