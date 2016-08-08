package socket

import (
	"github.com/Supernomad/quantum/common"
	"net"
	"syscall"
)

// Socket is a generic socket interface based on socket fd's and syscalls
type Socket struct {
	queues []int
}

// Close the socket
func (sock *Socket) Close() error {
	for i := 0; i < len(sock.queues); i++ {
		if err := syscall.Close(sock.queues[i]); err != nil {
			return err
		}
	}
	return nil
}

// Read a packet from the socket
func (sock *Socket) Read(buf []byte, queue int) (*common.Payload, bool) {
	n, _, err := syscall.Recvfrom(sock.queues[queue], buf, 0)
	if err != nil {
		return nil, false
	}
	return common.NewSockPayload(buf, n), true
}

// Write a packet to the socket
func (sock *Socket) Write(payload *common.Payload, to syscall.Sockaddr, queue int) bool {
	err := syscall.Sendto(sock.queues[queue], payload.Raw[:payload.Length], 0, to)
	if err != nil {
		return false
	}
	return true
}

// New socket
func New(address string, port int, numWorkers int) (*Socket, error) {
	var addr [4]byte
	copy(addr[:], net.ParseIP(address).To4())

	queues := make([]int, numWorkers)

	sa := &syscall.SockaddrInet4{
		Port: port,
		Addr: addr,
	}

	for i := 0; i < numWorkers; i++ {
		queue, err := createSocket()
		if err != nil {
			return nil, err
		}

		err = initSocket(queue, sa)
		if err != nil {
			return nil, err
		}

		queues[i] = queue
	}
	return &Socket{queues: queues}, nil
}

func initSocket(queue int, sa *syscall.SockaddrInet4) error {
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

func createSocket() (int, error) {
	return syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
}
