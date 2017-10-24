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
	apidomain = map[string]string{
		"live": "https://api.bitmark.com",
		"test": "https://api.test.bitmark.com",
	}
)

func init() {
	client = &http.Client{}
}

type Requester struct {
	domain    string
	requester string
	authKey   AuthKey
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
	fmt.Println(string(data))
	// TODO
	if resp.StatusCode/100 != 2 {
		var m struct {
			Message string
		}
		json.Unmarshal(data, &m)
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
