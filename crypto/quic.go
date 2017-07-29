// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package crypto

import (
	"crypto/tls"

	quic "github.com/lucas-clemente/quic-go"
)

// QuicContext is a wrapper around the quic protocol client/server implementation.
type QuicContext struct {
	quicCfg *quic.Config
	tlsCfg  *tls.Config
}

// Accept will block until a remote peer initiates a new session with this node, and returns the session and an error if any.
func (ctx *QuicContext) Accept() (*QuicSession, error) {
	return nil, nil
}

// Connect will block until this peer can initiate a new session with a remote peer, and returns the session and an error if any.
func (ctx *QuicContext) Connect() (*QuicSession, error) {
	return nil, nil
}

// Close will fully destroy this QuicContext, and return an error if any.
func (ctx *QuicContext) Close() error {
	return nil
}

// NewServerQuicContext returns a new server based quic context.
func NewServerQuicContext() (*QuicContext, error) {
	return nil, nil
}

// NewClientQuicContext returns a new client based quic context.
func NewClientQuicContext() (*QuicContext, error) {
	return nil, nil
}

// QuicSession is a wrapper around the quic protocol stream implementation.
type QuicSession struct {
}

// Read will read up to len(buf) bytes from the remote peer into the supplied slice, and returns the number of bytes read and an error if any.
func (session *QuicSession) Read(buf []byte) (int, error) {
	return -1, nil
}

// Write will write up to len(buf) byte to the remote peer from the supplied slice, and returns the number of bytes written and an error if any.
func (session *QuicSession) Write(buf []byte) (int, error) {
	return -1, nil
}

// Close will fully destroy this QuicSession, and return an error if any.
func (session *QuicSession) Close() error {
	return nil
}
