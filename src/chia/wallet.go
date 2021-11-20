package chia

import (
	"encoding/json"
	"fmt"

	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"

	"chia_monitor/src/config"
	"chia_monitor/src/utils"
	"chia_monitor/src/wechat"
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

//MonitorWallet 监控钱包状态
func MonitorWallet(wallet Wallet) {
	var event string
	var detail string
	var remark string
	//获取配置文件
	cfg := config.GetConfig()
	machineName := cfg.Monitor.MachineName

	//创建定时任务
	c := cron.New()
	err := c.AddFunc(cfg.Monitor.DailyCron, func() {
		//获取钱包余额
		walletRpcResult, err := wallet.GetWalletBalance()
		if err != nil {
			log.Error("Get wallet balance err: ", err)
			//发送获取rpc失败微信通知
			event = "RPC获取钱包状态失败"
			detail = err.Error()
			wechat.SendChiaMonitorNoticeToWechat(machineName, event, detail, remark)
		} else {
			//获取成功
			if walletRpcResult.Success {
				log.Info("Get waller state rpc result success!")
				//发送获取钱包余额微信通知
				event = "钱包余额查询"
				detail = fmt.Sprintf("钱包余额: %.12f", float64(walletRpcResult.WalletBalance.ConfirmedWalletBalance)/float64(1000000000000))
				remark = "获取钱包余额成功"
				wechat.SendChiaMonitorNoticeToWechat(machineName, event, detail, remark)
			} else {
				log.Info("Get waller state rpc result failed: ", walletRpcResult.Error)
				//发送获取钱包余额微信通知
				event = "钱包余额查询"
				detail = walletRpcResult.Error
				remark = "获取钱包余额失败"
				wechat.SendChiaMonitorNoticeToWechat(machineName, event, detail, remark)
			}
		}
	})

	if err != nil {
		log.Fatal("Start wallet cron task err: ", err)
		return
	}

	c.Start()
	log.Info("Start wallet cron task success!")
}
