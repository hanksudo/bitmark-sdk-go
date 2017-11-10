package bitmarksdk

import "errors"

func (acct *Account) IssueNewBitmarks(fileURL string, acs Accessibility, propertyName string, propertyMetadata map[string]string, quantity int) (string, []string, error) {
	af, err := readAssetFile(fileURL)
	if err != nil {
		return "", nil, err
	}

	if err := acct.api.UploadAsset(acct, af, acs); err != nil {
		return "", nil, err
	}

	asset, err := NewAssetRecord(propertyName, af.Fingerprint, propertyMetadata, acct)
	if err != nil {
		return "", nil, err
	}

	issues, err := NewIssueRecords(asset.Id(), acct, quantity)
	if err != nil {
		return "", nil, err
	}
	bitmarkIds, err := acct.api.issue(asset, issues)
	return asset.Id(), bitmarkIds, err
}

func (acct *Account) TransferBitmark(bitmarkId, receiver string) (string, error) {
	access, err := acct.api.getAssetAccess(acct, bitmarkId)
	if err != nil {
		return "", err
	}

	if access.SessData != nil {
		senderPublicKey, err := acct.api.getEncPubkey(access.Sender)
		if err != nil {
			return "", err
		}

		dataKey, err := dataKeyFromSessionData(acct, access.SessData, senderPublicKey)
		if err != nil {
			return "", err
		}

		recipientEncrPubkey, err := acct.api.getEncPubkey(receiver)
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

func (a *Account) DownloadAsset(bitmarkId string) ([]byte, error) {
	return a.api.DownloadAsset(a, bitmarkId)
}
