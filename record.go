package bitmarksdk

import (
	"encoding/hex"
	"errors"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/sha3"
)

const (
	assetRecordTag    = uint64(2)
	issueRecordTag    = uint64(3)
	transferRecordTag = uint64(4)

	minNameLength        = 1
	maxNameLength        = 64
	maxMetadataLength    = 2048
	minFingerprintLength = 1
	maxFingerprintLength = 1024
	assetIndexLength     = 64
	merkleDigestLength   = 32
)

// TODO: refine errors
var (
	ErrInvalidLength  = errors.New("text")
	ErrInvalidAccount = errors.New("text")
)

var nonceIndex uint64

type AssetRecord struct {
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
	Metadata    string `json:"metadata"`
	Registrant  string `json:"registrant"`
	Signature   string `json:"signature"`
}

func NewAssetRecord(name, fingerprint string, metadata map[string]string, registrant *Account) (*AssetRecord, error) {
	parts := make([]string, 0, len(metadata)*2)
	for key, val := range metadata {
		if key == "" || val == "" {
			continue
		}
		parts = append(parts, key, val)
	}
	compactMetadata := strings.Join(parts, "\u0000")

	if utf8.RuneCountInString(name) < minNameLength || utf8.RuneCountInString(name) > maxNameLength {
		return nil, ErrInvalidLength
	}

	if utf8.RuneCountInString(fingerprint) < minFingerprintLength || utf8.RuneCountInString(fingerprint) > maxFingerprintLength {
		return nil, ErrInvalidLength
	}

	if utf8.RuneCountInString(compactMetadata) > maxMetadataLength {
		return nil, ErrInvalidLength
	}

	if registrant == nil {
		return nil, ErrInvalidAccount
	}

	// pack and sign
	message := toVarint64(assetRecordTag)
	message = appendString(message, name)
	message = appendString(message, fingerprint)
	message = appendString(message, compactMetadata)
	message = appendBytes(message, registrant.bytes())
	signature := hex.EncodeToString(registrant.AuthKey.Sign(message))

	return &AssetRecord{name, fingerprint, compactMetadata, registrant.AccountNumber(), signature}, nil
}

func (asset *AssetRecord) Id() string {
	digest := sha3.Sum512([]byte(asset.Fingerprint))
	return hex.EncodeToString(digest[:])
}

type IssueRecord struct {
	AssetIndex string `json:"asset"`
	Owner      string `json:"owner"`
	Nonce      uint64 `json:"nonce"`
	Signature  string `json:"signature"`
}

func NewIssueRecord(assetIndex string, issuer *Account) (*IssueRecord, error) {
	assetIndexBytes, err := hex.DecodeString(assetIndex)
	if err != nil || len(assetIndexBytes) != assetIndexLength {
		return nil, ErrInvalidLength
	}

	if issuer == nil {
		return nil, ErrInvalidAccount
	}

	atomic.AddUint64(&nonceIndex, 1)
	nonce := uint64(time.Now().UTC().Unix())*1000 + nonceIndex%1000

	// pack and sign
	message := toVarint64(issueRecordTag)
	message = appendBytes(message, assetIndexBytes)
	message = appendBytes(message, issuer.bytes())
	message = appendUint64(message, nonce)
	signature := hex.EncodeToString(issuer.AuthKey.Sign(message))

	return &IssueRecord{
		assetIndex,
		issuer.AccountNumber(),
		nonce,
		signature,
	}, nil
}

func NewIssueRecords(assetIndex string, issuer *Account, quantity int) ([]*IssueRecord, error) {
	issues := make([]*IssueRecord, quantity)
	for i := 0; i < quantity; i++ {
		var issue *IssueRecord
		issue, err := NewIssueRecord(assetIndex, issuer)
		if err != nil {
			return nil, err
		}
		issues[i] = issue
	}
	return issues, nil
}

type TransferRecord struct {
	Link      string `json:"link"`
	Owner     string `json:"owner"`
	Signature string `json:"signature"`
}

func NewTransferRecord(txId string, receiver string, owner *Account) (*TransferRecord, error) {
	link, err := hex.DecodeString(txId)
	if err != nil || len(link) != merkleDigestLength {
		return nil, ErrInvalidLength
	}

	if owner == nil {
		return nil, ErrInvalidAccount
	}

	// pack and sign
	message := toVarint64(transferRecordTag)
	message = appendBytes(message, link)
	message = append(message, 0) // payment not supported
	message = appendBytes(message, AuthPublicKeyFromAccountNumber(receiver))
	signature := hex.EncodeToString(owner.AuthKey.Sign(message))

	return &TransferRecord{txId, receiver, signature}, nil
}

const varint64MaximumBytes = 9

func appendString(buffer []byte, s string) []byte {
	l := toVarint64(uint64(len(s)))
	buffer = append(buffer, l...)
	return append(buffer, s...)
}

func appendBytes(buffer []byte, data []byte) []byte {
	l := toVarint64(uint64(len(data)))
	buffer = append(buffer, l...)
	buffer = append(buffer, data...)
	return buffer
}

func appendUint64(buffer []byte, value uint64) []byte {
	valueBytes := toVarint64(value)
	buffer = append(buffer, valueBytes...)
	return buffer
}

func toVarint64(value uint64) []byte {
	result := make([]byte, 0, varint64MaximumBytes)
	if value < 0x80 {
		result = append(result, byte(value))
		return result
	}
	for i := 0; i < varint64MaximumBytes && value != 0; i++ {
		ext := uint64(0x80)
		if value < 0x80 {
			ext = 0x00
		}
		result = append(result, byte(value|ext))
		value >>= 7
	}
	return result
}