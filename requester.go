package bitmarksdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

var (
	client    *http.Client
	apidomain = map[Network]string{
		Livenet: "https://api.bitmark.com",
		Testnet: "https://api.test.bitmark.com",
	}
)

func init() {
	client = &http.Client{}
}

func submitRequest(req *http.Request, reply interface{}) ([]byte, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	if resp.StatusCode/100 != 2 {
		var m struct {
			Message string `json:"message"`
		}
		json.Unmarshal(data, &m)
		fmt.Println("error: ", m.Message)
		return nil, errors.New(m.Message)
	}

	if reply != nil {
		err = json.Unmarshal(data, reply)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}
