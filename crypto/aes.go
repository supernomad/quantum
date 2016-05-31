package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/logger"
)

const (
	ivLength   = aes.BlockSize
	padLength  = 4
	metaLength = ivLength + keyLength + padLength
)

const (
	// IV offsets
	ivStart = keyEnd
	ivEnd   = ivStart + ivLength

	// Key offsets
	keyStart = padEnd
	keyEnd   = keyStart + keyLength

	// Padding count offsets
	padStart = 0
	padEnd   = padLength

	// Payload offset
	payloadStart = ivEnd
)

type AES struct {
	log  *logger.Logger
	ecdh *ECDH
}

func RandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}

func (crypt *AES) Encrypt(payload common.Payload) common.Payload {
	// Get the padding size
	padding := aes.BlockSize - len(payload.Packet)%aes.BlockSize

	// Generate key based on ecdh public/private keypair
	pubKey, err := DeserializeKey(payload.PublicKey)
	if err != nil {
		crypt.log.Error("[AES]", "Error deserializing public key: public key length mismatch")
		return payload
	}
	key := crypt.ecdh.GenerateSharedSecret(pubKey)

	// Grab a new random iv
	iv, err := RandomBytes(ivLength)
	if err != nil {
		crypt.log.Error("[AES]", "Error generating random iv:", err)
		return payload
	}

	// Get the block ciper and encrypter
	block, err := aes.NewCipher(key)
	if err != nil {
		crypt.log.Error("[AES]", "Error generating ciper:", err)
		return payload
	}

	mode := cipher.NewCBCEncrypter(block, iv)

	// Generate the ciphertext array
	ciphertextLength := len(payload.Packet) + padding + metaLength
	ciphertext := make([]byte, ciphertextLength)

	binary.LittleEndian.PutUint32(ciphertext[padStart:padEnd], uint32(padding))

	copy(ciphertext[keyStart:keyEnd], crypt.ecdh.PublicKey[:])
	copy(ciphertext[ivStart:ivEnd], iv)
	copy(ciphertext[payloadStart:], payload.Packet)

	mode.CryptBlocks(ciphertext[payloadStart:], ciphertext[payloadStart:])

	payload.Packet = ciphertext[:]
	return payload
}

func (crypt *AES) Decrypt(payload common.Payload) common.Payload {
	padding := int(binary.LittleEndian.Uint32(payload.Packet[padStart:padEnd]))

	pubKey := payload.Packet[keyStart:keyEnd]
	key := crypt.ecdh.GenerateSharedSecret(pubKey)

	iv := payload.Packet[ivStart:ivEnd]

	block, err := aes.NewCipher(key)
	if err != nil {
		crypt.log.Error("Error generating ciper:", err)
		return payload
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	mode.CryptBlocks(payload.Packet[payloadStart:], payload.Packet[payloadStart:])

	payload.Packet = payload.Packet[payloadStart : len(payload.Packet)-padding]
	return payload
}

func NewAES(log *logger.Logger, ecdh *ECDH) *AES {
	return &AES{log: log, ecdh: ecdh}
}
