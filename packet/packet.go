package packet

/*
#include <linux/if_packet.h>
#include <linux/if_ether.h>
#include <linux/filter.h>
#include <arpa/inet.h>
#include <sys/mman.h>
#include <poll.h>
#include <unistd.h>
*/
import "C"

import (
	"fmt"
	"github.com/Supernomad/quantum/common"
	"syscall"
	"unsafe"
)

const (
	rxRing = C.PACKET_RX_RING
	txRing = C.PACKET_TX_RING

	numBlocks      = 64
	frameSize      = 65536
	blockSize      = frameSize * numBlocks
	framesPerBlock = blockSize / frameSize
	totalSize      = framesPerBlock * numBlocks * frameSize
)

type Packet struct {
	rings map[int][]*RingBuffer
	cfg   *common.Config
	sa    *syscall.SockaddrLinklayer
}

type RingBuffer struct {
	queue  int
	buffer []byte
	poll   C.struct_pollfd
	offset int
}

var tpreq C.struct_tpacket_req
var bpfCode [5]C.struct_sock_filter
var filter C.struct_sock_fprog

func init() {
	tpreq.tp_block_size = C.uint(blockSize)
	tpreq.tp_block_nr = C.uint(numBlocks)
	tpreq.tp_frame_size = C.uint(frameSize)
	tpreq.tp_frame_nr = C.uint(framesPerBlock * numBlocks)

	bpfCode[0] = C.struct_sock_filter{
		code: 0x00,
		jt:   0,
		jf:   0,
		k:    0x00000000,
	}
	bpfCode[1] = C.struct_sock_filter{
		code: 0x30,
		jt:   0,
		jf:   0,
		k:    0x00000009,
	}
	bpfCode[2] = C.struct_sock_filter{
		code: 0x15,
		jt:   0,
		jf:   1,
		k:    0x0000008a,
	}
	bpfCode[3] = C.struct_sock_filter{
		code: 0x06,
		jt:   0,
		jf:   0,
		k:    0x00010000,
	}
	bpfCode[4] = C.struct_sock_filter{
		code: 0x06,
		jt:   0,
		jf:   0,
		k:    0x00000000,
	}
	filter.len = 5
	filter.filter = &bpfCode[0]
}

func (packet *Packet) poll(rtype int, queue int, events C.short) error {
	packet.rings[rtype][queue].poll.fd = C.int(packet.rings[rtype][queue].queue)
	packet.rings[rtype][queue].poll.events = C.short(events)
	packet.rings[rtype][queue].poll.revents = 0
	_, err := C.poll(&packet.rings[rtype][queue].poll, 1, -1)
	return err
}

func (packet *Packet) nextFrame(rtype int, queue int, status C.ulong, events C.short) (*C.struct_tpacket_hdr, []byte, error) {
	pos := packet.rings[rtype][queue].offset * frameSize
	tph := (*C.struct_tpacket_hdr)(unsafe.Pointer(&packet.rings[rtype][queue].buffer[pos]))

	for tph.tp_status&status != status {
		err := packet.poll(rtype, queue, events)
		if err != nil {
			return nil, nil, fmt.Errorf("Poll error: %s", err)
		}
	}

	packet.rings[rtype][queue].offset++
	if packet.rings[rtype][queue].offset >= framesPerBlock*numBlocks {
		packet.rings[rtype][queue].offset = 0
	}

	var start, end int
	switch rtype {
	case rxRing:
		start = pos + int(tph.tp_net)
		end = start + int(tph.tp_snaplen)
	case txRing:
		var sll C.struct_sockaddr_ll
		start = pos + C.TPACKET_HDRLEN - int(unsafe.Sizeof(sll))
		end = start + frameSize
	}
	return tph, packet.rings[rtype][queue].buffer[start:end], nil
}

