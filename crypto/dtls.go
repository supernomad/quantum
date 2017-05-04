// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package crypto

/*
// Need the openssl library binary blobs from the included submodule in 'quantum/vendor/openssl'.
#cgo LDFLAGS: ${SRCDIR}/../vendor/openssl/libssl.a ${SRCDIR}/../vendor/openssl/libcrypto.a -ldl
#cgo CFLAGS: -g -Wno-deprecated -I${SRCDIR}/../vendor/openssl/include

// Include the dtls glue for cgo <-> go
#include <dtls.h>
*/
import "C"

import (
	"errors"
	"sync/atomic"
	"unsafe"
)

const (
	errorLen             = 120
	initialized    int32 = 1
	notInitialized int32 = 0
)

var state = notInitialized

func btoi(b bool) C.int {
	if b {
		return 1
	}
	return 0
}

func generateErrorStr() *C.char {
	buff := make([]byte, errorLen)
	return C.CString(string(buff))
}

// InitDTLS setups and configures the openssl libraries.
func InitDTLS() {
	if atomic.CompareAndSwapInt32(&state, notInitialized, initialized) {
		// Need to init the openssl libraries.
		C.init_dtls()
	}
}

// DestroyDTLS will safely terminate and free all openssl data.
func DestroyDTLS() {
	if atomic.CompareAndSwapInt32(&state, initialized, notInitialized) {
		// Need to destroy the openssl data.
		C.destroy_dtls()
	}
}

// DTLSContext is a wrapper around a cgo struct implementing a DTLS context via openssl.
type DTLSContext struct {
	// cgo binding for the dtls object.
	ctx *C.Context
}

// Accept will handle opening new DTLS sessions from remote nodes.
func (dtls *DTLSContext) Accept() (*DTLSSession, error) {
	// Get the various converted strings.
	err := generateErrorStr()

	// Ensure the C based strings get freed to mitigate any kind of memory leak.
	defer C.free(unsafe.Pointer(err))

	session := C.accept_dtls(dtls.ctx, err)

	if session == nil {
		return nil, errors.New(C.GoString(err))
	}

	return &DTLSSession{
		session: session,
		Fd:      int(C.get_dtls_fd(session)),
	}, nil
}

// Connect will handle opening a new DTLS session with a remote node.
func (dtls *DTLSContext) Connect(addr string, port int) (*DTLSSession, error) {
	// Get the various converted strings.
	err := generateErrorStr()
	addrstr := C.CString(addr)

	// Ensure the C based strings get freed to mitigate any kind of memory leak.
	defer C.free(unsafe.Pointer(err))
	defer C.free(unsafe.Pointer(addrstr))

	session := C.connect_dtls(dtls.ctx, addrstr, C.int(port), err)

	if session == nil {
		return nil, errors.New(C.GoString(err))
	}

	return &DTLSSession{
		session: session,
		Fd:      int(C.get_dtls_fd(session)),
	}, nil
}

// Close destroys all traces of the DTLS struct.
func (dtls *DTLSContext) Close() {
	// Call into cgo to destroy the context using the openssl free/shutdown functions.
	C.free_dtls_context(dtls.ctx)
}

// NewServerDTLSContext creates a new server based DTLS struct which is ready to accept connections from remote nodes.
func NewServerDTLSContext(fd int, addr string, port int, useV6 bool, verifyPeer bool, ca string, cert string, key string) (*DTLSContext, error) {
	// Get the various converted strings.
	err := generateErrorStr()
	addrstr := C.CString(addr)
	castr := C.CString(ca)
	certstr := C.CString(cert)
	keystr := C.CString(key)

	// Ensure the C based strings get freed to mitigate any kind of memory leak.
	defer C.free(unsafe.Pointer(err))
	defer C.free(unsafe.Pointer(addrstr))
	defer C.free(unsafe.Pointer(castr))
	defer C.free(unsafe.Pointer(certstr))
	defer C.free(unsafe.Pointer(keystr))

	ctx := C.init_server_dtls_context(C.int(fd), addrstr, C.int(port), btoi(useV6), btoi(verifyPeer), castr, certstr, keystr, err)

	if ctx == nil {
		return nil, errors.New(C.GoString(err))
	}

	return &DTLSContext{
		ctx: ctx,
	}, nil
}

// NewClientDTLSContext creates a new client based DTLS struct which is ready to connect to remote nodes.
func NewClientDTLSContext(addr string, useV6 bool, verifyPeer bool, ca string, cert string, key string) (*DTLSContext, error) {
	// Get the various converted strings.
	err := generateErrorStr()
	addrstr := C.CString(addr)
	castr := C.CString(ca)
	certstr := C.CString(cert)
	keystr := C.CString(key)

	// Ensure the C based strings get freed to mitigate any kind of memory leak.
	defer C.free(unsafe.Pointer(err))
	defer C.free(unsafe.Pointer(addrstr))
	defer C.free(unsafe.Pointer(castr))
	defer C.free(unsafe.Pointer(certstr))
	defer C.free(unsafe.Pointer(keystr))

	ctx := C.init_client_dtls_context(addrstr, btoi(useV6), btoi(verifyPeer), castr, certstr, keystr, err)

	if ctx == nil {
		return nil, errors.New(C.GoString(err))
	}

	return &DTLSContext{
		ctx: ctx,
	}, nil
}

// DTLSSession is a wrapper around a cgo struct implementing a DTLS session via openssl.
type DTLSSession struct {
	// cgo binding for the session object.
	session *C.Session
	Fd      int
}

// Read will read bytes from the session up to the size of the provided buffer.
func (session *DTLSSession) Read(buf []byte) (int, bool) {
	// Read on to the supplied buffer from the underlying SSL BIO.
	read := C.read_dtls(session.session, unsafe.Pointer(&buf[0]), C.int(len(buf)))

	if read <= 0 {
		return int(read), false
	}
	return int(read), true
}

// Write will write the bytes from the provided buffer to the session.
func (session *DTLSSession) Write(buf []byte) (int, bool) {
	// Write the supplied buffer on to the underlying SSL BIO.
	wrote := C.write_dtls(session.session, unsafe.Pointer(&buf[0]), C.int(len(buf)))

	if wrote <= 0 {
		return int(wrote), false
	}
	return int(wrote), true
}

// Close destroys all traces of the DTLSSession struct.
func (session *DTLSSession) Close() {
	// Call into cgo to destroy the session using the openssl free/shutdown functions/
	C.free_dtls_session(session.session)
}
