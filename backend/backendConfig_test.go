package backend

import (
	"github.com/Supernomad/quantum/common"
	"testing"
)

func TestGenerateStoreConfig(t *testing.T) {
	cfg := &common.Config{
		AuthEnabled: true,
		Username:    "hello",
		Password:    "goodbye",
		TLSEnabled:  true,
		TLSKey:      "../bin/certs/quantum1.quantum.dev.key",
		TLSCert:     "../bin/certs/quantum1.quantum.dev.crt",
		TLSCA:       "../bin/certs/ca.crt",
	}
	storeCfg, err := generateStoreConfig(cfg)
	if err != nil {
		t.Fatalf("generateStoreConfig returned an error: %v", err)
	}
	if storeCfg == nil {
		t.Fatalf("generateStoreConfig didn't return the correct value, storeCfg is nil.")
	}
	if storeCfg.Username != "hello" {
		t.Fatalf("generateStoreConfig didn't return the correct value, Username is not correct.")
	}
	if storeCfg.Password != "goodbye" {
		t.Fatalf("generateStoreConfig didn't return the correct value, Password is not correct.")
	}
	if storeCfg.TLS == nil {
		t.Fatalf("generateStoreConfig didn't return the correct value, TLS is nil.")
	}
}
