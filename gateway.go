package bitmarksdk

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func (r *Requester) IssueBitmarks(asset *AssetRecord, isssues []*IssueRecord) ([]string, error) {
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(
		map[string]interface{}{
			"assets": []*AssetRecord{asset},
			"issues": isssues,
		})

	req, _ := http.NewRequest("POST", r.domain+"/v1/issue", body)

	_, err := submitRequest(req, nil)
	if err != nil {
		return nil, err
	}

	// TODO
	return nil, nil
}
