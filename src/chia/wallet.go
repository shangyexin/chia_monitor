package chia

import (
	"chia_monitor/src/utils"
	log "github.com/sirupsen/logrus"
)

type WalletId struct {
	WalletId int `json:"wallet_id"`
}

type Wallet struct {
	BaseUrl  string
	CertPath string
	KeyPath  string
	WalletId int
}

// GetWalletBalance 获取钱包余额
func (w Wallet) GetWalletBalance() {
	url := w.BaseUrl + "get_wallet_balance"
	walletId := WalletId{WalletId: w.WalletId}
	//发起请求
	resp, err := utils.PostHttps(url, walletId, "application/json", w.CertPath, w.KeyPath)
	if err != nil {
		log.Error(err)
		return
	}
	log.Info(string(resp))
}
