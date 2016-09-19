package socket

import (
	"github.com/Supernomad/quantum/common"
	"net"
	"syscall"
)

// IP is a generic multique socket
type IP struct {
	queue int
	cfg   *common.Config
	sa    *syscall.SockaddrInet4
}

// Name address of the socket
func (sock *IP) Name() string {
	return sock.cfg.ListenAddress
}

// Open the socket
func (sock *IP) Open() error {
	var queue int
	var err error

	if !sock.cfg.ReuseFDS {
		queue, err = createIP()
		if err != nil {
			return err
		}
		err = initIP(queue, sock.sa)
		if err != nil {
			return err
		}
	} else {
		queue = 3 + 1 + 1
	}
	sock.queue = queue
	return nil
}

// Close the socket
func (sock *IP) Close() error {
	if err := syscall.Close(sock.queue); err != nil {
		return err
	}
	return nil
}

// GetFDs will return the underlying queue fds
func (sock *IP) GetFDs() []int {
	return []int{sock.queue}
}

// Read a packet from the socket
func (sock *IP) Read(buf []byte, queue int) (*common.Payload, bool) {
	n, _, err := syscall.Recvfrom(sock.queue, buf, 0)
	if err != nil {
		return nil, false
	}
	return common.NewIPPayload(buf, n), true
}

// Write a packet to the socket
func (sock *IP) Write(payload *common.Payload, mapping *common.Mapping, queue int) bool {
	err := syscall.Sendto(sock.queue, payload.Raw[:payload.Length], 0, mapping.Sockaddr)
	if err != nil {
		return false
	}
	return true
}

func newIP(cfg *common.Config) *IP {
	var addr [4]byte
	copy(addr[:], net.ParseIP(cfg.ListenAddress).To4())
	sa := &syscall.SockaddrInet4{
		Addr: addr,
	}

	return &IP{queue: -1, cfg: cfg, sa: sa}
}

func initIP(queue int, sa *syscall.SockaddrInet4) error {
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

func createIP() (int, error) {
	return syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, 138)
}
