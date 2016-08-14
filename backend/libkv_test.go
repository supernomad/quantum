package backend

import (
	"github.com/Supernomad/quantum/common"
	"testing"
)

var backend Backend

func TestNewLibkv(t *testing.T) {
	cfg := &common.Config{
		Datastore:   "mock",
		Prefix:      "quantum",
		MachineID:   "0293ohf0hf0w8hf809y4hsf0h",
		Endpoints:   []string{"etcd.quantum.dev:2379"},
		PrivateIP:   "10.1.1.1",
		AuthEnabled: true,
		Username:    "hello",
		Password:    "goodbye",
		TLSEnabled:  true,
		TLSKey:      "../bin/certs/quantum1.quantum.dev.key",
		TLSCert:     "../bin/certs/quantum1.quantum.dev.crt",
		TLSCA:       "../bin/certs/ca.crt",
	}
	var err error
	backend, err = newLibkv(cfg)
	if err != nil {
		t.Fatalf("newLibkv returned an error: %v", err)
	}
	if backend == nil {
		t.Fatal("newLibkv returned an empty backend.")
	}
}