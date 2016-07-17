package workers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"github.com/Supernomad/quantum/common"
)

var (
	outgoing *Outgoing
	incoming *Incoming
)

var (
	resolveIncomingResult, verifyResult, unsealResult,
	resolveOutgoingResult, sealResult, signResult *common.Payload
)

var resolveIncomingMapping, resolveOutgoingMapping, testMapping *common.Mapping
var intRemoteIP, intLocalIP uint32

func init() {
	key := make([]byte, 32)
	rand.Read(key)
	block, _ := aes.NewCipher(key)
	aesgcm, _ := cipher.NewGCM(block)

	remoteIP := "10.8.0.1"
	localIP := "10.8.0.2"

	intRemoteIP = common.IPtoInt(remoteIP)
	intLocalIP = common.IPtoInt(localIP)

	outgoing = NewOutgoing(nil, "10.8.0.2", nil, nil, nil)
	incoming = NewIncoming(nil, "10.8.0.1", nil, nil, nil)

	testMapping = &common.Mapping{PublicKey: make([]byte, 32), SecretKey: key, Cipher: aesgcm}

	outgoing.Mappings = make(map[uint32]*common.Mapping)
	outgoing.Mappings[intRemoteIP] = testMapping
	outgoing.Mappings[intLocalIP] = testMapping

	incoming.Mappings = make(map[uint32]*common.Mapping)
	incoming.Mappings[intRemoteIP] = testMapping
	incoming.Mappings[intLocalIP] = testMapping

	for i := 5; i < 10000; i++ {
		outgoing.Mappings[intRemoteIP+uint32(i)] = &common.Mapping{PublicKey: make([]byte, 32)}
		incoming.Mappings[intRemoteIP+uint32(i)] = &common.Mapping{PublicKey: make([]byte, 32)}
	}
}
