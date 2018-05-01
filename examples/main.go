package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	sdk "github.com/hanksudo/bitmark-sdk-go"
	imgcat "github.com/martinlindhe/imgcat/lib"
)

var (
	client *sdk.Client

	issuerSeed   string
	senderSeed   string
	receiverSeed string
	ownerSeed    string

	// issue
	assetPath string
	acs       string

	assetId string

	name        string
	rawMetadata string

	quantity int

	// transfer
	bitmarkId string
)

func parseVars() {
	subcmd := flag.NewFlagSet("subcmd", flag.ExitOnError)

	subcmd.StringVar(&issuerSeed, "issuer", "5XEECttxvRBzxzAmuV4oh6T1FcQu4mBg8eWd9wKbf8hweXsfwtJ8sfH", "Issuer Seed")
	subcmd.StringVar(&senderSeed, "sender", "5XEECttxvRBzxzAmuV4oh6T1FcQu4mBg8eWd9wKbf8hweXsfwtJ8sfH", "Sender Seed")
	subcmd.StringVar(&receiverSeed, "receiver", "5XEECt4yuMK4xqBLr9ky5FBWpkAR6VHNZSz8fUzZDXPnN3D9MeivTSA", "Receiver Seed")
	subcmd.StringVar(&ownerSeed, "owner", "5XEECttxvRBzxzAmuV4oh6T1FcQu4mBg8eWd9wKbf8hweXsfwtJ8sfH", "Owner Seed")

	subcmd.StringVar(&assetPath, "p", "", "")
	subcmd.StringVar(&acs, "acs", "public", "")
	subcmd.StringVar(&name, "name", "", "")
	subcmd.StringVar(&rawMetadata, "meta", "", "")
	subcmd.StringVar(&assetId, "aid", "", "Asset ID")
	subcmd.IntVar(&quantity, "quantity", 1, "")

	subcmd.StringVar(&bitmarkId, "bid", "", "Bitmark ID")

	if len(os.Args) < 2 {
		subcmd.PrintDefaults()
		os.Exit(2)
	}

	subcmd.Parse(os.Args[2:])
}

func toMedatadata() map[string]string {
	parts := strings.Split(rawMetadata, ",")
	metadata := make(map[string]string)
	if len(parts) > 0 {
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
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
		Network:    "testnet",
	}
	client = sdk.NewClient(cfg)
	log.Println("You are now on", client.Network)

	switch os.Args[1] {
	case "new-account":
		account, _ := client.CreateAccount()
		fmt.Println("Account Number:", account.AccountNumber())
		fmt.Println("-> seed:", account.Seed())
		fmt.Println("-> recovery phrase:", strings.Join(account.RecoveryPhrase(), " "))
	case "issue": // -p=<file path> -name=<name> -meta=<key1:val1,key2:val2> -acs=<accessibility> -quantity=<quantity>
		issuer, _ := client.RestoreAccountFromSeed(issuerSeed)
		fmt.Println("issuer:", issuer.AccountNumber())

		if assetPath == "" || !pathExists(assetPath) {
			panic("asset file not specified")
		}

		af, _ := sdk.NewAssetFileFromPath(assetPath, sdk.Accessibility(acs))

		assetInfo := &sdk.AssetInfo{
			Name:     name,
			Metadata: toMedatadata(),
		}

		fmt.Println("Asset ID:", af.Id())

		bitmarkIds, err := client.IssueByAssetFile(issuer, af, quantity, assetInfo)
		if err != nil {
			panic(err)
		}

		fmt.Println("Bitmark IDs:")
		for i, id := range bitmarkIds {
			fmt.Printf("\t[%d] %s\n", i, id)
		}
	case "issue-asset-id": // -aid=<asset id>
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
	case "transfer-offer": // -bid=<bitmark id>
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

		offerId, err := client.SubmitTransferOffer(sender, offer, map[string]interface{}{})
		if err != nil {
			panic(err)
		}

		fmt.Printf("transfer offer id: %s\n", offerId)

	case "bitmark":
		owner, err := client.RestoreAccountFromSeed(ownerSeed)
		if err != nil {
			panic(err)
		}
		fmt.Println("owner:", owner.AccountNumber())
		bitmark, err := client.GetBitmark(bitmarkId)
		if err != nil {
			panic(err)
		}
		spew.Dump(bitmark)
	case "bitmarks":
		issuer, err := client.RestoreAccountFromSeed(issuerSeed)
		if err != nil {
			panic(err)
		}
		filter := sdk.BitmarkFilter{
			Owner: issuer.AccountNumber(),
			Limit: 100,
			Asset: true,
		}
		bitmarks, err := client.QueryBitmarks(&filter)
		fmt.Printf("Had %d bitmarks.\n", len(bitmarks))
		if err != nil {
			panic(err)
		}

		for _, bitmark := range bitmarks {
			spew.Dump(bitmark)
		}
	case "download":
		owner, err := client.RestoreAccountFromSeed(ownerSeed)
		if err != nil {
			panic(err)
		}
		fmt.Println("owner:", owner.AccountNumber())

		fileName, content, err := client.DownloadAsset(owner, bitmarkId)
		if err != nil {
			fmt.Println("download failed: ", err)
			return
		}
		fmt.Println("File Name:", fileName)
		fmt.Println("File Content:", len(content))

		if include([]string{".jpg", ".png"}, filepath.Ext(fileName)) {
			f, _ := os.Create(fileName)
			f.Write(content)
			defer f.Close()

			imgcat.CatFile(fileName, os.Stdout)
		}
	}
}

func pathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// Index - returns the first index of the target string
// or -1 if no match is found
func index(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

// Include - returns true if the target string t is in the slice.
func include(vs []string, t string) bool {
	return index(vs, t) >= 0
}
