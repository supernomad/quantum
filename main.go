// Package main entrypoint for quantum
// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE
package main

import (
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/Supernomad/quantum/agg"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/datastore"
	"github.com/Supernomad/quantum/device"
	"github.com/Supernomad/quantum/socket"
	"github.com/Supernomad/quantum/workers"
)

func handleError(log *common.Logger, err error, stack string) {
	if err != nil {
		log.Error.Println(err.Error(), "Stack:", stack)
		os.Exit(1)
	}
}

func main() {
	log := common.NewLogger(true, true, true, true)
	wg := &sync.WaitGroup{}

	cfg, err := common.NewConfig(log)
	handleError(log, err, "common.NewConfig()")

	store, err := datastore.New(datastore.ETCDDatastore, log, cfg)
	handleError(log, err, "backend.New(backend.LIBKV, log, cfg)")

	err = store.Init()
	handleError(log, err, "store.Init()")

	dev := device.New(device.TUNDevice, cfg)
	err = dev.Open()
	handleError(log, err, "dev.Open()")

	sock := socket.New(socket.UDPSocket, cfg)
	err = sock.Open()
	handleError(log, err, "sock.Open()")

	aggregator := agg.New(log, cfg)

	outgoing := workers.NewOutgoing(cfg, aggregator, store, dev, sock)
	incoming := workers.NewIncoming(cfg, aggregator, store, dev, sock)

	wg.Add(2)
	aggregator.Start(wg)
	store.Start(wg)
	for i := 0; i < cfg.NumWorkers; i++ {
		incoming.Start(i)
		outgoing.Start(i)
	}

	log.Info.Println("[MAIN]", "Listening on TUN device:  ", dev.Name())
	log.Info.Println("[MAIN]", "TUN network space:        ", cfg.NetworkConfig.Network)
	log.Info.Println("[MAIN]", "TUN private IP address:   ", cfg.PrivateIP)
	log.Info.Println("[MAIN]", "TUN public IPv4 address:  ", cfg.PublicIPv4)
	log.Info.Println("[MAIN]", "TUN public IPv6 address:  ", cfg.PublicIPv6)
	log.Info.Println("[MAIN]", "Listening on UDP port:    ", strconv.Itoa(cfg.ListenPort))

	exit := make(chan struct{})
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	go func() {
		sig := <-signals
		switch {
		case sig == syscall.SIGHUP:
			log.Info.Println("[MAIN]", "Received reload signal from user. Reloading process.")

			sockFDS := sock.GetFDs()
			tunFDS := dev.Queues()

			files := make([]uintptr, 3+cfg.NumWorkers*2)
			files[0] = os.Stdin.Fd()
			files[1] = os.Stdout.Fd()
			files[2] = os.Stderr.Fd()

			for i := 0; i < cfg.NumWorkers; i++ {
				files[3+i] = uintptr(tunFDS[i])
				files[3+i+cfg.NumWorkers] = uintptr(sockFDS[i])
			}

			os.Setenv(common.RealDeviceNameEnv, dev.Name())
			pid, err := syscall.ForkExec(os.Args[0], os.Args, &syscall.ProcAttr{
				Env:   os.Environ(),
				Files: files,
			})
			handleError(log, err, "syscall.ForkExec(os.Args[0], os.Args, &syscall.ProcAttr{Env: os.Environ(), Files: files})")

			ioutil.WriteFile(cfg.PidFile, []byte(strconv.Itoa(pid)), 0644)
		case sig == syscall.SIGINT || sig == syscall.SIGTERM || sig == syscall.SIGKILL:
			log.Info.Println("[MAIN]", "Received termination signal from user. Terminating process.")
		}
		exit <- struct{}{}
	}()
	<-exit

	aggregator.Stop()
	store.Stop()
	incoming.Stop()
	outgoing.Stop()

	sock.Close()
	dev.Close()
	wg.Wait()
}
