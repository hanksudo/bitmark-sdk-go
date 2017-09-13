package bitmarksdk

import (
	"encoding/hex"
	"net/url"
	"strconv"

	bitmarklib "github.com/bitmark-inc/go-bitmarklib"
	"golang.org/x/crypto/sha3"
)

type txResponse []struct {
	TxId string `json:"txId"`
}

func (a *Account) IssueBitmarks(sess *Session, assetName string, assetContent []byte, assetMetadata map[string]string, quantity int) ([]string, error) {
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
	if err := submitAPIRequest(sess, "POST", "/v1/issue", body, &resp); err != nil {
		return nil, err
	}

	bitmarkIds := make([]string, 0)
	for _, tx := range resp {
		bitmarkIds = append(bitmarkIds, tx.TxId)
	}
	return bitmarkIds, nil
}

// TODO: switch to bitmarkId
func (a *Account) TransferBitmark(sess *Session, bitmarkId, receiverAccountNo string) (string, error) {
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
	if err := submitAPIRequest(sess, "POST", "/v1/transfer", body, &reply); err != nil {
		return "", err
	}

	return reply[0].TxId, nil
}

func (a *Account) ListBitmarks(sess *Session, size, total int, descending, pending bool) selector {
	to := "later"
	if descending {
		to = "earlier"
	}

	params := url.Values{}
	params.Set("owner", a.AccountNumber())
	params.Add("to", to)
	params.Add("limit", strconv.Itoa(size))
	params.Add("pending", strconv.FormatBool(pending))

	return selector{sess: sess, params: params, total: total}
}

type selector struct {
	sess   *Session
	params url.Values

	count int
	total int
	Items interface{}

	Err error
}

func (s *selector) Next() bool {
	var r struct {
		Bitmarks []Bitmark `json:"bitmarks"`
	}

	if err := submitAPIRequest(s.sess, "GET", "/v1/bitmarks?"+s.params.Encode(), nil, &r); err != nil {
		s.Err = err
		return false
	}

	if len(r.Bitmarks) == 0 || s.count == s.total {
		return false
	}

	index := len(r.Bitmarks)
	if len(r.Bitmarks) > s.total-s.count {
		index = s.total - s.count
	}

	s.Items = r.Bitmarks[:index]
	s.count += index
	s.params.Set("at", strconv.Itoa(r.Bitmarks[index-1].Offset))
	return true
}
