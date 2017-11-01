package bitmarksdk

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/ed25519"
)

const (
	AlgChaCha20Poly1305 = "chacha20poly1305"
)

type DataKey interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
	Bytes() []byte
	Algorithm() string
}

type ChaCha20DataKey struct {
	key []byte
}

func newChaCha20DataKey() (*ChaCha20DataKey, error) {
	key := make([]byte, chacha20poly1305.KeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}

	return &ChaCha20DataKey{key: key}, nil
}

// Encrypt the plaintext using zero nonce
func (k *ChaCha20DataKey) Encrypt(plaintext []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(k.key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, chacha20poly1305.NonceSize)

	ciphertext := aead.Seal(nil, nonce, plaintext, nil)

	return ciphertext, nil
}

// Decrypt the ciphertext using zero nonce
func (k *ChaCha20DataKey) Decrypt(ciphertext []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(k.key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, chacha20poly1305.NonceSize)

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func (k *ChaCha20DataKey) Bytes() []byte {
	return k.key
}

func (k *ChaCha20DataKey) Algorithm() string {
	return AlgChaCha20Poly1305
}

func NewDataKey() (DataKey, error) {
	return newChaCha20DataKey()
}

type SessionData struct {
	EncryptedDataKey          []byte
	EncryptedDataKeySignature []byte
	DataKeySignature          []byte
	DataKeyAlgorithm          string
}

type encodedSessionData struct {
	EncryptedDataKey          string `json:"enc_data_key"`
	EncryptedDataKeySignature string `json:"enc_data_key_sig"`
	DataKeySignature          string `json:"data_key_sig"`
	DataKeyAlgorithm          string `json:"data_key_alg"`
}

func (s *SessionData) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		&encodedSessionData{
			EncryptedDataKey:          hex.EncodeToString(s.EncryptedDataKey),
			EncryptedDataKeySignature: hex.EncodeToString(s.EncryptedDataKeySignature),
			DataKeySignature:          hex.EncodeToString(s.DataKeySignature),
			DataKeyAlgorithm:          s.DataKeyAlgorithm,
		})
}

func (s *SessionData) UnmarshalJSON(data []byte) error {
	var aux encodedSessionData
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	s.EncryptedDataKey, _ = hex.DecodeString(aux.EncryptedDataKey)
	s.EncryptedDataKeySignature, _ = hex.DecodeString(aux.EncryptedDataKeySignature)
	s.DataKeySignature, _ = hex.DecodeString(aux.DataKeySignature)
	s.DataKeyAlgorithm = aux.DataKeyAlgorithm
	return nil
}

func (s SessionData) String() string {
	b, _ := json.Marshal(&s)
	return string(b)
}

func createSessionData(acct *Account, key DataKey, recipientEncrPubkey []byte) (*SessionData, error) {
	encrDataKey, err := acct.EncrKey.Encrypt(key.Bytes(), recipientEncrPubkey)
	if err != nil {
		return nil, fmt.Errorf("unable to encrypt the data key: %v", err)
	}
	return &SessionData{
		EncryptedDataKey:          encrDataKey,
		EncryptedDataKeySignature: acct.AuthKey.Sign(encrDataKey),
		DataKeySignature:          acct.AuthKey.Sign(key.Bytes()),
		DataKeyAlgorithm:          key.Algorithm(),
	}, nil
}

func dataKeyFromSessionData(acct *Account, data *SessionData, senderEncrPubkey, senderAuthPubkey []byte) (DataKey, error) {
	key, err := acct.EncrKey.Decrypt(data.EncryptedDataKey, senderEncrPubkey)
	if err != nil {
		return nil, fmt.Errorf("session data not for the recipient: %v", err)
	}

	if !ed25519.Verify(senderAuthPubkey, data.EncryptedDataKey, data.EncryptedDataKeySignature) ||
		!ed25519.Verify(senderAuthPubkey, key, data.DataKeySignature) {
		return nil, fmt.Errorf("session data not from the sender: %v", err)
	}

	// switch data.DataKeyAlgorithm to determine which algorithm to generate data key
	// if more versions are supported in the future
	return &ChaCha20DataKey{key}, nil
}
