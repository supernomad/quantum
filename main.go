// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package main

import (
	"os"

	"github.com/Supernomad/quantum/agg"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/datastore"
	"github.com/Supernomad/quantum/device"
	"github.com/Supernomad/quantum/plugin"
	"github.com/Supernomad/quantum/socket"
	"github.com/Supernomad/quantum/workers"
)

func handleError(log *common.Logger, err error) {
	if err != nil {
		log.Error.Println(err.Error())
		os.Exit(1)
	}
}

func main() {
	log := common.NewLogger(common.InfoLogger)

	cfg, err := common.NewConfig(log)
	handleError(log, err)

	store, err := datastore.New(datastore.ETCDDatastore, log, cfg)
	handleError(log, err)

	err = store.Init()
	handleError(log, err)

	plugins := make([]plugin.Plugin, len(cfg.Plugins))
	for i := 0; i < len(plugins); i++ {
		plugin, err := plugin.New(cfg.Plugins[i], cfg)
		handleError(log, err)
		plugins[i] = plugin
	}

	dev, err := device.New(device.TUNDevice, cfg)
	handleError(log, err)

	sock, err := socket.New(cfg.NetworkConfig.Backend, cfg)
	handleError(log, err)

	aggregator := agg.New(log, cfg)

	outgoing := workers.NewOutgoing(cfg, aggregator, store, plugins, dev, sock)
	incoming := workers.NewIncoming(cfg, aggregator, store, plugins, dev, sock)

	aggregator.Start()
	store.Start()
	for i := 0; i < cfg.NumWorkers; i++ {
		incoming.Start(i)
		outgoing.Start(i)
	}

	fds := make([]int, cfg.NumWorkers*2)
	copy(fds[0:cfg.NumWorkers], dev.Queues())
	copy(fds[cfg.NumWorkers:cfg.NumWorkers*2], sock.Queues())

	signaler := common.NewSignaler(log, cfg, fds, map[string]string{common.RealDeviceNameEnv: dev.Name()})

	log.Info.Printf("[MAIN] Listening on TUN device:  %s", dev.Name())
	log.Info.Printf("[MAIN] TUN network space:        %s", cfg.NetworkConfig.Network)
	log.Info.Printf("[MAIN] TUN private IP address:   %s", cfg.PrivateIP)
	log.Info.Printf("[MAIN] TUN public IPv4 address:  %s", cfg.PublicIPv4)
	log.Info.Printf("[MAIN] TUN public IPv6 address:  %s", cfg.PublicIPv6)
	log.Info.Printf("[MAIN] Using backend:            %s", cfg.NetworkConfig.Backend)
	log.Info.Printf("[MAIN] Listening on port:        %d", cfg.ListenPort)

	err = signaler.Wait(true)
	handleError(log, err)

	aggregator.Stop()
	store.Stop()

	incoming.Stop()
	outgoing.Stop()

	err = sock.Close()
	handleError(log, err)

	err = dev.Close()
	handleError(log, err)
}
