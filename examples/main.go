package main

import (
	"fmt"
	"strings"

	sdk "github.com/bitmark-inc/bitmark-sdk-go"
)

func main() {
	// user, _ := sdk.NewUser(account.Testnet)
	// fmt.Println(user.Account.AccountNumber())
	// fmt.Println(user.Account.Seed())

	acct, _ := sdk.AccountFromSeed("5XEECt2hHWqFH9USKB9w6ByLVfpWWkHbhe9dFV8hHQA9uhX9B7Tt6Fg")
	// fmt.Println(acct.AccountNumber())

	bitmarkIds, err := acct.IssueNewBitmarks("/Users/linzyhu/Downloads/test.txt", sdk.Public, "test asset api", nil, 1)
	fmt.Println(err)
	fmt.Println("Bitmark ID: ", strings.Join(bitmarkIds, "\n"))

	// time.Sleep(5 * time.Minute)

	// acct.GetAsset("1bc0b5a9638ad2a70ad2001806229001a23295b8f2e5c1974e3716ac84a70bca")
}
