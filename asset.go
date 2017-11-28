package bitmarksdk

import (
	"encoding/hex"
	"io/ioutil"
	"path/filepath"

	"golang.org/x/crypto/sha3"
)

type Accessibility string

const (
	Public  Accessibility = "public"
	Private Accessibility = "private"
)

type Asset struct {
	Id       string
	Name     string
	Metadata map[string]string
	File     *AssetFile
}

type AssetFile struct {
	Path          string
	Name          string
	Content       []byte
	Fingerprint   string
	Accessibility Accessibility
}

func NewAssetFromFilePath(name string, metadata map[string]string, path string, acs Accessibility) (*Asset, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	file := &AssetFile{
		path,
		filepath.Base(path),
		content,
		computeFingerprint(content),
		acs,
	}

	return &Asset{
		Id:       computeAssetId(file.Fingerprint),
		Name:     name,
		Metadata: metadata,
		File:     file,
	}, nil
}

func AssetFromFilePath(path string, acs Accessibility) (*Asset, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	file := &AssetFile{
		path,
		filepath.Base(path),
		content,
		computeFingerprint(content),
		acs,
	}

	return &Asset{
		Id:   computeAssetId(file.Fingerprint),
		File: file,
	}, nil
}

func AssetFromId(id string) *Asset {
	return &Asset{Id: id}
}

func computeFingerprint(content []byte) string {
	digest := sha3.Sum512(content)
	return "01" + hex.EncodeToString(digest[:])
}

func computeAssetId(fingerprint string) string {
	assetIndex := sha3.Sum512([]byte(fingerprint))
	return hex.EncodeToString(assetIndex[:])
}
