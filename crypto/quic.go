// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package crypto

import (
	"crypto/tls"
	"net"
	"os"

	quic "github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/protocol"
)

func fdToPacket(fd int) (net.PacketConn, error) {
	f, err := os.NewFile(fd, "")
	if err != nil {
		return nil, err
	}

	return net.FilePacketConn(f)
}

// QuicServer is a wrapper around the quic protocol client/server implementation.
type QuicServer struct {
	quicCfg  *quic.Config
	tlsCfg   *tls.Config
	listener quic.Listener
}

// Accept will block until a remote peer initiates a new session with this node, and returns the session and an error if any.
func (ctx *QuicServer) Accept() (*QuicSession, error) {
	session, err := ctx.listener.Accept()
	if err != nil {
		return nil, err
	}
	stream, err := session.AcceptStream()
	if err != nil {
		return nil, err
	}

	return &QuicSession{
		session: session,
		stream:  stream,
	}, nil
}

// Connect will block until this peer can initiate a new session with a remote peer, and returns the session and an error if any.
func (ctx *QuicServer) Connect(fd int, addr string) (*QuicSession, error) {
	pconn, err := fdToPacket(fd)
	if err != nil {
		return nil, err
	}

	remoteAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	session, err := quic.Dial(pconn, remoteAddr, "", ctx.tlsCfg, ctx.quicCfg)
	if err != nil {
		return nil, err
	}

	stream, err := session.OpenStreamSync()
	if err != nil {
		return nil, err
	}

	return &QuicSession{
		session: session,
		stream:  stream,
	}, nil
}

// Close will fully destroy this QuicServer, and return an error if any.
func (ctx *QuicServer) Close() error {
	return ctx.listener.Close()
}

// NewServerQuicServer returns a new server based quic context.
func NewQuicServer(fd int, skipVerify bool, caFile string, certFile string, keyFile string) (*QuicServer, error) {
	quicCfg := &quic.Config{
		KeepAlive: true,
		Versions:  protocol.Version37,
	}
	tlsCfg := &tls.Config{}

	if certFile != "" && keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, errors.New("error reading the supplied tls certificate and/or key: " + err.Error())
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
		tlsCfg.BuildNameToCertificate()
	}

	tlsCfg.InsecureSkipVerify = skipVerify

	if caFile != "" {
		cert, err := ioutil.ReadFile(caFile)
		if err != nil {
			return nil, errors.New("error reading the supplied tls ca certificate: " + err.Error())
		}
		tlsCfg.RootCAs = x509.NewCertPool()
		tlsCfg.RootCAs.AppendCertsFromPEM(cert)
		tlsCfg.BuildNameToCertificate()
	}

	pconn, err := fdToPacket(fd)
	if err != nil {
		return nil, err
	}

	listener, err := quic.Listen(pconn, tlsConf, quicCfg)
	if err != nil {
		return nil, err
	}

	return &QuicServer{
		quicCfg:  quicCfg,
		tlsCfg:   tlsCfg,
		listener: listener,
	}, nil
}

// QuicSession is a wrapper around the quic protocol stream implementation.
type QuicSession struct {
	session quic.Session
	stream  quic.Stream
}

// Read will read up to len(buf) bytes from the remote peer into the supplied slice, and returns the number of bytes read and an error if any.
func (session *QuicSession) Read(buf []byte) (int, error) {
	return session.stream.Read(buf)
}

// Write will write up to len(buf) byte to the remote peer from the supplied slice, and returns the number of bytes written and an error if any.
func (session *QuicSession) Write(buf []byte) (int, error) {
	return session.stream.Write(buf)
}

// Close will fully destroy this QuicSession, and return an error if any.
func (session *QuicSession) Close() error {
	if err := session.session.Close(); err != nil {
		return err
	}
	return session.stream.Close()
}
