package socket

import (
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/logger"
	"net"
	"strconv"
)

type Socket struct {
	conn    *net.UDPConn
	log     *logger.Logger
	Address string
}

func (sock *Socket) Close() error {
	return sock.conn.Close()
}

func (sock *Socket) Listen() <-chan common.Payload {
	out := make(chan common.Payload, 1024)

	go func() {
		for {
			buf := make([]byte, 65535)
			n, err := sock.conn.Read(buf)
			if err != nil {
				sock.log.Warn("[UDP] Read Error:", err)
				continue
			}
			out <- common.Payload{Packet: buf[:n]}
		}
	}()

	return out
}

func (sock *Socket) Send(sealed <-chan common.Payload) {
	go func() {
		for payload := range sealed {
			addr, err := net.ResolveUDPAddr("udp", payload.Address)
			if err != nil {
				sock.log.Warn("[UDP] Resolve Address Error:", err)
				continue
			}

			n, err := sock.conn.WriteToUDP(payload.Packet, addr)
			if err != nil || n != len(payload.Packet) {
				sock.log.Warn("[UDP] Write Error:", err)
				continue
			}
		}
	}()
}

func New(address string, port int, log *logger.Logger) (*Socket, error) {
	saddr := address + ":" + strconv.Itoa(port)
	addr, err := net.ResolveUDPAddr("udp", saddr)

	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	return &Socket{conn: conn, log: log, Address: saddr}, nil
}
