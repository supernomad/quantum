// Copyright (c) 2016-2018 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package datastore

import (
	"github.com/supernomad/quantum/common"
)

// EtcdV3 datastore struct for interacting with the coreos etcd key/value datastore using the v3 api.
type EtcdV3 struct {
	cfg *common.Config
}

// Mapping returns a mapping and true based on the supplied uint32 representation of an ipv4 address if it exists within the datastore, otherwise it returns nil for the mapping and false.
func (etcd *EtcdV3) Mapping(ip uint32) (*common.Mapping, bool) {
	return nil, true
}

// GatewayMapping should retun the mapping and true if it exists specifically for destinations outside of the quantum network, if the mapping doesn't exist it will return nil and false.
func (etcd *EtcdV3) GatewayMapping() (*common.Mapping, bool) {
	return nil, true
}

// Init the Etcd datastore which will open any necessary connections, preform an initial sync of the datastore, and define the local mapping in the datastore.
func (etcd *EtcdV3) Init() error {
	return nil
}

// Start periodic synchronization, and DHCP lease refresh with the datastore, as well as start watching for changes in network topology.
func (etcd *EtcdV3) Start() {
}

// Stop synchronizing with the backend and shutdown open connections.
func (etcd *EtcdV3) Stop() {
}

func newEtcdV3(cfg *common.Config) (Datastore, error) {
	return &EtcdV3{
		cfg: cfg,
	}, nil
}
