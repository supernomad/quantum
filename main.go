package main

import (
	"github.com/Supernomad/quantum/agg"
	"github.com/Supernomad/quantum/backend"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/socket"
	"github.com/Supernomad/quantum/workers"
	"github.com/go-playground/log"
	"github.com/go-playground/log/handlers/console"
	"os"
	"strconv"
)

const version string = "0.6.0"

func handleError(err error) {
	if err != nil {
		os.Exit(1)
	}
}

func main() {
	cLog := console.New()
	log.RegisterHandler(cLog, log.AllLevels...)
	log.Infof("Starting up quantum v%s", version)

	cfg, err := common.NewConfig()
	handleError(err)

	store, err := backend.New(cfg)
	handleError(err)

	err = store.Init()
	handleError(err)
	defer store.Stop()

	tunnel := socket.New(socket.TUNSock, cfg)
	err = tunnel.Open()
	handleError(err)
	defer tunnel.Close()

	sock := socket.New(socket.UDPSock, cfg)
	err = sock.Open()
	handleError(err)
	defer sock.Close()

	outgoing := workers.NewOutgoing(cfg.PrivateIP, cfg.NumWorkers, store, tunnel, sock)
	defer outgoing.Stop()

	incoming := workers.NewIncoming(cfg.PrivateIP, cfg.NumWorkers, store, tunnel, sock)
	defer incoming.Stop()

	aggregator := agg.New(cfg, incoming.QueueStats, outgoing.QueueStats)
	defer aggregator.Stop()

	aggregator.Start()
	store.Start()
	for i := 0; i < cfg.NumWorkers; i++ {
		incoming.Start(i)
		outgoing.Start(i)
	}

	log.Info("Listening on TUN device:  ", tunnel.Name())
	log.Info("TUN network space:        ", store.NetworkCfg.Network)
	log.Info("TUN private IP address:   ", cfg.PrivateIP)
	log.Info("TUN public IP address:    ", cfg.PublicIP)
	log.Info("Listening on UDP address: ", cfg.ListenAddress+":"+strconv.Itoa(cfg.ListenPort))

	stop := make(chan bool)
	defer close(stop)
	<-stop
}
