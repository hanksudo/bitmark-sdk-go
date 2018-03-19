package bitmarksdk

import (
	"fmt"
	"testing"
)

func TestBitmarFilter(t *testing.T) {
	filter := BitmarkFilter{}
	expectedEncodedURLValues := fmt.Sprintf("asset=%s&owner_sent=%s&pending=%s",
		"false",
		"false",
		"false",
	)
	if toURLValues(&filter).Encode() != expectedEncodedURLValues {
		t.Fail()
	}

	filter = BitmarkFilter{Limit: 101}
	expectedEncodedURLValues = fmt.Sprintf("asset=%s&limit=%s&owner_sent=%s&pending=%s",
		"false",
		"100",
		"false",
		"false",
	)
	if toURLValues(&filter).Encode() != expectedEncodedURLValues {
		t.Fail()
	}

	filter = BitmarkFilter{
		AssetId:   "ce9cb9f5fba848b4b0b9a505cf864384c7268a599b172de5049d1b65ff018f178089441a46a7e82a1a7997af110da8bf81758ae295d50a0c7571d6f7007bffa3",
		Issuer:    "eNc14gzHM1SqVBVpuLo6Qabj24hcopice29cZzbrFYmUrM8KyE",
		Owner:     "eNc14gzHM1SqVBVpuLo6Qabj24hcopice29cZzbrFYmUrM8KyE",
		OwnerSent: true,
		Asset:     true,
		Pending:   true,
		To:        "later",
		At:        171798,
		Limit:     100,
	}
	expectedEncodedURLValues = fmt.Sprintf("asset=%s&asset_id=%s&at=%s&issuer=%s&limit=%s&owner=%s&owner_sent=%s&pending=%s&to=%s",
		"true",
		"ce9cb9f5fba848b4b0b9a505cf864384c7268a599b172de5049d1b65ff018f178089441a46a7e82a1a7997af110da8bf81758ae295d50a0c7571d6f7007bffa3",
		"171798",
		"eNc14gzHM1SqVBVpuLo6Qabj24hcopice29cZzbrFYmUrM8KyE",
		"100",
		"eNc14gzHM1SqVBVpuLo6Qabj24hcopice29cZzbrFYmUrM8KyE",
		"true",
		"true",
		"later",
	)
	if toURLValues(&filter).Encode() != expectedEncodedURLValues {
		t.Fail()
	}
}
