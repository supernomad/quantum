package socket

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
	"encoding/binary"
	"fmt"
	"github.com/Supernomad/quantum/common"
	"golang.org/x/net/ipv4"
	"syscall"
	"time"
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
	id       int
	mmaps    [][]byte
	queues   []int
	flushers []chan struct{}
	rings    map[int][]ringBuffer
	cfg      *common.Config
	sa       *syscall.SockaddrLinklayer
}

type ringBuffer struct {
	raw    []byte
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

func (sock *Packet) Name() string {
	return sock.cfg.ListenAddress
}

func (sock *Packet) Open() error {
	for i := 0; i < sock.cfg.NumWorkers; i++ {
		queue, raw, rx, tx, err := initPacket(sock.sa)
		if err != nil {
			return err
		}

		sock.queues[i] = queue
		sock.mmaps[i] = raw
		sock.flushers[i] = make(chan struct{}, numBlocks*framesPerBlock)
		sock.rings[rxRing][i] = ringBuffer{
			raw:    rx,
			poll:   C.struct_pollfd{},
			offset: 0,
		}
		sock.rings[txRing][i] = ringBuffer{
			raw:    tx,
			poll:   C.struct_pollfd{},
			offset: 0,
		}
		go sock.flush(i)
	}
	return nil
}

func (sock *Packet) Close() error {
	for i := 0; i < sock.cfg.NumWorkers; i++ {
		if err := syscall.Munmap(sock.mmaps[i]); err != nil {
			return err
		}
		if err := syscall.Close(sock.queues[i]); err != nil {
			return err
		}
	}
	return nil
}

func (sock *Packet) GetFDs() []int {
	return sock.queues
}

func (sock *Packet) poll(rtype int, queue int, events C.short) error {
	sock.rings[rtype][queue].poll.fd = C.int(sock.queues[queue])
	sock.rings[rtype][queue].poll.events = C.short(events)
	sock.rings[rtype][queue].poll.revents = 0
	_, err := C.poll(&sock.rings[rtype][queue].poll, 1, -1)
	return err
}

func (sock *Packet) Read(buf []byte, queue int) (*common.Payload, bool) {
	pos := sock.rings[rxRing][queue].offset * frameSize
	tph := (*C.struct_tpacket_hdr)(unsafe.Pointer(&sock.rings[rxRing][queue].raw[pos]))

	for tph.tp_status&C.TP_STATUS_USER != C.TP_STATUS_USER {
		err := sock.poll(rxRing, queue, C.POLLIN)
		if err != nil {
			return nil, false
		}
	}

	loc := pos + int(tph.tp_net)
	copy(buf, sock.rings[rxRing][queue].raw[loc:loc+int(tph.tp_snaplen)])
	tph.tp_status = C.TP_STATUS_KERNEL

	sock.rings[rxRing][queue].offset++
	if sock.rings[rxRing][queue].offset >= framesPerBlock*numBlocks {
		sock.rings[rxRing][queue].offset = 0
	}

	return common.NewIPPayload(buf), true
}

func checksum(buf []byte) {
	buf[10] = 0
	buf[11] = 0

	var csum uint32
	for i := 0; i < len(buf); i += 2 {
		csum += uint32(buf[i]) << 8
		csum += uint32(buf[i+1])
	}
	for {
		if csum <= 65535 {
			break
		}
		csum = (csum >> 16) + uint32(uint16(csum))
	}
	binary.BigEndian.PutUint16(buf[10:12], ^uint16(csum))
}

func (sock *Packet) Write(payload *common.Payload, mapping *common.Mapping, queue int) bool {
	pos := sock.rings[txRing][queue].offset * frameSize
	tph := (*C.struct_tpacket_hdr)(unsafe.Pointer(&sock.rings[txRing][queue].raw[pos]))
	for tph.tp_status != C.TP_STATUS_AVAILABLE {
		err := sock.poll(txRing, queue, C.POLLOUT)
		if err != nil {
			return false
		}
	}

	iph := &ipv4.Header{
		Version:  4,
		Len:      20,
		TOS:      0,
		TotalLen: payload.Length,
		ID:       sock.id,
		Flags:    ipv4.DontFragment,
		FragOff:  0,
		TTL:      64,
		Protocol: 138,
		Checksum: 0,
		Src:      sock.cfg.PublicIPAddr,
		Dst:      mapping.Addr,
		Options:  nil,
	}
	sock.id++

	buf, _ := iph.Marshal()
	checksum(buf)
	copy(payload.IPHeader[:], buf[:])

	var sll C.struct_sockaddr_ll
	loc := pos + C.TPACKET_HDRLEN - int(unsafe.Sizeof(sll))

	copy(sock.rings[txRing][queue].raw[loc:loc+payload.Length], payload.Raw[:payload.Length])

	tph.tp_len = C.uint(payload.Length)
	tph.tp_status = C.TP_STATUS_SEND_REQUEST

	sock.rings[txRing][queue].offset++
	if sock.rings[txRing][queue].offset >= framesPerBlock*numBlocks {
		sock.rings[txRing][queue].offset = 0
	}

	if sock.rings[txRing][queue].offset%2 == 0 {
		sock.flushers[queue] <- struct{}{}
	}
	return true
}

func (sock *Packet) flush(queue int) {
	ticker := time.NewTicker(time.Millisecond * 10)
	for {
		select {
		case <-sock.flushers[queue]:
			syscall.Sendto(sock.queues[queue], nil, 0, sock.sa)
		case <-ticker.C:
			syscall.Sendto(sock.queues[queue], nil, 0, sock.sa)
		}
	}
}

func newPacket(cfg *common.Config) *Packet {
	mmaps := make([][]byte, cfg.NumWorkers)
	queues := make([]int, cfg.NumWorkers)
	flushers := make([]chan struct{}, cfg.NumWorkers)
	rings := make(map[int][]ringBuffer)
	rings[rxRing] = make([]ringBuffer, cfg.NumWorkers)
	rings[txRing] = make([]ringBuffer, cfg.NumWorkers)

	return &Packet{
		mmaps:    mmaps,
		queues:   queues,
		rings:    rings,
		flushers: flushers,
		id:       0,
		cfg:      cfg,
		sa: &syscall.SockaddrLinklayer{
			Protocol: uint16(C.htons(C.ETH_P_IP)),
			Ifindex:  cfg.PublicInterface,
			Pkttype:  C.PACKET_OTHERHOST,
		},
	}
}

func initPacket(sa *syscall.SockaddrLinklayer) (int, []byte, []byte, []byte, error) {
	// create the socket
	queue, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_DGRAM, C.ETH_P_IP)
	if err != nil {
		return -1, nil, nil, nil, fmt.Errorf("Create socket error: %v", err)
	}

	// bind the socket to the public interface
	err = syscall.Bind(queue, sa)
	if err != nil {
		return -1, nil, nil, nil, fmt.Errorf("Bind socket error: %v", err)
	}

	// set the socket C.PACKET_FANOUT options
	fanoutVar := (C.PACKET_FANOUT_LB | C.PACKET_FANOUT_FLAG_ROLLOVER | C.PACKET_FANOUT_FLAG_DEFRAG) << 16
	err = syscall.SetsockoptInt(queue, syscall.SOL_PACKET, C.PACKET_FANOUT, fanoutVar)
	if err != nil {
		return -1, nil, nil, nil, fmt.Errorf("Fanout option error: %v", err)
	}

	// set the base bpf filter to only get quantum traffic
	_, err = C.setsockopt(C.int(queue), C.SOL_SOCKET, C.int(C.SO_ATTACH_FILTER), unsafe.Pointer(&filter), C.socklen_t(unsafe.Sizeof(filter)))
	if err != nil {
		return -1, nil, nil, nil, fmt.Errorf("Bpf option error: %v", err)
	}

	// setup the rx ring buffer
	_, err = C.setsockopt(C.int(queue), C.SOL_PACKET, C.int(rxRing), unsafe.Pointer(&tpreq), C.socklen_t(unsafe.Sizeof(tpreq)))
	if err != nil {
		return -1, nil, nil, nil, fmt.Errorf("RxRing option error: %v", err)
	}

	// setup the tx ring buffer
	_, err = C.setsockopt(C.int(queue), C.SOL_PACKET, C.int(txRing), unsafe.Pointer(&tpreq), C.socklen_t(unsafe.Sizeof(tpreq)))
	if err != nil {
		return -1, nil, nil, nil, fmt.Errorf("TxRing option error: %v", err)
	}

	// generate the actual memory buffer for the rx/tx ring buffers
	raw, err := syscall.Mmap(queue, 0, totalSize*2, C.PROT_READ|C.PROT_WRITE, C.MAP_SHARED)
	if err != nil {
		return -1, nil, nil, nil, fmt.Errorf("Mmap error: %v", err)
	}
	return queue, raw, raw[:totalSize], raw[totalSize:], nil
}
