package bitmarksdk

import (
	"encoding/hex"

	bitmarklib "github.com/bitmark-inc/go-bitmarklib"
	"golang.org/x/crypto/sha3"
)

type txResponse []struct {
	TxId string `json:"txId"`
}

func (s *Session) IssueBitmarks(a *Account, assetName string, assetContent []byte, assetMetadata map[string]string, quantity int) ([]Bitmark, error) {
	digest := sha3.Sum512(assetContent)
	fingerprint := hex.EncodeToString(digest[:])

	// pack and sign asset record
	asset := bitmarklib.NewAsset(assetName, fingerprint)
	if err := asset.SetMeta(assetMetadata); err != nil {
		return nil, err
	}
	if err := asset.Sign(a.AuthKeyPair); err != nil {
		return nil, err
	}

	// pack and sign issue records
	issues := make([]bitmarklib.Issue, quantity)
	for i := 0; i < quantity; i++ {
		issue := bitmarklib.NewIssue(asset.AssetIndex())
		if err := issue.Sign(a.AuthKeyPair); err != nil {
			return nil, err
		}
		issues[i] = issue
	}

	// submit issue request
	body := map[string]interface{}{
		"assets": []bitmarklib.Asset{asset},
		"issues": issues,
	}
	var resp txResponse
	if err := submitAPIRequest(s, "POST", "/v1/issue", body, &resp); err != nil {
		return nil, err
	}

	bitmarks := make([]Bitmark, 0)
	for _, tx := range resp {
		bitmarks = append(bitmarks, Bitmark{Id: tx.TxId})
	}
	return bitmarks, nil
}

// TODO: switch to bitmarkId
func (s *Session) TransferBitmark(a *Account, bitmarkId, receiverAccountNo string) (string, error) {
	// pack and sign transfer record
	transfer, err := bitmarklib.NewTransfer(bitmarkId, receiverAccountNo)
	if err != nil {
		return "", err
	}
	if err := transfer.Sign(a.AuthKeyPair); err != nil {
		return "", err
	}

	// submit transfer request
	body := map[string]interface{}{
		"transfer": transfer,
	}
	var reply []struct {
		TxId string `json:"txId"`
	}
	if err := submitAPIRequest(s, "POST", "/v1/transfer", body, &reply); err != nil {
		return "", err
	}

	return reply[0].TxId, nil
}
