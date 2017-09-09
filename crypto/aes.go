// Copyright (c) 2016-2017 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// SaltLength is the length that the passed in salt slice should be for AES objects.
	SaltLength = 32
	iterations = 10000
)

// AES represents an aes-256-gcm AEAD cipher object.
type AES struct {
	block cipher.Block
	aead  cipher.AEAD
	salt  []byte
}

// EncryptedSize returns the minimum size of the data buffer for encryption, which includes the gcm tag size + nonce size.
func (crypt *AES) EncryptedSize(data []byte) int {
	return len(data) + crypt.aead.Overhead() + crypt.aead.NonceSize()
}

// DecryptedSize returns the minimum size of the data buffer for encryption, which includes the gcm tag size + nonce size.
func (crypt *AES) DecryptedSize(data []byte) int {
	return len(data) - crypt.aead.Overhead() - crypt.aead.NonceSize()
}

// Encrypt takes the data buffer and encrypts up to length bytes in place, while injecting the nonce and gcm tag at the end and signing the additional data.
//
// additional may be nil.
func (crypt *AES) Encrypt(data []byte, length int, additional []byte) (int, error) {
	nonce := make([]byte, crypt.aead.NonceSize())

	n, err := rand.Read(nonce)
	if err != nil {
		return -1, err
	}

	crypt.aead.Seal(data[:0], nonce[:n], data[:length], additional)
	copy(data[length+crypt.aead.Overhead():], nonce[:n])
	return crypt.EncryptedSize(data[:length]), nil
}

// Decrypt takes the data buffer and decrypts it and verifies the additional data.
//
// additional and data must be the same buffers passed to Encrypt.
func (crypt *AES) Decrypt(data []byte, additional []byte) (int, error) {
	length := len(data) - crypt.aead.NonceSize()
	nonce := data[length:]
	_, err := crypt.aead.Open(data[:0], nonce, data[:length], additional)
	return crypt.DecryptedSize(data), err
}

// NewAES returns a new AEAD based cipher object based on the passed in secret and salt.
func NewAES(secret, salt []byte) (*AES, error) {
	key := pbkdf2.Key(secret, salt, iterations, keyLength, sha512.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &AES{
		block: block,
		aead:  aead,
		salt:  salt,
	}, nil
}
