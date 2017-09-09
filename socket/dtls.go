// Copyright (c) 2016-2017 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package socket

import (
	"errors"
	"sync"
	"syscall"

	"github.com/supernomad/quantum/common"
	"github.com/supernomad/quantum/crypto"
)

// DTLS socket struct for managing a multi-queue openssl based DTLS socket.
type DTLS struct {
	cfg     *common.Config
	stop    bool
	queues  []int
	pollFds []int
	events  [][]syscall.EpollEvent
	servers []*crypto.DTLSContext
	clients []*crypto.DTLSContext
	mux     sync.Mutex
	writers []map[string]*crypto.DTLSSession
	readers []map[int32]*crypto.DTLSSession
}

// Close the DTLS socket and removes associated network configuration.
func (dtls *DTLS) Close() error {
	dtls.stop = true

	// Close the DTLS servers.
	if dtls.servers != nil {
		for i := 0; i < dtls.cfg.NumWorkers; i++ {
			dtls.servers[i].Close()
		}
		dtls.servers = nil
	}

	// Close the DTLS clients.
	if dtls.clients != nil {
		for i := 0; i < dtls.cfg.NumWorkers; i++ {
			dtls.clients[i].Close()
		}
		dtls.clients = nil
	}

	// Close the pollFds.
	if dtls.pollFds != nil {
		for i := 0; i < dtls.cfg.NumWorkers; i++ {
			err := syscall.Close(dtls.pollFds[i])
			if err != nil {
				return err
			}
			dtls.events[i] = nil
		}
		dtls.pollFds = nil
		dtls.events = nil
	}

	// Close the DTLS writer sessions.
	if dtls.writers != nil {
		for i := 0; i < dtls.cfg.NumWorkers; i++ {
			for _, session := range dtls.writers[i] {
				session.Close()
			}
		}
		dtls.writers = nil
	}

	// Close the DTLS reader sessions.
	if dtls.readers != nil {
		for i := 0; i < dtls.cfg.NumWorkers; i++ {
			for _, session := range dtls.readers[i] {
				session.Close()
			}
		}
		dtls.readers = nil
	}

	return nil
}

// Queues will return the underlying DTLS socket file descriptors.
func (dtls *DTLS) Queues() []int {
	return dtls.queues
}

// Read a packet off the specified DTLS socket queue and return a *common.Payload representation of the packet.
func (dtls *DTLS) Read(queue int, buf []byte) (*common.Payload, bool) {
	n, err := syscall.EpollWait(dtls.pollFds[queue], dtls.events[queue], -1)
	if err != nil || n < 0 {
		return nil, false
	}

	session, ok := dtls.readers[queue][dtls.events[queue][0].Fd]
	if !ok {
		return nil, false
	}

	read, ok := session.Read(buf)
	if !ok {
		delete(dtls.readers[queue], dtls.events[queue][0].Fd)
		return nil, false
	}

	return common.NewSockPayload(buf, read), true
}

// Write a *common.Payload to the specified DTLS socket queue.
func (dtls *DTLS) Write(queue int, payload *common.Payload, mapping *common.Mapping) bool {
	session, ok := dtls.getWriter(queue, mapping)
	if !ok {
		return false
	}

	wrote, ok := session.Write(payload.Raw[:payload.Length])
	if !ok || wrote != payload.Length {
		delete(dtls.writers[queue], mapping.Address)
		return false
	}

	return true
}

func (dtls *DTLS) handleReader(queue int, session *crypto.DTLSSession) {
	var event syscall.EpollEvent

	event.Events = syscall.EPOLLIN
	event.Fd = int32(session.Fd)

	syscall.EpollCtl(dtls.pollFds[queue], syscall.EPOLL_CTL_ADD, session.Fd, &event)

	dtls.readers[queue][event.Fd] = session
}

func (dtls *DTLS) getWriter(queue int, mapping *common.Mapping) (*crypto.DTLSSession, bool) {
	if session, ok := dtls.writers[queue][mapping.Address]; ok {
		return session, ok
	}

	dtls.mux.Lock()
	defer dtls.mux.Unlock()

	if session, ok := dtls.writers[queue][mapping.Address]; ok {
		return session, ok
	}

	session, err := dtls.clients[queue].Connect(mapping.Address, mapping.Port)
	if err != nil {
		return nil, false
	}

	dtls.writers[queue][mapping.Address] = session
	return session, true
}

func (dtls *DTLS) accept(queue int) {
	for !dtls.stop {
		session, err := dtls.servers[queue].Accept()
		if err != nil {
			continue
		}

		dtls.handleReader(queue, session)
	}
}

func newDTLS(cfg *common.Config) (*DTLS, error) {
	crypto.InitDTLS()

	dtls := &DTLS{
		cfg:     cfg,
		stop:    false,
		queues:  make([]int, cfg.NumWorkers),
		pollFds: make([]int, cfg.NumWorkers),
		events:  make([][]syscall.EpollEvent, cfg.NumWorkers),
		servers: make([]*crypto.DTLSContext, cfg.NumWorkers),
		clients: make([]*crypto.DTLSContext, cfg.NumWorkers),
		writers: make([]map[string]*crypto.DTLSSession, cfg.NumWorkers),
		readers: make([]map[int32]*crypto.DTLSSession, cfg.NumWorkers),
	}

	for i := 0; i < dtls.cfg.NumWorkers; i++ {
		var queue int
		var err error

		if !dtls.cfg.ReuseFDS {
			queue, err = createUDPSocket(dtls.cfg.IsIPv6Enabled, dtls.cfg.ListenAddr)
			if err != nil {
				return dtls, errors.New("error creating the DTLS socket: " + err.Error())
			}
		} else {
			queue = 3 + dtls.cfg.NumWorkers + i
		}

		dtls.queues[i] = queue

		server, err := crypto.NewServerDTLSContext(queue, cfg.ListenIP.String(), cfg.ListenPort, cfg.IsIPv6Enabled, !cfg.DTLSSkipVerify, cfg.DTLSCA, cfg.DTLSCert, cfg.DTLSKey)
		if err != nil {
			return dtls, err
		}

		dtls.servers[i] = server

		client, err := crypto.NewClientDTLSContext(cfg.ListenIP.String(), cfg.IsIPv6Enabled, !cfg.DTLSSkipVerify, cfg.DTLSCA, cfg.DTLSCert, cfg.DTLSKey)
		if err != nil {
			return dtls, err
		}

		dtls.clients[i] = client

		pollFd, err := syscall.EpollCreate1(0)
		if err != nil {
			return dtls, errors.New("Error creating epoll file descriptor: " + err.Error())
		}

		dtls.pollFds[i] = pollFd
		dtls.events[i] = make([]syscall.EpollEvent, 1)

		dtls.writers[i] = make(map[string]*crypto.DTLSSession)
		dtls.readers[i] = make(map[int32]*crypto.DTLSSession)

		go dtls.accept(i)
	}

	return dtls, nil
}
