package main

import (
	"github.com/Supernomad/quantum/agg"
	"github.com/Supernomad/quantum/backend"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/inet"
	"github.com/Supernomad/quantum/socket"
	"github.com/Supernomad/quantum/workers"
	"os"
	"os/signal"
	"strconv"
	"syscall"
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

	tunnel := inet.New(inet.TUNInterface, cfg)
	err = tunnel.Open()
	handleError(log, err)

	sock := socket.New(socket.UDPSocket, cfg)
	err = sock.Open()
	handleError(log, err)

	outgoing := workers.NewOutgoing(cfg.PrivateIP, cfg.NumWorkers, store, tunnel, sock)

	incoming := workers.NewIncoming(cfg.PrivateIP, cfg.NumWorkers, store, tunnel, sock)

	aggregator := agg.New(log, cfg, incoming.QueueStats, outgoing.QueueStats)

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

	signals := make(chan os.Signal, 1)
	done := make(chan struct{}, 1)

	signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	go func() {
		for {
			sig := <-signals
			switch {
			case sig == syscall.SIGHUP:
				log.Info.Println("[MAIN]", "Recieved:", sig.String(), "Reloading process.")

				sockFDS := sock.GetFDs()
				tunFDS := tunnel.GetFDs()

				files := make([]uintptr, 3+cfg.NumWorkers*2)
				files[0] = os.Stdin.Fd()
				files[1] = os.Stdout.Fd()
				files[2] = os.Stderr.Fd()
				log.Info.Println("[MAIN]", "Sock FDs:", sockFDS)
				log.Info.Println("[MAIN]", "Tun FDs:", tunFDS)
				for i := 0; i < cfg.NumWorkers; i++ {
					files[3+i] = uintptr(tunFDS[i])
					files[3+i+cfg.NumWorkers] = uintptr(sockFDS[i])
				}
				log.Info.Println("[MAIN]", "Files:", files)
				os.Setenv(common.RealInterfaceNameEnv, tunnel.Name())
				env := os.Environ()
				attr := &syscall.ProcAttr{
					Env:   env,
					Files: files,
				}

				incoming.Stop()
				outgoing.Stop()

				arg0 := os.Args[0]
				args := os.Args
				args = append(args, common.ReloadTrigger)
				_, err := syscall.ForkExec(arg0, args, attr)
				handleError(log, err)

				done <- struct{}{}
			case sig == syscall.SIGINT || sig == syscall.SIGTERM || sig == syscall.SIGKILL:
				log.Info.Println("[MAIN]", "Recieved:", sig.String(), "Terminating process.")

				aggregator.Stop()
				incoming.Stop()
				outgoing.Stop()
				sock.Close()
				store.Stop()
				tunnel.Close()

				done <- struct{}{}
			}
		}
	}()

	<-done
}
