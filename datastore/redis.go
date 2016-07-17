package datastore

import (
	"github.com/Supernomad/quantum/config"
)

type Redis struct {
}

func newRedis(cfg *config.Config) (DatastoreBackend, error) {
	return nil, nil
}
