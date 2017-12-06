package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/nacl/box"

	sdk "github.com/bitmark-inc/bitmark-sdk-go"
)

var (
	seed string

	chain string

	path string

	// issue
	filepath string
	acs      string

	assetId string

	name        string
	rawMetadata string

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

	subcmd.StringVar(&filepath, "p", "", "")
	subcmd.StringVar(&acs, "acs", "public", "")
	subcmd.StringVar(&name, "name", "", "")
	subcmd.StringVar(&rawMetadata, "meta", "", "")
	subcmd.StringVar(&assetId, "aid", "", "")
	subcmd.IntVar(&quantity, "quantity", 1, "")

	subcmd.StringVar(&receiver, "receiver", "eZpG6Wi9SQvpDatEP7QGrx6nvzwd6s6R8DgMKgDbDY1R5bjzb9", "")
	subcmd.StringVar(&bitmarkId, "bid", "", "")

	subcmd.Parse(os.Args[2:])
}

func toMedatadata() map[string]string {
	parts := strings.Split(rawMetadata, ",")
	metadata := make(map[string]string)
	if len(parts) > 1 {
		for _, part := range parts {
			z := strings.Split(part, ":")
			metadata[z[0]] = z[1]
		}
	}
	return metadata
}

func test() {
	senderPublicKey, senderPrivateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	recipientPublicKey, recipientPrivateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	// The shared key can be used to speed up processing when using the same
	// pair of keys repeatedly.
	senderSharedEncryptKey := new([32]byte)
	box.Precompute(senderSharedEncryptKey, recipientPublicKey, senderPrivateKey)
	fmt.Println(hex.EncodeToString(senderSharedEncryptKey[:]))

	recipientSharedEncryptKey := new([32]byte)
	box.Precompute(recipientSharedEncryptKey, senderPublicKey, recipientPrivateKey)
	fmt.Println(hex.EncodeToString(senderSharedEncryptKey[:]))
}

func main() {
	session := sdk.NewSession(&http.Client{Timeout: 5 * time.Second})

	parseVars()
	acct, _ := session.RestoreAccountFromSeed(seed)
	fmt.Println("Account Number: ", acct.AccountNumber())

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
		var bitmarkIds []string
		var err error
		if filepath != "" {
			af, _ := sdk.NewAssetFile(filepath, sdk.Accessibility(acs))
			if name != "" {
				af.Describe(name, toMedatadata())
			}
			fmt.Println("Asset ID:", af.Id())
			bitmarkIds, err = acct.IssueByAssetFile(af, quantity)
		} else {
			bitmarkIds, err = acct.IssueByAssetId(assetId, quantity)
		}

		if err != nil {
			fmt.Println("issue failed: ", err)
			return
		}
		fmt.Println("Bitmark ID:")
		for i, id := range bitmarkIds {
			fmt.Printf("\t[%d] %s", i, id)
		}
		fmt.Println()
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
		fmt.Println("File Name:", fileName)
		fmt.Println("File Content:", string(content))
		// file, _ := os.Create(path + "/" + fileName)
		// file.Write(content)
		// file.Close()
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
