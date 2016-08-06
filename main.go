package main

import (
	"github.com/Supernomad/quantum/config"
	"github.com/Supernomad/quantum/datastore"
	"github.com/Supernomad/quantum/logger"
	"github.com/Supernomad/quantum/socket"
	"github.com/Supernomad/quantum/tun"
	"github.com/Supernomad/quantum/workers"
	"os"
	"runtime"
	"strconv"
)

func handleError(err error, log *logger.Logger) {
	if err != nil {
		log.Error.Println("[Initialization Error]", err)
		os.Exit(1)
	}
}

func main() {
	debugingEnabled := os.Getenv("QUANTUM_DEBUG") == "true"

	cores := runtime.NumCPU()
	runtime.GOMAXPROCS(cores * 2)

	cfg := config.New()
	log := logger.New(debugingEnabled)

	store, err := datastore.New(cfg)
	handleError(err, log)
	defer store.Stop()

	tunnel, err := tun.New(cfg.InterfaceName, cfg.PrivateIP, store.Network, cores, log)
	handleError(err, log)
	defer tunnel.Close()

	sock, err := socket.New(cfg.ListenAddress, cfg.ListenPort, cores, log)
	handleError(err, log)
	defer sock.Close()

	outgoing := workers.NewOutgoing(log, cfg.PrivateIP, store.Mappings, tunnel, sock)
	defer outgoing.Stop()

	incoming := workers.NewIncoming(log, cfg.PrivateIP, store.Mappings, tunnel, sock)
	defer incoming.Stop()

	store.Start()
	for i := 0; i < cores; i++ {
		incoming.Start(i)
		outgoing.Start(i)
	}

	log.Info.Println("Listening on TUN device:  ", tunnel.Name)
	log.Info.Println("TUN network space:        ", store.Network)
	log.Info.Println("TUN private IP address:   ", cfg.PrivateIP)
	log.Info.Println("TUN public IP address:    ", cfg.PublicIP)
	log.Info.Println("Listening on UDP address: ", cfg.ListenAddress+":"+strconv.Itoa(cfg.ListenPort))

	stop := make(chan bool)
	defer close(stop)
	<-stop
}
