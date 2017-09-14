package bitmarksdk

import (
	"net/url"
	"strconv"

	"github.com/google/go-querystring/query"
)

type Filter struct {
	AssetId string `url:"asset_id"`
	Issuer  string `url:"issuer"`
	Owner   string `url:"owner"`
	Sent    bool   `url:"sent"`
	Pending bool   `url:"pending"`
}

func (s *Session) ListBitmarks(f Filter, size, total int, descending bool) selector {
	params, _ := query.Values(f)

	to := "later"
	if descending {
		to = "earlier"
	}
	params.Add("to", to)
	params.Add("limit", strconv.Itoa(size))

	return selector{sess: s, params: params, total: total}
}

type selector struct {
	sess   *Session
	params url.Values

	count int
	total int
	Items interface{}

	Err error
}

func (s *selector) Next() bool {
	var r struct {
		Bitmarks []Bitmark `json:"bitmarks"`
	}

	if err := submitAPIRequest(s.sess, "GET", "/v1/bitmarks?"+s.params.Encode(), nil, &r); err != nil {
		s.Err = err
		return false
	}

	if len(r.Bitmarks) == 0 || s.count == s.total {
		return false
	}

	index := len(r.Bitmarks)
	if len(r.Bitmarks) > s.total-s.count {
		index = s.total - s.count
	}

	s.Items = r.Bitmarks[:index]
	s.count += index
	s.params.Set("at", strconv.Itoa(r.Bitmarks[index-1].Offset))
	return true
}
