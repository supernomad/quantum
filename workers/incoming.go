package workers

import (
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/crypto"
	"github.com/Supernomad/quantum/logger"
	"github.com/Supernomad/quantum/nat"
	"github.com/Supernomad/quantum/socket"
	"github.com/Supernomad/quantum/tun"
)

type Incoming struct {
	gcm    *crypto.GCM
	tunnel *tun.Tun
	nat    *nat.Nat
	sock   *socket.Socket
	quit   chan bool
}

func (incoming *Incoming) Start(queue int) {
	go func() {
	loop:
		for {
			select {
			case <-incoming.quit:
				return
			default:
				payload, ok := incoming.sock.Read()
				if !ok {
					continue loop
				}
				payload, ok = incoming.nat.ResolveIncoming(payload)
				if !ok {
					continue loop
				}
				payload, ok = incoming.gcm.Unseal(payload)
				if !ok {
					continue loop
				}
				incoming.tunnel.Write(payload, queue)
			}
		}
	}()
}

func (incoming *Incoming) Stop() {
	go func() {
		incoming.quit <- true
	}()
}

func NewIncoming(log *logger.Logger, privateIP string, mappings map[uint64]common.Mapping, tunnel *tun.Tun, sock *socket.Socket) *Incoming {
	gcm := crypto.NewGCM(log)
	nat := nat.New(privateIP, mappings, log)
	return &Incoming{
		gcm:    gcm,
		tunnel: tunnel,
		sock:   sock,
		nat:    nat,
		quit:   make(chan bool),
	}
}
