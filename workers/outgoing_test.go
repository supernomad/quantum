package workers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/etcd"
	"testing"
)

var out *Outgoing = NewOutgoing(nil, "10.8.0.2", nil, nil, nil)

var sealResult *common.Payload

var resolveResult *common.Payload
var resolveMapping *common.Mapping

func benchmarkSeal(payload *common.Payload, mapping *common.Mapping, b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if payload, pass := out.Seal(payload, mapping); pass {
			sealResult = payload
		} else {
			panic("Seal failed something is wrong.")
		}
	}
}

func BenchmarkSeal(b *testing.B) {
	buf := make([]byte, common.MaxPacketLength)
	key := make([]byte, 32)

	rand.Read(buf)
	payload := common.NewTunPayload(buf, 1500)

	rand.Read(key)
	block, _ := aes.NewCipher(key)
	aesgcm, _ := cipher.NewGCM(block)
	mapping := &common.Mapping{SecretKey: key, Cipher: aesgcm}

	benchmarkSeal(payload, mapping, b)
}

func benchmarkResolve(payload *common.Payload, b *testing.B) {
	b.ResetTimer()
	// really compiler optimizations... really
	var m *common.Mapping
	var p *common.Payload

	for n := 0; n < b.N; n++ {
		if payload, mapping, pass := out.Resolve(payload); pass {
			m = mapping
			p = payload
		} else {
			panic("Resolve failed something is wrong.")
		}
	}
	resolveResult = p
	resolveMapping = m
}

func BenchmarkResolve(b *testing.B) {
	remoteIp := "10.8.0.1"
	intRemoteIp := etcd.IP4toInt(remoteIp)

	out.Mappings = make(map[uint32]*common.Mapping)
	out.Mappings[intRemoteIp] = &common.Mapping{PublicKey: make([]byte, 32)}

	for i := 0; i < 10000; i++ {
		out.Mappings[intRemoteIp+uint32(i)] = &common.Mapping{PublicKey: make([]byte, 32)}
	}

	buf := make([]byte, common.MaxPacketLength)
	payload := common.NewTunPayload(buf, 1500)

	binary.LittleEndian.PutUint32(payload.Packet[16:20], intRemoteIp)
	benchmarkResolve(payload, b)
}
