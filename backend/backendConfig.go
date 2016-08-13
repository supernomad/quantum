package backend

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/Supernomad/quantum/common"
	"github.com/docker/libkv/store"
	"io/ioutil"
)

func generateStoreConfig(cfg *common.Config) (*store.Config, error) {
	storeCfg := &store.Config{PersistConnection: true}

	if cfg.AuthEnabled {
		storeCfg.Username = cfg.Username
		storeCfg.Password = cfg.Password
	}

	if cfg.TLSEnabled {
		storeCfg.TLS = &tls.Config{}
		if cfg.TLSKey != "" && cfg.TLSCert != "" {
			cert, err := tls.LoadX509KeyPair(cfg.TLSCert, cfg.TLSKey)
			if err != nil {
				return nil, err
			}
			storeCfg.TLS.Certificates = []tls.Certificate{cert}
			storeCfg.TLS.BuildNameToCertificate()
		}
		if cfg.TLSCA != "" {
			cert, err := ioutil.ReadFile(cfg.TLSCA)
			if err != nil {
				return nil, err
			}
			storeCfg.TLS.RootCAs = x509.NewCertPool()
			storeCfg.TLS.RootCAs.AppendCertsFromPEM(cert)
			storeCfg.TLS.BuildNameToCertificate()
		}
	}

	return storeCfg, nil
}
