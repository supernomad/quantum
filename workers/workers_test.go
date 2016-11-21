package workers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"github.com/Supernomad/quantum/backend"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/inet"
	"github.com/Supernomad/quantum/socket"
	"sync"
)

var (
	outgoing  *Outgoing
	incoming  *Incoming
	tun       inet.Interface
	sock      socket.Socket
	store     *backend.Mock
	privateIP = "10.1.1.1"
	wg        = &sync.WaitGroup{}
)

var (
	resolveIncomingResult, verifyResult, unsealResult,
	resolveOutgoingResult, sealResult, signResult *common.Payload

	resolveIncomingMapping, resolveOutgoingMapping, testMapping *common.Mapping
)

func init() {
	ip := "10.8.0.1"

	store = &backend.Mock{}
	tun = inet.New(inet.MOCKInterface, nil)
	sock = socket.New(socket.MOCKSocket, nil)

	key := make([]byte, 32)
	rand.Read(key)

	block, _ := aes.NewCipher(key)
	aesgcm, _ := cipher.NewGCM(block)

	testMapping = &common.Mapping{PublicKey: make([]byte, 32), Cipher: aesgcm}

	store.Mapping = testMapping

	incoming = NewIncoming(&common.Config{NumWorkers: 1, PrivateIP: ip}, store, tun, sock)
	outgoing = NewOutgoing(&common.Config{NumWorkers: 1, PrivateIP: ip}, store, tun, sock)
}
