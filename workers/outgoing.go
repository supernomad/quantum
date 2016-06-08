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

func (out *Outgoing) Start() {
	go func() {
	loop:
		for {
			select {
			case <-out.quit:
				return
			default:
				payload, ok := out.tunnel.Read()
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

func NewOutgoing(log *logger.Logger, ecdh *crypto.ECDH, mappings map[uint64]common.Mapping, tunnel *tun.Tun, sock *socket.Socket) *Outgoing {
	nat := nat.New(mappings, log)
	gcm := crypto.NewGCM(log, ecdh)
	return &Outgoing{
		nat:    nat,
		gcm:    gcm,
		tunnel: tunnel,
		sock:   sock,
		quit:   make(chan bool),
	}
}
