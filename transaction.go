package bitmarksdk

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func issue(asset *AssetRecord, issues []*IssueRecord) ([]string, error) {
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(map[string]interface{}{
		"assets": []*AssetRecord{asset},
		"issues": issues,
	})
	req, _ := http.NewRequest("POST", "https://api.devel.bitmark.com/v1/issue", body)

	bitmarks := make([]struct {
		TxId string `json:"txId"`
	}, 0)
	if _, err := submitRequest(req, &bitmarks); err != nil {
		return nil, err
	}

	bitmarkIds := make([]string, 0)
	for _, b := range bitmarks {
		bitmarkIds = append(bitmarkIds, b.TxId)
	}

	return bitmarkIds, nil
}
