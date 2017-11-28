package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	sdk "github.com/bitmark-inc/bitmark-sdk-go"
)

var (
	seed string

	chain string

	path string

	// issue
	acs      string
	name     string
	metadata string
	quantity int

	// transfer
	bitmarkId string
	receiver  string
)

func parseVars() {
	subcmd := flag.NewFlagSet("subcmd", flag.ExitOnError)
	subcmd.StringVar(&seed, "seed", "5XEECttxvRBzxzAmuV4oh6T1FcQu4mBg8eWd9wKbf8hweXsfwtJ8sfH", "")

	subcmd.StringVar(&chain, "chain", "test", "")

	subcmd.StringVar(&path, "path", "", "")
	subcmd.StringVar(&acs, "acs", "public", "")
	subcmd.StringVar(&name, "name", "Bitmark GO SDK trial", "")
	subcmd.StringVar(&metadata, "metadata", "", "")
	subcmd.IntVar(&quantity, "quantity", 1, "")

	subcmd.StringVar(&receiver, "receiver", "eZpG6Wi9SQvpDatEP7QGrx6nvzwd6s6R8DgMKgDbDY1R5bjzb9", "")
	subcmd.StringVar(&bitmarkId, "bmkid", "", "")

	subcmd.Parse(os.Args[2:])
}

func main() {
	session := sdk.NewSession(&http.Client{Timeout: 5 * time.Second})

	parseVars()
	acct, _ := session.RestoreAccountFromSeed(seed)

	switch os.Args[1] {
	case "create_account":
		network := sdk.Livenet
		if chain != "livenet" {
			network = sdk.Testnet
		}
		newacct, _ := session.CreateAccount(network)
		fmt.Println("account number", newacct.AccountNumber())
		fmt.Println("seed", newacct.Seed())
		fmt.Println("recovery phrase", strings.Join(newacct.RecoveryPhrase(), " "))
	case "issue":
		asset, _ := sdk.NewAssetFromFilePath("test", map[string]string{"author": "linzy"}, "/Users/linzyhu/Downloads/test.txt", "private")
		// asset := sdk.AssetFromId("ff1f12ef1d160dfdb1086fe8158d9c72857f3006c97adc086f91c80f04a321a738a3151b47b353f4199640ed9063bb177d6aeed440640c060ea3b0623e9d8de4")
		// asset, _ := sdk.AssetFromFilePath("/Users/linzyhu/Downloads/test.txt", "private")

		assetId, bitmarkIds, err := acct.Issue(asset, quantity)
		if err != nil {
			fmt.Println("issue failed: ", err)
			return
		}
		fmt.Println("Account Number: ", acct.AccountNumber())
		fmt.Println("Asset ID\n", assetId)
		fmt.Println("Bitmark ID")
		for i, id := range bitmarkIds {
			fmt.Printf("[%d] %s", i, id)
		}
	case "transfer":
		txId, err := acct.TransferBitmark(bitmarkId, receiver)
		if err != nil {
			fmt.Println("transfer failed: ", err)
			return
		}
		fmt.Println("Transaction ID: ", txId)
	case "download":
		fileName, content, err := acct.DownloadAsset(bitmarkId)
		if err != nil {
			fmt.Println("download failed: ", err)
			return
		}
		file, _ := os.Create(path + "/" + fileName)
		file.Write(content)
		file.Close()
	case "rent":
		err := acct.RentBitmark("b706b45f41ca4b3445603614d3286cdf18094c831c76fb679a2e63343bae1fc5", receiver, 1)
		if err != nil {
			fmt.Println("rent failed: ", err)
			return
		}
	case "list_leases":
		leases, err := acct.ListLeases()
		if err != nil {
			fmt.Println("lease failed: ", err)
			return
		}
		for _, lease := range leases {
			data, _ := acct.DownloadAssetByLease(lease)
			fmt.Printf("Content: %s", string(data))
		}
	}
}
