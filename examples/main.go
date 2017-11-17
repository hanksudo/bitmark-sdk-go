package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	sdk "github.com/bitmark-inc/bitmark-sdk-go"
)

func main() {
	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}
	session := sdk.NewSession(httpClient)

	acctA, _ := session.RestoreAccountFromSeed("5XEECttxvRBzxzAmuV4oh6T1FcQu4mBg8eWd9wKbf8hweXsfwtJ8sfH")
	fmt.Println("ACCOUNT NO OF USER A: ", acctA.AccountNumber()) // e1pFRPqPhY2gpgJTpCiwXDnVeouY9EjHY6STtKwdN6Z4bp4sog

	acctB, _ := session.RestoreAccountFromSeed("5XEECt4yuMK4xqBLr9ky5FBWpkAR6VHNZSz8fUzZDXPnN3D9MeivTSA")
	fmt.Println("ACCOUNT NO OF USER B: ", acctB.AccountNumber()) // eZpG6Wi9SQvpDatEP7QGrx6nvzwd6s6R8DgMKgDbDY1R5bjzb9

	// Issue a new bitmark under account A
	_, bitmarkIds, err := acctA.IssueBitmarks("/Users/linzyhu/Downloads/test.txt", sdk.Public, "test asset api", map[string]string{"test_no": "24"}, 1)
	if err != nil {
		fmt.Println("issue failed: ", err)
		return
	}
	fmt.Println("Bitmark ID: ", strings.Join(bitmarkIds, "\n"))

	// Read the asset content of the bitmark for account A
	// content, err := acctA.DownloadAsset("ad3f124f236c8241281dff4de9089ae359702f464058862f630a68b7d969c0b6")
	// if err != nil {
	// 	fmt.Println("download failed: ", err)
	// 	return
	// }
	// fmt.Println("Asset Content: ", string(content))

	// Transfer the bitmark to account B
	// txId, err := acctA.TransferBitmark("ad3f124f236c8241281dff4de9089ae359702f464058862f630a68b7d969c0b6", acctB.AccountNumber())
	// if err != nil {
	// 	fmt.Println("transfer failed: ", err)
	// 	return
	// }
	// fmt.Println("Transaction ID: ", txId)

	// Read the asset content of the bitmark for account B
	// content, err := acctA.DownloadAsset("cbdeba02cf5f3bc1adf1ba59ebb705e543697889378123ee7184c36c2d0b6bd1")
	// if err != nil {
	// 	fmt.Println("download failed: ", err)
	// 	return
	// }
	// fmt.Println("Asset Content: ", string(content))
}
