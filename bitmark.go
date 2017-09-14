package bitmarksdk

import (
	"fmt"
	"time"
)

type Bitmark struct {
	Id          string    `json:"id"`
	AssetId     string    `json:"asset_id"`
	Issuer      string    `json:"issuer"`
	IssuerAt    time.Time `json:"issued_at"`
	CreatedAt   time.Time `json:"created_at"`
	Owner       string    `json:"owner"`
	Head        string    `json:"head"`
	HeadId      string    `json:"head_id"`
	Status      string    `json:"status"`
	Offset      int       `json:"offset"`
	BlockNumber int       `json:"block_number"`
	Provenance  []Holder  `json:"provenance"`
}

type Holder struct {
	TxId      string    `json:"tx_id"`
	Owner     string    `json:"owner"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func (b *Bitmark) Update(sess *Session) error {
	var r struct {
		Bitmark Bitmark `json:"bitmark"`
	}

	path := fmt.Sprintf("/v1/bitmarks/%s?pending=true", b.Id)
	if err := submitAPIRequest(sess, "GET", path, nil, &r); err != nil {
		return err
	}

	*b = r.Bitmark
	return nil
}
