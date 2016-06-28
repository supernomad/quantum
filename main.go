package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/config"
	"github.com/Supernomad/quantum/ecdh"
	"github.com/Supernomad/quantum/etcd"
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
		log.Error("[MAIN] Init error: ", err)
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

	signkey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	verifykey := signkey.Public().(*ecdsa.PublicKey)
	handleError(err, log)

	etcd, err := etcd.New(cfg.EtcdHost, cfg.EtcdKey, privkey, log)
	handleError(err, log)

	err = etcd.SyncMappings()
	handleError(err, log)

	mapping := common.NewMapping(cfg.PublicIP+":"+strconv.Itoa(cfg.ListenPort), pubkey[:], verifykey)
	handleError(err, log)

	etcd.SetMapping(cfg.PrivateIP, mapping)
	etcd.Heartbeat(cfg.PrivateIP, mapping)
	etcd.Watch()

	tunnel, err := tun.New(cfg.InterfaceName, cfg.PrivateIP+"/"+cfg.SubnetMask, cores, log)
	handleError(err, log)

	defer tunnel.Close()

	sock, err := socket.New(cfg.ListenAddress, cfg.ListenPort, cores, log)
	handleError(err, log)

	defer sock.Close()

	outgoing := workers.NewOutgoing(log, cfg.PrivateIP, signkey, etcd.Mappings, tunnel, sock)
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
