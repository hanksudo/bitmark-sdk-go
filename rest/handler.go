package main

import (
	"net/http"

	sdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/gin-gonic/gin"
)

func handleCreateAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		acct, _ := sdk.NewAccount(netork)

		c.JSON(http.StatusOK, gin.H{
			"account_number":  acct.AccountNumber(),
			"seed":            acct.Seed(),
			"recovery_phrase": acct.RecoveryPhrase(),
		})
	}
}

type issueRequest struct {
	Seed          string            `json:"seed"`
	FilePath      string            `json:"file_path"`
	Accessibility sdk.Accessibility `json:"accessibility"`
	Name          string            `json:"property_name"`
	Metadata      map[string]string `json:"property_metadata"`
	Quantity      int               `json:"quantity"`
}

func handleIssue() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req issueRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
			return
		}

		acct, err := sdk.AccountFromSeed(req.Seed)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid seed"})
			return
		}

		assetId, bitmarkIds, err := acct.IssueNewBitmarks(req.FilePath, req.Accessibility, req.Name, req.Metadata, req.Quantity)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid seed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"asset_id":    assetId,
			"bitmark_ids": bitmarkIds,
		})
	}
}

type transferRequest struct {
	Seed      string `json:"seed"`
	BitmarkId string `json:"bitmark_id"`
	Receiver  string `json:"receiver"`
}

func handleTransfer() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req transferRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
			return
		}

		acct, err := sdk.AccountFromSeed(req.Seed)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid seed"})
			return
		}

		txId, err := acct.TransferBitmark(req.BitmarkId, req.Receiver)
		if err != nil {
			log.Error(err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"message": err})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"txId": txId,
		})
	}
}

type downloadRequest struct {
	Seed      string `json:"seed"`
	BitmarkId string `json:"bitmark_id"`
}

func handleDownloadAsset() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req transferRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
			return
		}

		acct, err := sdk.AccountFromSeed(req.Seed)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid seed"})
			return
		}

		content, err := acct.DownloadAsset(req.BitmarkId)
		if err != nil {
			log.Error(err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"message": "download failed"})
			return
		}

		c.Data(http.StatusOK, http.DetectContentType(content), content)
		return
	}
}
