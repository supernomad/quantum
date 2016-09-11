package main

import (
	"github.com/Supernomad/quantum/agg"
	"github.com/Supernomad/quantum/backend"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/inet"
	"github.com/Supernomad/quantum/socket"
	"github.com/Supernomad/quantum/workers"
	"os"
	"strconv"
)

const version string = "0.6.0"

func handleError(log *common.Logger, err error) {
	if err != nil {
		log.Error.Println(err)
		os.Exit(1)
	}
}

func main() {
	log := common.NewLogger()

	cfg, err := common.NewConfig()
	handleError(log, err)

	store, err := backend.New(backend.LIBKV, log, cfg)
	handleError(log, err)

	err = store.Init()
	handleError(log, err)
	defer store.Stop()

	tunnel := inet.New(inet.TUNInterface, cfg)
	err = tunnel.Open()
	handleError(log, err)
	defer tunnel.Close()

	sock := socket.New(socket.UDPSocket, cfg)
	err = sock.Open()
	handleError(log, err)
	defer sock.Close()

	outgoing := workers.NewOutgoing(cfg.PrivateIP, cfg.NumWorkers, store, tunnel, sock)
	defer outgoing.Stop()

	incoming := workers.NewIncoming(cfg.PrivateIP, cfg.NumWorkers, store, tunnel, sock)
	defer incoming.Stop()

	aggregator := agg.New(log, cfg, incoming.QueueStats, outgoing.QueueStats)
	defer aggregator.Stop()

	aggregator.Start()
	store.Start()
	for i := 0; i < cfg.NumWorkers; i++ {
		incoming.Start(i)
		outgoing.Start(i)
	}

	log.Info.Println("[MAIN]", "Listening on TUN device:  ", tunnel.Name())
	log.Info.Println("[MAIN]", "TUN network space:        ", cfg.NetworkConfig.Network)
	log.Info.Println("[MAIN]", "TUN private IP address:   ", cfg.PrivateIP)
	log.Info.Println("[MAIN]", "TUN public IP address:    ", cfg.PublicIP)
	log.Info.Println("[MAIN]", "Listening on UDP address: ", cfg.ListenAddress+":"+strconv.Itoa(cfg.ListenPort))

	stop := make(chan bool)
	defer close(stop)
	<-stop
}
