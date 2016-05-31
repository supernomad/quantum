package main

import (
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/config"
	"github.com/Supernomad/quantum/crypto"
	"github.com/Supernomad/quantum/etcd"
	"github.com/Supernomad/quantum/logger"
	"github.com/Supernomad/quantum/nat"
	"github.com/Supernomad/quantum/socket"
	"github.com/Supernomad/quantum/tun"
	"os"
	"strconv"
)

func main() {
	debugingEnabled := os.Getenv("ESDN_DEBUG") == "true"

	cfg := config.New()
	log := logger.New(debugingEnabled)
	ecdh, err := crypto.NewEcdh(log)
	if err != nil {
		log.Error("[MAIN] Init error: ", err)
		os.Exit(1)
	}

	etcd, err := etcd.New(cfg.EtcdHost, cfg.EtcdKey, log)
	if err != nil {
		log.Error("[MAIN] Init error: ", err)
		os.Exit(1)
	}

	mappings, err := etcd.GetMappings()
	if err != nil {
		log.Error("[MAIN] Init error: ", err)
		os.Exit(1)
	}

	mapping := common.Mapping{
		Address:   cfg.PublicIP + ":" + strconv.Itoa(cfg.ListenPort),
		PublicKey: crypto.SerializeKey(ecdh.PublicKey),
	}
	etcd.SetMapping(cfg.PrivateIP, mapping)

	etcd.Heartbeat(cfg.PrivateIP, mapping)
	etcd.Watch(mappings)

	tunnel, err := tun.New(cfg.InterfaceName, cfg.PrivateIP+"/"+cfg.SubnetMask, log)
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

	nat := nat.New(mappings, log)

	aes := crypto.NewAES(log, ecdh)

	// Outgoing
	outgoing := tunnel.Listen()
	encrypted := make(chan common.Payload, 1024)
	go func() {
		for payload := range outgoing {
			go func(work common.Payload) {
				work = nat.ResolveOutgoing(work)
				work = aes.Encrypt(work)
				encrypted <- work
			}(payload)
		}
	}()
	sock.Send(encrypted)

	// Incoming
	incoming := sock.Listen()
	decrypted := make(chan common.Payload, 1024)
	go func() {
		for payload := range incoming {
			go func(work common.Payload) {
				work = nat.ResolveIncoming(work)
				work = aes.Decrypt(work)
				decrypted <- work
			}(payload)
		}
	}()
	tunnel.Send(decrypted)

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
