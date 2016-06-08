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

func (sock *Socket) Listen() <-chan *common.Payload {
	out := make(chan *common.Payload, 1024)

	go func() {
		for {
			buf := make([]byte, common.MaxPacketLength)
			n, err := sock.conn.Read(buf)
			if err != nil {
				sock.log.Warn("[UDP] Read Error:", err)
				continue
			}
			out <- common.NewSockPayload(buf, n)
		}
	}()

	return out
}

func (sock *Socket) Read() (*common.Payload, bool) {
	buf := make([]byte, common.MaxPacketLength)
	n, err := sock.conn.Read(buf)
	if err != nil {
		sock.log.Warn("[UDP] Read Error:", err)
		return nil, false
	}
	return common.NewSockPayload(buf, n), true
}

func (sock *Socket) Send(encrypted <-chan *common.Payload) {
	go func() {
		for payload := range encrypted {
			addr, err := net.ResolveUDPAddr("udp", payload.Address)
			if err != nil {
				sock.log.Warn("[UDP] Resolve Address Error:", err)
				continue
			}

			_, err = sock.conn.WriteToUDP(payload.Raw[:payload.Length], addr)
			if err != nil {
				sock.log.Warn("[UDP] Write Error:", err)
				continue
			}
		}
	}()
}

func (sock *Socket) Write(payload *common.Payload) bool {
	addr, err := net.ResolveUDPAddr("udp", payload.Address)
	if err != nil {
		sock.log.Warn("[UDP] Resolve Address Error:", err)
		return false
	}

	_, err = sock.conn.WriteToUDP(payload.Raw[:payload.Length], addr)
	if err != nil {
		sock.log.Warn("[UDP] Write Error:", err)
		return false
	}

	return true
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
