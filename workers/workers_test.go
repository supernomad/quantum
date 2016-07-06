package workers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/etcd"
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
var intRemoteIp, intLocalIp uint32

func init() {
	key := make([]byte, 32)
	rand.Read(key)
	block, _ := aes.NewCipher(key)
	aesgcm, _ := cipher.NewGCM(block)

	remoteIp := "10.8.0.1"
	localIp := "10.8.0.2"

	intRemoteIp = etcd.IP4toInt(remoteIp)
	intLocalIp = etcd.IP4toInt(localIp)

	outgoing = NewOutgoing(nil, "10.8.0.2", nil, nil, nil)
	incoming = NewIncoming(nil, "10.8.0.1", nil, nil, nil)

	testMapping = &common.Mapping{PublicKey: make([]byte, 32), SecretKey: key, Cipher: aesgcm}

	outgoing.Mappings = make(map[uint32]*common.Mapping)
	outgoing.Mappings[intRemoteIp] = testMapping
	outgoing.Mappings[intLocalIp] = testMapping

	incoming.Mappings = make(map[uint32]*common.Mapping)
	incoming.Mappings[intRemoteIp] = testMapping
	incoming.Mappings[intLocalIp] = testMapping

	for i := 5; i < 10000; i++ {
		outgoing.Mappings[intRemoteIp+uint32(i)] = &common.Mapping{PublicKey: make([]byte, 32)}
		incoming.Mappings[intRemoteIp+uint32(i)] = &common.Mapping{PublicKey: make([]byte, 32)}
	}
}
