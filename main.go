package bitmarksdk

import "errors"

func (acct *Account) IssueNewBitmarks(fileURL string, acs Accessibility, propertyName string, propertyMetadata map[string]string, quantity int) ([]string, error) {
	af, err := readAssetFile(fileURL)
	if err != nil {
		return nil, err
	}

	if err := acct.api.UploadAsset(acct, af, acs); err != nil {
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

	return acct.api.issue(asset, issues)
}

func (acct *Account) TransferBitmark(bitmarkId, receiver string) (string, error) {
	access, err := acct.api.getAssetAccess(acct, bitmarkId)
	if err != nil {
		return "", err
	}

	senderPublicKey, err := getEncrPubkey(access.Sender)
	if err != nil {
		return "", err
	}

	dataKey, err := dataKeyFromSessionData(acct, access.SessData, senderPublicKey)
	if err != nil {
		return "", err
	}

	recipientEncrPubkey, err := getEncrPubkey(receiver)
	if err != nil {
		return "", err
	}

	data, err := createSessionData(acct, dataKey, recipientEncrPubkey)
	if err != nil {
		return "", err
	}

	err = acct.api.updateSession(acct, bitmarkId, receiver, data)
	if err != nil {
		return "", err
	}

	bmk, err := acct.api.getBitmark(bitmarkId)
	if err != nil {
		return "", err
	}

	if acct.AccountNumber() != bmk.Owner {
		return "", errors.New("not bitmark owner")
	}

	tr, err := NewTransferRecord(bmk.HeadId, receiver, acct)
	if err != nil {
		return "", err
	}

	return acct.api.transfer(tr)
}
