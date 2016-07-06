package socket

import (
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/logger"
	"net"
	"syscall"
)

type Socket struct {
	queues []int
	log    *logger.Logger
}

func (sock *Socket) Close() error {
	for i := 0; i < len(sock.queues); i++ {
		if err := syscall.Close(sock.queues[i]); err != nil {
			return err
		}
	}
	return nil
}

func (sock *Socket) Read(buf []byte, queue int) (*common.Payload, bool) {
	n, _, err := syscall.Recvfrom(sock.queues[queue], buf, 0)
	if err != nil {
		sock.log.Warn("[UDP] Read Error:", err)
		return nil, false
	}
	return common.NewSockPayload(buf, n), true
}

func (sock *Socket) Write(payload *common.Payload, to syscall.Sockaddr, queue int) bool {
	err := syscall.Sendto(sock.queues[queue], payload.Raw[:payload.Length], 0, to)
	if err != nil {
		sock.log.Warn("[UDP]", "Write Error:", err)
		return false
	}
	return true
}

func New(address string, port int, numWorkers int, log *logger.Logger) (*Socket, error) {
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
	return &Socket{queues: queues, log: log}, nil
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
