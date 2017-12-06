package bitmarksdk

import (
	"errors"
	"fmt"
)

// Issue is about to create a new issue on the bitmark blockchain. There are two circumstances
// If an asset name is given, it will try to issue a brand new stuff. In this case, it first
// checks whether it is a duplocated asset in the blockchain. If not, the dedicated file will be
// upload. On the other hand, the api try to issue with an existed asset if the asset name is
// not provided. It will then check if an asset file is given. If so, it will upload that again.
// This case usually happens for a private asset.
// Finally, submit the issue record.
func (acct *Account) Issue(asset *Asset, quantity int) (string, []string, error) {
	// issue new bitmarks
	if asset.Name != "" {
		if a, err := acct.api.getAsset(asset.Id); err != nil && err.Error() != "Not Found" {
			return "", nil, fmt.Errorf("asset record already registered with name = '%s' and metata = '%s'", a.Name, a.Metadata)
		}

		if asset.File == nil {
			return "", nil, errors.New("asset file not provided")
		}

		if err := acct.api.uploadAsset(acct, asset); err != nil {
			return "", nil, fmt.Errorf("service failed: %s", err.Error())
		}

		aRecord, aerr := NewAssetRecord(asset.Name, asset.File.Fingerprint, asset.Metadata, acct)
		if aerr != nil {
			return "", nil, aerr
		}
		iRecords, err := NewIssueRecords(asset.Id, acct, quantity)
		if err != nil {
			return "", nil, err
		}
		bitmarkIds, err := acct.api.issue(aRecord, iRecords)
		return asset.Id, bitmarkIds, err
	}

	// issue more bitmarks
	if asset.File != nil {
		if err := acct.api.uploadAsset(acct, asset); err != nil {
			return "", nil, fmt.Errorf("service failed: %s", err.Error())
		}
	}

	iRecords, err := NewIssueRecords(asset.Id, acct, quantity)
	if err != nil {
		return "", nil, err
	}
	bitmarkIds, err := acct.api.issue(nil, iRecords)
	return asset.Id, bitmarkIds, err
}

// func (acct *Account) IssueBitmarks(fileURL string, acs Accessibility, propertyName string, propertyMetadata map[string]string, quantity int) (string, []string, error) {
// 	af, err := readAssetFile(fileURL)
// 	if err != nil {
// 		return "", nil, err
// 	}
//
// 	if uerr := acct.api.uploadAsset(acct, af, acs); uerr != nil {
// 		switch uerr.Error() {
// 		case "asset should have been uploaded":
// 			// TODO: might need to notify SDK users that the asset is already registered
// 		default:
// 			return "", nil, err
// 		}
// 	}
// 	asset, err := NewAssetRecord(propertyName, af.Fingerprint, propertyMetadata, acct)
// 	if err != nil {
// 		return "", nil, err
// 	}
//
// 	issues, err := NewIssueRecords(asset.Id(), acct, quantity)
// 	if err != nil {
// 		return "", nil, err
// 	}
// 	bitmarkIds, err := acct.api.issue(asset, issues)
// 	return asset.Id(), bitmarkIds, err
// }

// TransferBitmark will transfer a bitmark to others. It will check the owner of a bitmark
// which is going to transfer. If it is valid, a transfer request will be submitted.
// If the target bitmark is private, it will generate a new session data for the new
// receiver.
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

func (acct *Account) RentBitmark(bitmarkId, receiver string, days uint) error {
	access, err := acct.api.getAssetAccess(acct, bitmarkId)
	if access.SessData == nil {
		return errors.New("no need to rent public assets")
	}

	dataKey, err := dataKeyFromSessionData(acct, access.SessData, acct.EncrKey.PublicKeyBytes())
	if err != nil {
		return err
	}

	recipientEncrPubkey, err := acct.api.getEncPubkey(receiver)
	if err != nil {
		return err
	}

	data, err := createSessionData(acct, dataKey, recipientEncrPubkey)
	if err != nil {
		return err
	}

	return acct.api.updateLease(acct, bitmarkId, receiver, days, data)
}

func (acct *Account) ListLeases() ([]*accessByRenting, error) {
	return acct.api.listLeases(acct)
}

func (acct *Account) DownloadAssetByLease(access *accessByRenting) ([]byte, error) {
	req, _ := newAPIRequest("GET", access.URL, nil)
	content, err := acct.api.submitRequest(req, nil)
	if err != nil {
		return nil, err
	}

	encrPubkey, err := acct.api.getEncPubkey(access.Owner)
	if err != nil {
		return nil, fmt.Errorf("fail to get enc public key: %s", err.Error())
	}

	dataKey, err := dataKeyFromSessionData(acct, access.SessData, encrPubkey)
	if err != nil {
		return nil, err
	}

	plaintext, err := dataKey.Decrypt(content)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
