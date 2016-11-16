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

	intRemoteIP, intLocalIP uint32
)

func init() {
	remoteIP := "10.8.0.1"
	localIP := "10.8.0.2"

	store = &backend.Mock{
		Mappings: make(map[uint32]*common.Mapping),
	}
	tun = inet.New(inet.MOCKInterface, nil)
	sock = socket.New(socket.MOCKSocket, nil)

	key := make([]byte, 32)
	rand.Read(key)
	block, _ := aes.NewCipher(key)
	aesgcm, _ := cipher.NewGCM(block)

	intRemoteIP = common.IPtoInt(remoteIP)
	intLocalIP = common.IPtoInt(localIP)

	testMapping = &common.Mapping{PublicKey: make([]byte, 32), Cipher: aesgcm}

	store.Mappings[intRemoteIP] = testMapping
	store.Mappings[intLocalIP] = testMapping

	for i := 5; i < 10000; i++ {
		store.Mappings[intRemoteIP+uint32(i)] = &common.Mapping{PublicKey: make([]byte, 32)}
	}

	incoming = NewIncoming(&common.Config{NumWorkers: 1, PrivateIP: remoteIP}, store, tun, sock)
	outgoing = NewOutgoing(&common.Config{NumWorkers: 1, PrivateIP: localIP}, store, tun, sock)
}
