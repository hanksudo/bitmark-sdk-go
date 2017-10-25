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
		sessData, e := createSessionData(acct, dataKey)
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
		Sender   string       `json:"string"`
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

	encrPubkey, _ := c.getEncPubkey(result.Sender)
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
