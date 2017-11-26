// Copyright (c) 2016-2017 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package router

import (
	"encoding/binary"
	"net"

	"github.com/supernomad/quantum/common"
	"github.com/supernomad/quantum/datastore"
)

// Router controls how quantum packets are routed to their destination.
type Router struct {
	cfg   *common.Config
	store datastore.Datastore
}

// Resolve takes the passed in destination and returns the corresponding mapping in the quantum network.
func (rt *Router) Resolve(destination net.IP) (*common.Mapping, bool) {
	dip := binary.LittleEndian.Uint32(destination)

	// Returning a standard mapping if the requested destination exists in the quantum network.
	if rt.cfg.NetworkConfig.IPNet.Contains(destination) {
		return rt.store.Mapping(dip)
	}

	// Return the gateway mapping if it exists.
	return rt.store.GatewayMapping(dip)
}

// New returns a Router struct based on the passed in configuration and key/value store.
func New(cfg *common.Config, store datastore.Datastore) *Router {
	return &Router{
		cfg:   cfg,
		store: store,
	}
}
