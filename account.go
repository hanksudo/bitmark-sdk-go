package bitmarksdk

import (
	"crypto/rand"
	"io"

	bitmarklib "github.com/bitmark-inc/go-bitmarklib"
	"golang.org/x/crypto/nacl/secretbox"
)

var (
	seedNonce = [24]byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	authSeedCountBM = [16]byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0xe7,
	}
	encrSeedCountBM = [16]byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0xe8,
	}
)

type Account struct {
	rootSeed    [32]byte
	IsTest      bool
	AuthKeyPair *bitmarklib.KeyPair
	EncrKeyPair *bitmarklib.EncrKeyPair
}

func (s *Session) NewAccount() (*Account, error) {
	var rootSeed [32]byte
	if _, err := io.ReadFull(rand.Reader, rootSeed[:]); err != nil {
		return nil, err
	}

	return accountFromRootSeed(rootSeed, s.chain == Testnet, 1)
}

func AccountFromMasterKey(masterKey [33]byte) (*Account, error) {
	rootSeed, isTest, version := parseMasterKey(masterKey)
	return accountFromRootSeed(rootSeed, isTest, version)
}

func AccountFromRecoveryPhrase(phrase []string) (*Account, error) {
	masterKey, err := phraseToBytes(phrase)
	if err != nil {
		return nil, err
	}

	rootSeed, isTest, version := parseMasterKey(masterKey)
	return accountFromRootSeed(rootSeed, isTest, version)
}

func accountFromRootSeed(rootSeed [32]byte, isTest bool, version int) (*Account, error) {
	authSeed := createAuthSeed(rootSeed)
	encrSeed := createEncrSeed(rootSeed)

	// ed25519 for auth key
	authKeypair, err := bitmarklib.NewKeyPairFromSeed(authSeed, isTest, bitmarklib.ED25519)
	if err != nil {
		return nil, err
	}

	// curve25519 for encr key
	encrKeypair, err := bitmarklib.NewEncrKeyPairFromSeed(encrSeed)
	if err != nil {
		return nil, err
	}

	return &Account{rootSeed, isTest, authKeypair, encrKeypair}, nil
}

func (a *Account) AccountNumber() string {
	return a.AuthKeyPair.Account().String()
}

func (a *Account) RecoveryPhrase() []string {
	key := constructMasterKey(a.rootSeed, a.IsTest, 1)
	return bytesToPhrase(key)
}

func parseMasterKey(masterKey [33]byte) (rootSeed [32]byte, test bool, version int) {
	keyPrefix := masterKey[0]
	test = keyPrefix&0x01 != 0x00
	version = int(keyPrefix >> 1)

	copy(rootSeed[:], masterKey[1:])
	return
}

func constructMasterKey(rootSeed [32]byte, test bool, version int) (masterKey [33]byte) {
	keyPrefix := (version << 1)
	if test {
		keyPrefix |= 0x01
	}

	copy(masterKey[:1], []byte{byte(keyPrefix)})
	copy(masterKey[1:], rootSeed[:])
	return
}

// set nonce = 999 for creating auth seed
func createAuthSeed(seed [32]byte) []byte {
	return secretbox.Seal([]byte{}, authSeedCountBM[:], &seedNonce, &seed)
}

// set nonce = 1000 for creating encr seed
func createEncrSeed(seed [32]byte) []byte {
	return secretbox.Seal([]byte{}, encrSeedCountBM[:], &seedNonce, &seed)
}
