package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	sdk "github.com/bitmark-inc/bitmark-sdk-go"
)

var (
	client *sdk.Client

	issuerSeed   string
	senderSeed   string
	receiverSeed string
	ownerSeed    string

	// issue
	filepath string
	acs      string

	assetId string

	name        string
	rawMetadata string

	quantity int

	// transfer
	bitmarkId string
)

func parseVars() {
	subcmd := flag.NewFlagSet("subcmd", flag.ExitOnError)

	subcmd.StringVar(&issuerSeed, "issuer", "5XEECttxvRBzxzAmuV4oh6T1FcQu4mBg8eWd9wKbf8hweXsfwtJ8sfH", "")
	subcmd.StringVar(&senderSeed, "sender", "5XEECttxvRBzxzAmuV4oh6T1FcQu4mBg8eWd9wKbf8hweXsfwtJ8sfH", "")
	subcmd.StringVar(&receiverSeed, "receiver", "5XEECt4yuMK4xqBLr9ky5FBWpkAR6VHNZSz8fUzZDXPnN3D9MeivTSA", "")
	subcmd.StringVar(&ownerSeed, "owner", "5XEECttxvRBzxzAmuV4oh6T1FcQu4mBg8eWd9wKbf8hweXsfwtJ8sfH", "")

	subcmd.StringVar(&filepath, "p", "", "")
	subcmd.StringVar(&acs, "acs", "public", "")
	subcmd.StringVar(&name, "name", "", "")
	subcmd.StringVar(&rawMetadata, "meta", "", "")
	subcmd.StringVar(&assetId, "aid", "", "")
	subcmd.IntVar(&quantity, "quantity", 1, "")

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

func main() {
	parseVars()

	cfg := &sdk.Config{
		HTTPClient:  &http.Client{Timeout: 5 * time.Second},
		Network:     "testnet",
		APIEndpoint: "https://api.test.bitmark.com",
		KeyEndpoint: "https://key.assets.test.bitmark.com",
	}
	client = sdk.NewClient(cfg)

	switch os.Args[1] {
	case "newacct":
		account, _ := client.CreateAccount()
		fmt.Println("Account Number:", account.AccountNumber())
		fmt.Println("-> seed:", account.Seed())
		fmt.Println("-> recovery phrase:", strings.Join(account.RecoveryPhrase(), " "))
	case "afile-issue": // -p=<file path> -name=<name> -meta=<key1:val1,key2:val2> -acs=<accessibility> -quantity=<quantity>
		issuer, _ := client.RestoreAccountFromSeed(issuerSeed)
		fmt.Println("issuer:", issuer.AccountNumber())

		if filepath == "" {
			panic("asset file not specified")
		}

		af, _ := sdk.NewAssetFileFromPath(filepath, sdk.Accessibility(acs))
		if name != "" {
			af.Describe(name, toMedatadata())
		}
		fmt.Println("Asset ID:", af.Id())

		bitmarkIds, err := client.IssueByAssetFile(issuer, af, quantity)
		if err != nil {
			panic(err)
		}

		fmt.Println("Bitmark IDs:")
		for i, id := range bitmarkIds {
			fmt.Printf("\t[%d] %s\n", i, id)
		}
	case "aid-issue": // -aid=<asset id>
		issuer, _ := client.RestoreAccountFromSeed(issuerSeed)
		fmt.Println("issuer:", issuer.AccountNumber())

		bitmarkIds, err := client.IssueByAssetId(issuer, assetId, quantity)
		if err != nil {
			panic(err)
		}

		fmt.Println("Bitmark IDs:")
		for i, id := range bitmarkIds {
			fmt.Printf("\t[%d] %s\n", i, id)
		}
	case "1sig-trf": // -bid=<bitmark id>
		sender, _ := client.RestoreAccountFromSeed(senderSeed)
		fmt.Println("sender:", sender.AccountNumber())
		receiver, _ := client.RestoreAccountFromSeed(receiverSeed)
		fmt.Println("receiver:", receiver.AccountNumber())

		txId, err := client.Transfer(sender, bitmarkId, receiver.AccountNumber())
		if err != nil {
			panic(err)
		}
		fmt.Println("Transaction ID: ", txId)
	case "2sig-trf": // -bid=<bitmark id>
		sender, _ := client.RestoreAccountFromSeed(senderSeed)
		fmt.Println("sender:", sender.AccountNumber())
		receiver, _ := client.RestoreAccountFromSeed(receiverSeed)
		fmt.Println("receiver:", receiver.AccountNumber())

		// sign by sender
		offer, err := client.SignTransferOffer(sender, bitmarkId, receiver.AccountNumber(), true)
		if err != nil {
			panic(err)
		}
		data, _ := json.Marshal(offer)
		fmt.Printf("transfer offer by sender: %s\n", string(data))

		// sign by receiver
		transfer, _ := offer.Countersign(receiver)
		txId, err := client.CountersignedTransfer(transfer)
		if err != nil {
			panic(err)
		}
		fmt.Println("Transaction ID: ", txId)
	case "download":
		owner, _ := client.RestoreAccountFromSeed(ownerSeed)
		fmt.Println("owner:", owner.AccountNumber())

		fileName, content, err := client.DownloadAsset(owner, bitmarkId)
		if err != nil {
			fmt.Println("download failed: ", err)
			return
		}
		fmt.Println("File Name:", fileName)
		fmt.Println("File Content:", string(content))
	}
}
