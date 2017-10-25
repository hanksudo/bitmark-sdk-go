package bitmarksdk

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"

	"golang.org/x/crypto/sha3"
)

type Accessibility string

const (
	Public  Accessibility = "public"
	Private Accessibility = "private"
)

type assetFile struct {
	Name        string
	Content     []byte
	Fingerprint string
}

func readAssetFile(u string) (*assetFile, error) {
	result, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadFile(result.Path)
	if err != nil {
		return nil, err
	}
	return &assetFile{
		filepath.Base(result.Path),
		content,
		computeFingerprint(content),
	}, err

	// switch result.Scheme {
	// case "file":
	// 	content, err := ioutil.ReadFile(result.Path)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	return &assetFile{
	// 		filepath.Base(result.Path),
	// 		content,
	// 		computeFingerprint(content),
	// 	}, err
	// default:
	// 	return nil, errors.New("scheme not supported")
	// }
}

func computeFingerprint(content []byte) string {
	digest := sha3.Sum512(content)
	return "01" + hex.EncodeToString(digest[:])
}

func uploadAsset(acct *Account, af *assetFile, acs Accessibility) error {
	assetId := computeAssetId(af.Fingerprint)

	body := new(bytes.Buffer)

	bodyWriter := multipart.NewWriter(body)
	bodyWriter.WriteField("assetId", assetId)
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

	// TODO
	// req, _ := http.NewRequest("POST", "http://0.0.0.0:8087/v1/assets", body)
	req, _ := http.NewRequest("POST", "https://api.devel.bitmark.com/v1/assets", body)

	req.Header.Set("Content-Type", bodyWriter.FormDataContentType())

	acct.signRequest(req, "uploadAsset", assetId)

	_, err = submitRequest(req, nil)
	return err
}

func computeAssetId(fingerprint string) string {
	assetIndex := sha3.Sum512([]byte(fingerprint))
	return hex.EncodeToString(assetIndex[:])
}

func downloadAsset(acct *Account, bitmarkId string) ([]byte, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.devel.bitmarks.com/v1/bitmarks/%s/asset", bitmarkId), nil)

	var result struct {
		URL      string       `json:"url"`
		Sender   string       `json:"string"`
		SessData *SessionData `json:"session_data"`
	}
	if _, err := submitRequest(req, &result); err != nil {
		return nil, err
	}

	req, _ = http.NewRequest("GET", result.URL, nil)
	content, err := submitRequest(req, nil)
	if err != nil {
		return nil, err
	}

	if result.SessData == nil {
		return content, nil
	}

	encrPubkey, _ := getEncrPubkey(result.Sender)
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

func getEncrPubkey(acctNo string) ([]byte, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://key.assets.devel.bitmark.com/%s", acctNo), nil)

	var result struct {
		Key string `json:"encryption_pubkey"`
	}
	if _, err := submitRequest(req, &result); err != nil {
		return nil, err
	}

	return hex.DecodeString(result.Key)
}
