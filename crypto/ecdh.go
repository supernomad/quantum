package crypto

import (
	"encoding/base64"
	"github.com/Supernomad/quantum/logger"
	"golang.org/x/crypto/curve25519"
)

const (
	keyLength = 32
)

type ECDH struct {
	log        *logger.Logger
	PublicKey  []byte
	PrivateKey []byte
}

func NewEcdh(log *logger.Logger) (*ECDH, error) {
	var pub, priv [keyLength]byte

	rand, err := RandomBytes(keyLength)
	if err != nil {
		return nil, err
	}

	copy(priv[:], rand)
	curve25519.ScalarBaseMult(&pub, &priv)

	return &ECDH{
		log:        log,
		PrivateKey: priv[:],
		PublicKey:  pub[:],
	}, nil
}

func (e *ECDH) GenerateSharedSecret(pubKey []byte) []byte {
	var secret, pub, priv [keyLength]byte

	copy(pub[:], pubKey)
	copy(priv[:], e.PrivateKey)

	curve25519.ScalarMult(&secret, &priv, &pub)
	return secret[:]
}

func SerializeKey(key []byte) string {
	return base64.StdEncoding.EncodeToString(key[:])
}

func DeserializeKey(serialized string) ([]byte, error) {
	var key [keyLength]byte

	deserialized, err := base64.StdEncoding.DecodeString(serialized)
	if err != nil {
		return nil, err
	}

	copy(key[:], deserialized)
	return key[:], nil
}
