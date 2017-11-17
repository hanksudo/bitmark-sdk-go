package bitmarksdk

import (
	"errors"
	"fmt"
)

func (acct *Account) IssueBitmarks(fileURL string, acs Accessibility, propertyName string, propertyMetadata map[string]string, quantity int) (string, []string, error) {
	af, err := readAssetFile(fileURL)
	if err != nil {
		return "", nil, err
	}

	if uerr := acct.api.uploadAsset(acct, af, acs); uerr != nil {
		switch uerr.Error() {
		case "asset should have been uploaded":
			// TODO: might need to notify SDK users that the asset is already registered
		default:
			return "", nil, err
		}
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

		err = acct.api.addSessionData(acct, bitmarkId, receiver, data)
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

func (acct *Account) DownloadAsset(bitmarkId string) (string, []byte, error) {
	access, err := acct.api.getAssetAccess(acct, bitmarkId)
	if err != nil {
		return "", nil, err
	}

	fileName, content, err := acct.api.getAssetContent(access.URL)
	if err != nil {
		return "", nil, err
	}

	if access.SessData == nil { // public asset
		return fileName, content, nil
	}

	encrPubkey, err := acct.api.getEncPubkey(access.Sender)
	if err != nil {
		return "", nil, fmt.Errorf("fail to get enc public key: %s", err.Error())
	}

	dataKey, err := dataKeyFromSessionData(acct, access.SessData, encrPubkey)
	if err != nil {
		return "", nil, err
	}

	plaintext, err := dataKey.Decrypt(content)
	if err != nil {
		return "", nil, err
	}

	return fileName, plaintext, nil
}
