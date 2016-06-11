package workers

import (
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/crypto"
	"github.com/Supernomad/quantum/logger"
	"github.com/Supernomad/quantum/nat"
	"github.com/Supernomad/quantum/socket"
	"github.com/Supernomad/quantum/tun"
)

type Outgoing struct {
	nat    *nat.Nat
	gcm    *crypto.GCM
	tunnel *tun.Tun
	sock   *socket.Socket
	quit   chan bool
}

func (out *Outgoing) Start(queue int) {
	go func() {
	loop:
		for {
			select {
			case <-out.quit:
				return
			default:
				payload, ok := out.tunnel.Read(queue)
				if !ok {
					continue loop
				}
				payload, ok = out.nat.ResolveOutgoing(payload)
				if !ok {
					continue loop
				}
				payload, ok = out.gcm.Seal(payload)
				if !ok {
					continue loop
				}
				out.sock.Write(payload)
			}
		}
	}()
}

func (out *Outgoing) Stop() {
	go func() {
		out.quit <- true
	}()
}

func NewOutgoing(log *logger.Logger, privateIP string, mappings map[uint64]common.Mapping, tunnel *tun.Tun, sock *socket.Socket) *Outgoing {
	gcm := crypto.NewGCM(log)
	nat := nat.New(privateIP, mappings, log)
	return &Outgoing{
		nat:    nat,
		gcm:    gcm,
		tunnel: tunnel,
		sock:   sock,
		quit:   make(chan bool),
	}
}
