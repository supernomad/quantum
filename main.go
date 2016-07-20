package main

import (
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/config"
	"github.com/Supernomad/quantum/datastore"
	"github.com/Supernomad/quantum/ecdh"
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

	pubkey, privkey, err := ecdh.GenerateECKeyPair()
	handleError(err, log)

	mapping := common.NewMapping(cfg.PublicIP+":"+strconv.Itoa(cfg.ListenPort), pubkey[:])
	handleError(err, log)

	etcd, err := datastore.New(privkey, mapping, cfg)
	handleError(err, log)

	err = etcd.Start()
	handleError(err, log)

	tunnel, err := tun.New(cfg.InterfaceName, cfg.PrivateIP+"/"+cfg.SubnetMask, cores, log)
	handleError(err, log)

	defer tunnel.Close()

	sock, err := socket.New(cfg.ListenAddress, cfg.ListenPort, cores, log)
	handleError(err, log)

	defer sock.Close()

	outgoing := workers.NewOutgoing(log, cfg.PrivateIP, etcd.Mappings, tunnel, sock)
	defer outgoing.Stop()

	incoming := workers.NewIncoming(log, cfg.PrivateIP, etcd.Mappings, tunnel, sock)
	defer incoming.Stop()

	for i := 0; i < cores; i++ {
		incoming.Start(i)
		outgoing.Start(i)
	}

	log.Info.Println("Started up successfuly.")
	log.Info.Println("Listening on TUN device: ", tunnel.Name)
	log.Info.Println("TUN private IP address:  ", cfg.PrivateIP)
	log.Info.Println("TUN private subnet mask: ", cfg.SubnetMask)
	log.Info.Println("TUN public IP address:   ", cfg.PublicIP)
	log.Info.Println("Listening on UDP address:", cfg.ListenAddress+":"+strconv.Itoa(cfg.ListenPort))

	stop := make(chan bool)
	defer close(stop)
	<-stop
}
