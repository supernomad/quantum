package main

import (
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/config"
	"github.com/Supernomad/quantum/crypto"
	"github.com/Supernomad/quantum/etcd"
	"github.com/Supernomad/quantum/logger"
	"github.com/Supernomad/quantum/socket"
	"github.com/Supernomad/quantum/tun"
	"github.com/Supernomad/quantum/workers"
	"os"
	"runtime"
	"strconv"
)

func main() {
	debugingEnabled := os.Getenv("QUANTUM_DEBUG") == "true"

	cores := runtime.NumCPU()
	runtime.GOMAXPROCS(cores)

	cfg := config.New()
	log := logger.New(debugingEnabled)

	ecdh, err := crypto.NewEcdh(log)
	if err != nil {
		log.Error("[MAIN] Init error: ", err)
		os.Exit(1)
	}

	etcd, err := etcd.New(cfg.EtcdHost, cfg.EtcdKey, ecdh, log)
	if err != nil {
		log.Error("[MAIN] Init error: ", err)
		os.Exit(1)
	}

	err = etcd.SyncMappings()
	if err != nil {
		log.Error("[MAIN] Init error: ", err)
		os.Exit(1)
	}

	mapping := common.Mapping{
		Address:   cfg.PublicIP + ":" + strconv.Itoa(cfg.ListenPort),
		PublicKey: ecdh.PublicKey[:],
	}

	etcd.SetMapping(cfg.PrivateIP, mapping)
	etcd.Heartbeat(cfg.PrivateIP, mapping)
	etcd.Watch()

	tunnel, err := tun.New(cfg.InterfaceName, cfg.PrivateIP+"/"+cfg.SubnetMask, cores, log)
	defer tunnel.Close()

	if err != nil {
		log.Error("[MAIN] Init error: ", err)
		os.Exit(1)
	}

	sock, err := socket.New(cfg.ListenAddress, cfg.ListenPort, log)
	defer sock.Close()

	if err != nil {
		log.Error("[MAIN] Init error: ", err)
		os.Exit(1)
	}

	outgoing := workers.NewOutgoing(log, cfg.PrivateIP, etcd.Mappings, tunnel, sock)
	defer outgoing.Stop()

	incoming := workers.NewIncoming(log, cfg.PrivateIP, etcd.Mappings, tunnel, sock)
	defer incoming.Stop()

	for i := 0; i < cores; i++ {
		incoming.Start(i)
		outgoing.Start(i)
	}

	log.Info("[MAIN] Started up successfuly.")
	log.Info("[MAIN] Listening on TUN device: ", tunnel.Name)
	log.Info("[MAIN] TUN private IP address:  ", cfg.PrivateIP)
	log.Info("[MAIN] TUN private subnet mask: ", cfg.SubnetMask)
	log.Info("[MAIN] TUN public IP address:   ", cfg.PublicIP)
	log.Info("[MAIN] Listening on UDP address:", cfg.ListenAddress+":"+strconv.Itoa(cfg.ListenPort))

	stop := make(chan bool)
	defer close(stop)
	<-stop
}
