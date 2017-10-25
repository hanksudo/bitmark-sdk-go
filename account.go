package bitmarksdk

import (
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/sha3"
)

const (
	pubkeyMask     = 0x01
	testnetMask    = 0x02
	algorithmShift = 4
	checksumLength = 4
)

type Account struct {
	seed    *Seed
	network Network
	api     *APIClient
	AuthKey AuthKey
	EncrKey EncrKey
}

func NewAccount(network Network) (*Account, error) {
	seed, err := NewSeed(SeedVersion1, network)
	if err != nil {
		return nil, err
	}

	authKey, err := NewAuthKey(seed)
	if err != nil {
		return nil, err
	}

	encrKey, err := NewEncrKey(seed)
	if err != nil {
		return nil, err
	}

	apiClient := NewAPIClient(network)
	account := &Account{seed, network, apiClient, authKey, encrKey}
	err = account.api.setEncPubkey(account)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func AccountFromSeed(s string) (*Account, error) {
	seed, err := SeedFromBase58(s)
	if err != nil {
		return nil, err
	}

	authKey, err := NewAuthKey(seed)
	if err != nil {
		return nil, err
	}

	encrKey, err := NewEncrKey(seed)
	if err != nil {
		return nil, err
	}

	apiClient := NewAPIClient(seed.network)

	return &Account{seed, seed.network, apiClient, authKey, encrKey}, nil
}

// TODO
func AccountFromRecoveryPhrase(s string) (*Account, error) {
	return nil, nil
}

func (acct *Account) Network() Network {
	return acct.seed.network
}

func (acct *Account) Seed() string {
	return acct.seed.String()
}

// TODO
func (acct *Account) RecoveryPhrase() []string {
	return nil
}

func (acct *Account) AccountNumber() string {
	buffer := acct.bytes()
	checksum := sha3.Sum256(buffer)
	buffer = append(buffer, checksum[:checksumLength]...)
	return toBase58(buffer)
}

func (acct *Account) bytes() []byte {
	keyVariant := byte(acct.AuthKey.Algorithm()<<algorithmShift) | pubkeyMask
	if acct.seed.network == Testnet {
		keyVariant |= testnetMask
	}
	return append([]byte{keyVariant}, acct.AuthKey.PublicKeyBytes()...)
}

func (acct *Account) signRequest(req *http.Request, action, resource string) {
	ts := strconv.FormatInt(time.Now().UnixNano()/1000000, 10)
	parts := []string{
		action,
		resource,
		acct.AccountNumber(),
		ts,
	}
	message := strings.Join(parts, "|")
	sig := hex.EncodeToString(acct.AuthKey.Sign([]byte(message)))

	req.Header.Add("requester", acct.AccountNumber())
	req.Header.Add("timestamp", ts)
	req.Header.Add("signature", sig)
}

func AuthPublicKeyFromAccountNumber(acctNo string) []byte {
	buffer := fromBase58(acctNo)
	return buffer[:len(buffer)-checksumLength]
}

func (a *Account) DownloadAsset(bitmarkId string) ([]byte, error) {
	return a.api.DownloadAsset(a, bitmarkId)
}
