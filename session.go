package bitmarksdk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type chain int

const (
	Livenet chain = iota
	Testnet chain = iota
)

var (
	client = new(http.Client)

	domain = map[chain]string{
		Livenet: "https://api.bitmark.com",
		Testnet: "https://api.test.bitmark.com",
	}
)

type Session struct {
	chain chain

	apiId     string
	apiSecret string
}

func NewSession(chain chain, apiId, apiSecret string) *Session {
	return &Session{chain, apiId, apiSecret}
}

func submitAPIRequest(s *Session, method, path string, body interface{}, reply interface{}) error {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(body)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s%s", domain[s.chain], path)
	req, err := http.NewRequest(method, url, b)
	if nil != err {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	if resp.StatusCode/100 != 2 {
		return errors.New(string(data))
	}

	if reply != nil {
		err = json.Unmarshal(data, reply)
		if err != nil {
			return err
		}
	}

	return nil
}
