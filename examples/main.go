package main

import (
	"fmt"
	"time"

	bitmarksdk "github.com/bitmark-inc/bitmark-sdk-go"
)

func main() {
	// session is used to store configurations and credentials for using bitmark services
	sess := bitmarksdk.NewSession(bitmarksdk.Testnet, "apiId", "apiSecret")

	// issue 2 bitmarks under accountA
	accountA, _ := bitmarksdk.NewAccount(sess)
	bitmarkIds, _ := accountA.IssueBitmarks(
		sess,                                          // issuer account
		"Bitmark 101",                                 // asset name
		[]byte(fmt.Sprint(time.Now().Unix())),         // asset content (use current time to represent a new asset for demo)
		map[string]string{"desc": "bitmark tutorial"}, // asset metadata
		2,
	)
	for count, id := range bitmarkIds {
		fmt.Printf("[%d] bitmark id: %s\n", count+1, id)
	}

	// transfer the first bitmark to another account B
	accountB, _ := bitmarksdk.NewAccount(sess)
	accountA.TransferBitmark(sess, bitmarkIds[0], accountB.AccountNumber())

	// list bitmarks under account A
	size := 10
	total := 100
	descending := true
	pending := true
	selector := accountA.ListBitmarks(sess, size, total, descending, pending)
	for selector.Next() {
		if selector.Err != nil {
			fmt.Println(selector.Err.Error())
		}

		for _, bitmark := range selector.Items.([]bitmarksdk.Bitmark) {
			fmt.Println(bitmark.Id)
		}
	}
}
