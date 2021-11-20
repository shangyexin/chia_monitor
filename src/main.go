package main

import (
	"flag"

	log "github.com/sirupsen/logrus"

	"chia_monitor/src/chia"
	"chia_monitor/src/config"
	"chia_monitor/src/logger"
)

const blockChainUrl = "https://127.0.0.1:8555/"
const walletUrl = "https://127.0.0.1:9256/"
const farmerUrl = "https://127.0.0.1:8559/"

var (
	nodeEventTest   bool
	walletEventTest bool
	farmerEventTest bool
)

func main() {
	log.Info("Start chia monitor...")
	//获取命令行参数
	flag.Parse()

	//测试节点事件
	if nodeEventTest {
		chia.TestNodeEvent()
		return
	}

	//获取配置文件
	cfg := config.GetConfig()

	//区块链对象
	blockChain := chia.BlockChain{
		BaseUrl:  blockChainUrl,
		CertPath: cfg.FullNodeCertPath.CertPath,
		KeyPath:  cfg.FullNodeCertPath.KeyPath,
		WalletId: 1,
	}
	//监控区块链状态
	go chia.MonitorBlockState(blockChain)

	//钱包对象
	wallet := chia.Wallet{
		BaseUrl:  walletUrl,
		CertPath: cfg.WalletCertPath.CertPath,
		KeyPath:  cfg.WalletCertPath.KeyPath,
		WalletId: 1,
	}
	//监控钱包状态
	go chia.MonitorWallet(wallet)

	//农民对象
	farmer := chia.Farmer{
		BaseUrl:               farmerUrl,
		CertPath:              cfg.WalletCertPath.CertPath,
		KeyPath:               cfg.WalletCertPath.KeyPath,
		IsSearchForPrivateKey: false,
	}
	//监控耕种状态
	go chia.MonitorFarmer(farmer)

	//监控矿池状态
	go chia.MonitorPool(farmer)

	//监控矿池收益
	go chia.MonitorPoolEarning(cfg.PoolName)

	select {}
}

func init() {
	flag.BoolVar(&nodeEventTest, "n", false, "测试node事件")
	flag.BoolVar(&walletEventTest, "w", false, "测试wallet事件")
	flag.BoolVar(&farmerEventTest, "f", false, "测试farmer事件")
	//获取配置文件
	cfg := config.GetConfig()
	//初始化日志模块
	logger.InitLog(cfg.LogConfig.LogDir, cfg.LogConfig.AppName, cfg.LogConfig.IsProduction)
}
