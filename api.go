package bitmarksdk

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type APIClient struct {
	client       *http.Client
	apiServer    string
	assetsServer string
}

type APIRequest struct {
	*http.Request
}

type assetAccess struct {
	URL      string       `json:"url"`
	Sender   string       `json:"sender"`
	SessData *SessionData `json:"session_data"`
}

func (r APIRequest) Sign(acct *Account, action, resource string) error {
	ts := strconv.FormatInt(time.Now().UnixNano()/1000000, 10)
	parts := []string{
		action,
		resource,
		acct.AccountNumber(),
		ts,
	}
	message := strings.Join(parts, "|")
	sig := hex.EncodeToString(acct.AuthKey.Sign([]byte(message)))

	r.Header.Add("requester", acct.AccountNumber())
	r.Header.Add("timestamp", ts)
	r.Header.Add("signature", sig)
	return nil
}

func NewAPIRequest(method, url string, body io.Reader) (*APIRequest, error) {
	r, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	return &APIRequest{r}, nil
}

func (api *APIClient) submitRequest(req *APIRequest, reply interface{}) ([]byte, error) {
	resp, err := api.client.Do(req.Request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode/100 != 2 {
		var m struct {
			Message string `json:"message"`
		}
		err := json.Unmarshal(data, &m)
		if err != nil {
			return nil, errors.New(string(data))
		}
		return nil, errors.New(m.Message)
	}

	if reply != nil {
		err = json.Unmarshal(data, reply)
		if err != nil {
			return nil, fmt.Errorf("json decode error: %s, data: %s", err.Error(), string(data))
		}
	}

	return data, nil
}

func NewAPIClient(network Network) *APIClient {
	c := &http.Client{
		Timeout: 3 * time.Second,
	}

	api := &APIClient{
		client: c,
	}

	switch network {
	case Testnet:
		api.apiServer = "api.test.bitmark.com"
		api.assetsServer = "assets.test.bitmark.com"
	case Livenet:
		api.apiServer = "api.bitmark.com"
		api.assetsServer = "assets.bitmark.com"
	}

	return api
}

func (c *APIClient) UploadAsset(acct *Account, af *assetFile, acs Accessibility) error {
	assetId := computeAssetId(af.Fingerprint)

	body := new(bytes.Buffer)

	bodyWriter := multipart.NewWriter(body)
	bodyWriter.WriteField("asset_id", assetId)
	bodyWriter.WriteField("accessibility", string(acs))

	fileWriter, err := bodyWriter.CreateFormFile("file", af.Name)
	if err != nil {
		return err
	}

	switch acs {
	case Public:
		if _, e := fileWriter.Write(af.Content); e != nil {
			return err
		}
	case Private:
		dataKey, e := NewDataKey()
		if e != nil {
			return err
		}
		encryptedContent, e := dataKey.Encrypt(af.Content)
		if e != nil {
			return err
		}
		sessData, e := createSessionData(acct, dataKey, acct.EncrKey.PublicKeyBytes())
		if e != nil {
			return err
		}
		if _, e := fileWriter.Write(encryptedContent); e != nil {
			return err
		}
		bodyWriter.WriteField("session_data", sessData.String())
	}

	err = bodyWriter.Close()
	if err != nil {
		return err
	}

	u := url.URL{
		Scheme: "https",
		Host:   c.apiServer,
		Path:   "/v1/assets",
	}
	req, _ := NewAPIRequest("POST", u.String(), body)

	req.Header.Set("Content-Type", bodyWriter.FormDataContentType())
	req.Sign(acct, "uploadAsset", assetId)

	_, err = c.submitRequest(req, nil)
	return err
}

func (c *APIClient) getAssetAccess(acct *Account, bitmarkId string) (*assetAccess, error) {
	u := url.URL{
		Scheme: "https",
		Host:   c.apiServer,
		Path:   fmt.Sprintf("/v1/bitmarks/%s/asset", bitmarkId),
	}

	req, _ := NewAPIRequest("GET", u.String(), nil)
	req.Sign(acct, "downloadAsset", bitmarkId)

	var result assetAccess
	if _, err := c.submitRequest(req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *APIClient) DownloadAsset(acct *Account, bitmarkId string) ([]byte, error) {
	u := url.URL{
		Scheme: "https",
		Host:   c.apiServer,
		Path:   fmt.Sprintf("/v1/bitmarks/%s/asset", bitmarkId),
	}

	req, _ := NewAPIRequest("GET", u.String(), nil)
	req.Sign(acct, "downloadAsset", bitmarkId)

	var result struct {
		URL      string       `json:"url"`
		Sender   string       `json:"sender"`
		SessData *SessionData `json:"session_data"`
	}
	if _, err := c.submitRequest(req, &result); err != nil {
		return nil, err
	}

	req, _ = NewAPIRequest("GET", result.URL, nil)
	content, err := c.submitRequest(req, nil)
	if err != nil {
		return nil, err
	}

	if result.SessData == nil {
		return content, nil
	}

	encrPubkey, err := c.getEncPubkey(result.Sender)
	if err != nil {
		return nil, fmt.Errorf("fail to get enc public key: %s", err.Error())
	}
	dataKey, err := dataKeyFromSessionData(acct, result.SessData, encrPubkey)
	if err != nil {
		return nil, err
	}

	plaintext, err := dataKey.Decrypt(content)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func (c *APIClient) getEncPubkey(acctNo string) ([]byte, error) {
	u := url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("key.%s", c.assetsServer),
		Path:   fmt.Sprintf("/%s", acctNo),
	}

	req, _ := NewAPIRequest("GET", u.String(), nil)

	var result struct {
		Key string `json:"encryption_pubkey"`
	}
	if _, err := c.submitRequest(req, &result); err != nil {
		return nil, err
	}

	return hex.DecodeString(result.Key)
}

func (c *APIClient) setEncPubkey(acct *Account) error {
	u := url.URL{
		Scheme: "https",
		Host:   c.apiServer,
		Path:   fmt.Sprintf("/v1/encryption_keys/%s", acct.AccountNumber()),
	}

	signature := hex.EncodeToString(acct.AuthKey.Sign(acct.EncrKey.PublicKeyBytes()))

	reqBody := map[string]string{
		"encryption_pubkey": fmt.Sprintf("%064x", acct.EncrKey.PublicKeyBytes()),
		"signature":         signature,
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(reqBody)
	if err != nil {
		return err
	}

	req, _ := NewAPIRequest("POST", u.String(), &buf)
	_, err = c.submitRequest(req, nil)
	return err
}

func (c *APIClient) updateSession(acct *Account, bitmarkId, receiver string, data *SessionData) error {
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(map[string]interface{}{
		"bitmark_id":   bitmarkId,
		"owner":        receiver,
		"session_data": data,
	})

	req, _ := NewAPIRequest("POST", "http://0.0.0.0:8087/v2/session", body)
	req.Sign(acct, "updateSession", data.String())
	_, err := c.submitRequest(req, nil)
	return err
}

func (c *APIClient) issue(asset *AssetRecord, issues []*IssueRecord) ([]string, error) {
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(map[string]interface{}{
		"assets": []*AssetRecord{asset},
		"issues": issues,
	})
	req, _ := NewAPIRequest("POST", "https://api.devel.bitmark.com/v1/issue", body)

	bitmarks := make([]struct {
		TxId string `json:"txId"`
	}, 0)
	if _, err := c.submitRequest(req, &bitmarks); err != nil {
		return nil, err
	}

	bitmarkIds := make([]string, 0)
	for _, b := range bitmarks {
		bitmarkIds = append(bitmarkIds, b.TxId)
	}

	return bitmarkIds, nil
}

func (c *APIClient) transfer(t *TransferRecord) (string, error) {
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(map[string]interface{}{
		"transfer": t,
	})
	req, _ := NewAPIRequest("POST", "https://api.devel.bitmark.com/v1/issue", body)

	txs := make([]struct {
		TxId string `json:"txId"`
	}, 0)
	if _, err := c.submitRequest(req, &txs); err != nil {
		return "", err
	}

	return txs[0].TxId, nil
}

func (c *APIClient) getBitmark(bitmarkId string) (*Bitmark, error) {
	req, _ := NewAPIRequest("GET", "https://api.devel.bitmark.com/v1/bitmarks/"+bitmarkId, nil)

	var bmk Bitmark
	_, err := c.submitRequest(req, &bmk)
	return &bmk, err
}