func initPacket(rtype int, sa *syscall.SockaddrLinklayer) (int, []byte, error) {
	// create the socket
	queue, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_DGRAM, C.ETH_P_IP)
	if err != nil {
		return -1, nil, fmt.Errorf("Create socket error: %v", err)
	}

	// bind the socket to the public interface
	err = syscall.Bind(queue, sa)
	if err != nil {
		return -1, nil, fmt.Errorf("Bind socket error: %v", err)
	}

	// set the socket C.PACKET_FANOUT options
	fanoutVar := (C.PACKET_FANOUT_LB | C.PACKET_FANOUT_FLAG_ROLLOVER | C.PACKET_FANOUT_FLAG_DEFRAG) << 16
	err = syscall.SetsockoptInt(queue, syscall.SOL_PACKET, C.PACKET_FANOUT, fanoutVar)
	if err != nil {
		return -1, nil, fmt.Errorf("Fanout option error: %v", err)
	}

	// set the base bpf filter to only get quantum traffic
	_, err = C.setsockopt(C.int(queue), C.SOL_SOCKET, C.int(C.SO_ATTACH_FILTER), unsafe.Pointer(&filter), C.socklen_t(unsafe.Sizeof(filter)))
	if err != nil {
		return -1, nil, fmt.Errorf("Bpf option error: %v", err)
	}

	// setup the ring buffer
	_, err = C.setsockopt(C.int(queue), C.SOL_PACKET, C.int(rtype), unsafe.Pointer(&tpreq), C.socklen_t(unsafe.Sizeof(tpreq)))
	if err != nil {
		return -1, nil, fmt.Errorf("TxRing option error: %v", err)
	}

	// generate the actual memory buffer for the rx/tx ring buffers
	ring, err := syscall.Mmap(queue, 0, totalSize, C.PROT_READ|C.PROT_WRITE, C.MAP_SHARED)
	if err != nil {
		return -1, nil, fmt.Errorf("Mmap error: %v", err)
	}
	return queue, ring, nil
}

func (packet *Packet) GetFDs() []int {
	fds := make([]int, packet.cfg.NumWorkers*2)
	for i := 0; i < packet.cfg.NumWorkers; i++ {
		fds[i] = packet.rings[rxRing][i].queue
		fds[i+packet.cfg.NumWorkers] = packet.rings[txRing][i].queue
	}

	return fds
}

func (packet *Packet) Open() error {
	for i := 0; i < packet.cfg.NumWorkers; i++ {
		rxq, rxr, err := initPacket(rxRing, packet.sa)
		if err != nil {
			return fmt.Errorf("Init rx packet error: %s", err)
		}
		packet.rings[rxRing][i] = &RingBuffer{
			queue:  rxq,
			buffer: rxr,
			poll:   C.struct_pollfd{},
			offset: 0,
		}

		txq, txr, err := initPacket(txRing, packet.sa)
		if err != nil {
			return fmt.Errorf("Init tx packet error: %s", err)
		}
		packet.rings[txRing][i] = &RingBuffer{
			queue:  txq,
			buffer: txr,
			poll:   C.struct_pollfd{},
			offset: 0,
		}
	}
	return nil
}

func (packet *Packet) Close() error {
	for i := 0; i < packet.cfg.NumWorkers; i++ {
		if err := syscall.Munmap(packet.rings[rxRing][i].buffer); err != nil {
			return fmt.Errorf("Munmap error: %s", err)
		}
		if err := syscall.Close(packet.rings[rxRing][i].queue); err != nil {
			return fmt.Errorf("Close error: %s", err)
		}
		if err := syscall.Munmap(packet.rings[txRing][i].buffer); err != nil {
			return fmt.Errorf("Munmap error: %s", err)
		}
		if err := syscall.Close(packet.rings[txRing][i].queue); err != nil {
			return fmt.Errorf("Close error: %s", err)
		}
	}
	return nil
}

func (packet *Packet) Recv(buf []byte, queue int) error {
	tph, frame, err := packet.nextFrame(rxRing, queue, C.TP_STATUS_USER, C.POLLIN)
	if err != nil {
		return fmt.Errorf("Get next rx frame error: %s", err)
	}
	copy(buf[:], frame[:])
	tph.tp_status = C.TP_STATUS_KERNEL
	return nil
}

func (packet *Packet) Send(buf []byte, queue int) error {
	tph, frame, err := packet.nextFrame(txRing, queue, C.TP_STATUS_AVAILABLE, C.POLLOUT)
	if err != nil {
		return fmt.Errorf("Get next tx frame error: %s", err)
	}
	copy(frame[:], buf[:])
	tph.tp_len = C.uint(len(buf))
	tph.tp_status = C.TP_STATUS_SEND_REQUEST
	return nil
}

func (packet *Packet) Flush(queue int) error {
	return syscall.Sendto(packet.rings[txRing][queue].queue, nil, 0, packet.sa)
}

func New(cfg *common.Config) *Packet {
	rings := make(map[int][]*RingBuffer)
	rings[rxRing] = make([]*RingBuffer, cfg.NumWorkers)
	rings[txRing] = make([]*RingBuffer, cfg.NumWorkers)
	return &Packet{
		rings: rings,
		cfg:   cfg,
		sa: &syscall.SockaddrLinklayer{
			Protocol: uint16(C.htons(C.ETH_P_IP)),
			Ifindex:  cfg.PublicInterface,
			Pkttype:  C.PACKET_OTHERHOST,
		},
	}
}
