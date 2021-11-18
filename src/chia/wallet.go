package chia

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"chia_monitor/src/utils"
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

type WalletRpcResult struct {
	WalletBalance struct {
		ConfirmedWalletBalance   int `json:"confirmed_wallet_balance"`
		MaxSendAmount            int `json:"max_send_amount"`
		PendingChange            int `json:"pending_change"`
		PendingCoinRemovalCount  int `json:"pending_coin_removal_count"`
		SpendableBalance         int `json:"spendable_balance"`
		UnconfirmedWalletBalance int `json:"unconfirmed_wallet_balance"`
		UnspentCoinCount         int `json:"unspent_coin_count"`
		WalletID                 int `json:"wallet_id"`
	} `json:"wallet_balance"`
	Error   string `json:"error"`
	Success bool   `json:"success"`
}

// GetWalletBalance 获取钱包余额
func (w Wallet) GetWalletBalance() (walletRpcResult WalletRpcResult, err error) {
	url := w.BaseUrl + "get_wallet_balance"
	walletId := WalletId{WalletId: w.WalletId}
	//发起请求
	resp, err := utils.PostHttps(url, walletId, "application/json", w.CertPath, w.KeyPath)
	if err != nil {
		log.Error(err)
		return
	}
	log.Debug(string(resp))

	err = json.Unmarshal(resp, &walletRpcResult)

	return walletRpcResult, err
}
