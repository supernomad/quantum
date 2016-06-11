package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/logger"
)

type GCM struct {
	log *logger.Logger
}

func RandomBytes(b []byte) ([]byte, error) {
	_, err := rand.Read(b)
	return b, err
}

func (gcm *GCM) Seal(payload *common.Payload) (*common.Payload, bool) {
	// Grab a new random nonce
	_, err := RandomBytes(payload.Nonce)
	if err != nil {
		gcm.log.Error("[GCM]", "Error generating random nonce:", err)
		return payload, false
	}

	// Get the block ciper
	block, err := aes.NewCipher(payload.Mapping.SecretKey)
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
	aesgcm.Seal(payload.Packet[:0], payload.Nonce, payload.Packet, nil)
	return payload, true
}

func (gcm *GCM) Unseal(payload *common.Payload) (*common.Payload, bool) {
	// Get the block ciper
	block, err := aes.NewCipher(payload.Mapping.SecretKey)
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

	_, err = aesgcm.Open(payload.Packet[:0], payload.Nonce, payload.Packet, nil)
	if err != nil {
		gcm.log.Error("[GCM]", "Error decrypting/authenticating packet:", err)
		return payload, false
	}

	return payload, true
}

func NewGCM(log *logger.Logger) *GCM {
	return &GCM{log: log}
}
