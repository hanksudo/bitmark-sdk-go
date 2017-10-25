package main

import (
	"fmt"
	"strings"

	sdk "github.com/bitmark-inc/bitmark-sdk-go"
)

func main() {
	acct, _ := sdk.AccountFromSeed("5XEECt2hHWqFH9USKB9w6ByLVfpWWkHbhe9dFV8hHQA9uhX9B7Tt6Fg")
	fmt.Println("Account Number: ", acct.AccountNumber())

	bitmarkIds, err := acct.IssueNewBitmarks("/Users/linzyhu/Downloads/test.txt", sdk.Private, "test asset api", nil, 1)
	if err != nil {
		fmt.Println("issue failed: ", err)
		return
	}
	fmt.Println("Bitmark ID: ", strings.Join(bitmarkIds, "\n"))
}
