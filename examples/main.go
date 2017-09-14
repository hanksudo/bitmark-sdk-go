package main

import (
	"fmt"
	"strings"
	"time"

	bitmarksdk "github.com/bitmark-inc/bitmark-sdk-go"
)

func main() {
	// session is used to store configurations and credentials for using bitmark services
	sess := bitmarksdk.NewSession(bitmarksdk.Testnet, "apiId", "apiSecret")

	// create 2 accounts
	accountA, _ := sess.NewAccount()
	accountB, _ := sess.NewAccount()

	// print recovery phrase for account A
	fmt.Println(strings.Join(accountA.RecoveryPhrase(), " "))
	fmt.Println(strings.Join(accountB.RecoveryPhrase(), " "))

	// issue 2 bitmarks under account A
	bitmarks, _ := sess.IssueBitmarks(
		accountA,                                      // issuer account
		"Bitmark 101",                                 // asset name
		[]byte(fmt.Sprint(time.Now().Unix())),         // asset content (use current time to represent a new asset for demo)
		map[string]string{"desc": "bitmark tutorial"}, // asset metadata
		2,
	)
	fmt.Printf("\nBitmarks under account A %s\n", accountA.AccountNumber())
	for _, bitmark := range bitmarks {
		fmt.Printf("[bitmark] id: %s\n", bitmark.Id)
	}

	// waiting for the first bitmark to be confirmed
	ticker := time.NewTicker(time.Second * 30)
	stop := make(chan bool, 1)
	go func() {
		for {
			select {
			case <-ticker.C:
				if bitmarks[0].Update(sess); bitmarks[0].Status == "confirmed" {
					stop <- true
					return
				}
			}
		}
	}()
	<-stop

	// transfer the first bitmark to another account B
	sess.TransferBitmark(accountA, bitmarks[0].Id, accountB.AccountNumber())

	filter := bitmarksdk.Filter{
		Owner:   accountA.AccountNumber(),
		Pending: true,
	}
	size := 10
	total := 100
	descending := true

	// list bitmarks under account A
	fmt.Printf("\nBitmarks under account A %s\n", accountA.AccountNumber())
	selector := sess.ListBitmarks(filter, size, total, descending)
	for selector.Next() {
		if selector.Err != nil {
			fmt.Println(selector.Err.Error())
		}

		for _, bitmark := range selector.Items.([]bitmarksdk.Bitmark) {
			fmt.Printf("[bitmark] id: %s status: %s \n", bitmark.Id, bitmark.Status)
		}
	}

	// list bitmarks under account B
	filter.Owner = accountB.AccountNumber()
	fmt.Printf("\nBitmarks under account B %s\n", accountB.AccountNumber())
	selector = sess.ListBitmarks(filter, size, total, descending)
	for selector.Next() {
		if selector.Err != nil {
			fmt.Println(selector.Err.Error())
		}

		for _, bitmark := range selector.Items.([]bitmarksdk.Bitmark) {
			fmt.Printf("[bitmark] id: %s status: %s \n", bitmark.Id, bitmark.Status)
		}
	}
}
