package socket

import (
	"github.com/Supernomad/quantum/common"
	"net"
	"syscall"
)

// UDP is a generic multique socket
type UDP struct {
	queues []int
	cfg    *common.Config
	sa     *syscall.SockaddrInet4
}

// Open the socket
func (sock *UDP) Open() error {
	for i := 0; i < sock.cfg.NumWorkers; i++ {
		var queue int
		var err error

		if !sock.cfg.ReuseFDS {
			queue, err = createUDP()
			if err != nil {
				return err
			}
			err = initUDP(queue, sock.sa)
			if err != nil {
				return err
			}
		} else {
			queue = 3 + sock.cfg.NumWorkers + i
		}
		sock.queues[i] = queue
	}
	return nil
}

// Close the socket
func (sock *UDP) Close() error {
	for i := 0; i < len(sock.queues); i++ {
		if err := syscall.Close(sock.queues[i]); err != nil {
			return err
		}
	}
	return nil
}

// GetFDs will return the underlying queue fds
func (sock *UDP) GetFDs() []int {
	return sock.queues
}

// Read a packet from the socket
func (sock *UDP) Read(buf []byte, queue int) (*common.Payload, bool) {
	n, _, err := syscall.Recvfrom(sock.queues[queue], buf, 0)
	if err != nil {
		return nil, false
	}
	return common.NewSockPayload(buf, n), true
}

// Write a packet to the socket
func (sock *UDP) Write(payload *common.Payload, mapping *common.Mapping, queue int) bool {
	sa := &syscall.SockaddrInet4{
		Port: mapping.PublicPort,
	}
	copy(sa.Addr[:], mapping.Addr.To4())
	err := syscall.Sendto(sock.queues[queue], payload.Raw[:payload.Length], 0, sa)
	if err != nil {
		return false
	}
	return true
}

func newUDP(cfg *common.Config) *UDP {
	var addr [4]byte
	copy(addr[:], net.ParseIP(cfg.ListenAddress).To4())
	sa := &syscall.SockaddrInet4{
		Port: cfg.ListenPort,
		Addr: addr,
	}

	queues := make([]int, cfg.NumWorkers)

	return &UDP{queues: queues, cfg: cfg, sa: sa}
}

func initUDP(queue int, sa *syscall.SockaddrInet4) error {
	err := syscall.SetsockoptInt(queue, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		return err
	}

	err = syscall.Bind(queue, sa)
	if err != nil {
		return err
	}

	return nil
}

func createUDP() (int, error) {
	return syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
}
