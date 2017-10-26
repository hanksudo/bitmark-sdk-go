package bitmarksdk

import (
	"encoding/hex"
	"io/ioutil"
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

func computeAssetId(fingerprint string) string {
	assetIndex := sha3.Sum512([]byte(fingerprint))
	return hex.EncodeToString(assetIndex[:])
}
