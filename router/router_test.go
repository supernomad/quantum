// Copyright (c) 2016-2017 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package router

import (
	"net"
	"testing"

	"github.com/supernomad/quantum/common"
	"github.com/supernomad/quantum/datastore"
)

func TestResolve(t *testing.T) {
	cfg := &common.Config{}
	store := &datastore.Mock{InternalMapping: &common.Mapping{MachineID: "In Network"}, InternalGatewayMapping: &common.Mapping{MachineID: "Out of Network"}}

	base, ipnet, _ := net.ParseCIDR("10.8.0.0/24")
	destInNet := net.ParseIP("10.8.0.1")
	destOutNet := net.ParseIP("8.8.8.8")

	netCfg := &common.NetworkConfig{BaseIP: base, IPNet: ipnet}
	cfg = &common.Config{NetworkConfig: netCfg}

	rt := New(cfg, store)

	if mapping, ok := rt.Resolve(destInNet); !ok || mapping.MachineID != "In Network" {
		t.Fatal("Router did not properly recognize an in network ip address.")
	}

	if mapping, ok := rt.Resolve(destOutNet); !ok || mapping.MachineID != "Out of Network" {
		t.Fatal("Router did not properly recognize an out of network ip address.")
	}
}
