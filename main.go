package bitmarksdk

func (acct *Account) IssueNewBitmarks(fileURL string, acs Accessibility, propertyName string, propertyMetadata map[string]string, quantity int) ([]string, error) {
	af, err := readAssetFile(fileURL)
	if err != nil {
		return nil, err
	}

	if e := uploadAsset(acct, af, acs); e != nil {
		return nil, err
	}

	asset, err := NewAssetRecord(propertyName, af.Fingerprint, propertyMetadata, acct)
	if err != nil {
		return nil, err
	}

	issues, err := NewIssueRecords(asset.Id(), acct, quantity)
	if err != nil {
		return nil, err
	}

	return issue(asset, issues)
}

// func (acct *Account) TransferBitmark(bitmarkId, receiver string) (string, error) {
// transfer, err := record.NewTransfer(txId, receiver, acct.account)
// if err != nil {
// 	return "", err
// }
//
// return u.requester.Transfer(transfer)
// }
