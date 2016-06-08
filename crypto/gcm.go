package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/logger"
)

type GCM struct {
	log  *logger.Logger
	ecdh *ECDH
}

func RandomBytes(b []byte) ([]byte, error) {
	_, err := rand.Read(b)
	return b, err
}

func (gcm *GCM) Seal(payload *common.Payload) (*common.Payload, bool) {
	// Generate the shared key
	key := gcm.ecdh.GenerateSharedSecret(payload.PublicKey)

	// Grab a new random nonce
	_, err := RandomBytes(payload.Nonce)
	if err != nil {
		gcm.log.Error("[GCM]", "Error generating random nonce:", err)
		return payload, false
	}

	// Get the block ciper
	block, err := aes.NewCipher(key)
	if err != nil {
		gcm.log.Error("[GCM]", "Error generating ciper:", err)
		return payload, false
	}

	// Get the GCM block wrapper
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		gcm.log.Error("[GCM]", "Error generating gcm wrapper:", err)
		return payload, false
	}

	// Seal the packet and associated meta data
	aesgcm.Seal(payload.Packet[:0], payload.Nonce, payload.Packet, gcm.ecdh.PublicKey[:])

	copy(payload.Key, gcm.ecdh.PublicKey[:])
	return payload, true
}

func (gcm *GCM) Unseal(payload *common.Payload) (*common.Payload, bool) {
	// Generate the shared key
	key := gcm.ecdh.GenerateSharedSecret(payload.Key)

	// Get the block ciper
	block, err := aes.NewCipher(key)
	if err != nil {
		gcm.log.Error("[GCM]", "Error generating ciper:", err)
		return payload, false
	}

	// Get the GCM block wrapper
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		gcm.log.Error("[GCM]", "Error generating gcm wrapper:", err)
		return payload, false
	}

	_, err = aesgcm.Open(payload.Packet[:0], payload.Nonce, payload.Packet, payload.Key)
	if err != nil {
		gcm.log.Error("[GCM]", "Error decrypting/authenticating packet:", err)
		return payload, false
	}

	return payload, true
}

func NewGCM(log *logger.Logger, ecdh *ECDH) *GCM {
	return &GCM{log: log, ecdh: ecdh}
}
