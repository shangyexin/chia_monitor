package main

import (
	"chia_monitor/src/chia"
	"chia_monitor/src/config"
	"chia_monitor/src/logger"
	"flag"
	log "github.com/sirupsen/logrus"
)

const blockChainUrl = "https://localhost:8555/"
const walletUrl = "https://localhost:9256/"
const farmerUrl = "https://localhost:8559/"

func main() {
	log.Info("Start chia monitor...")
	//获取配置文件
	cfg := config.GetConfig()

	//区块链对象
	blockChain := chia.BlockChain{
		BaseUrl:  blockChainUrl,
		CertPath: cfg.FullNodeCertPath.CertPath,
		KeyPath:  cfg.FullNodeCertPath.KeyPath,
		WalletId: 1,
	}
	//获取区块链状态
	blockchainStateRpcResult, err := blockChain.GetBlockchainState()
	if err != nil {
		log.Error("Get blockchain state failed: ", err)
		return
	}
	if blockchainStateRpcResult.Success == true {
		log.Debug("Get blockchain state rpc result success!")
		if blockchainStateRpcResult.BlockchainState.Sync.Synced == true {
			log.Debug("Block chain is synced!")
		} else {
			log.Error("Block chain is not synced!")
		}
	} else {
		log.Debug("Get blockchain state rpc result failed!")
	}
	////钱包对象
	//wallet := chia.Wallet{
	//	BaseUrl:  walletUrl,
	//	CertPath: cfg.WalletCertPath.CertPath,
	//	KeyPath:  cfg.WalletCertPath.KeyPath,
	//	WalletId: 1,
	//}
	////获取钱包余额
	//wallet.GetWalletBalance()
	//
	////农民对象
	//farmer := chia.Farmer{
	//	BaseUrl:               farmerUrl,
	//	CertPath:              cfg.WalletCertPath.CertPath,
	//	KeyPath:               cfg.WalletCertPath.KeyPath,
	//	IsSearchForPrivateKey: false,
	//}
	////获取收割机状态
	//farmer.GetHarvesters()
}

func init() {
	flag.Parse()
	//获取配置文件
	cfg := config.GetConfig()
	//初始化日志模块
	logger.InitLog(cfg.LogConfig.LogDir, cfg.LogConfig.AppName, cfg.LogConfig.IsProduction)
}
