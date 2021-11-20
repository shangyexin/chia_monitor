package main

import (
	"flag"

	log "github.com/sirupsen/logrus"

	"chia_monitor/src/chia"
	"chia_monitor/src/config"
	"chia_monitor/src/logger"
)

var (
	nodeEventTest   bool
	walletEventTest bool
	farmerEventTest bool
)

func main() {
	//获取配置文件
	cfg := config.GetConfig()
	log.Infof("Start %s monitor...", cfg.Coin.Name)

	//获取命令行参数
	flag.Parse()
	//测试节点事件
	if nodeEventTest {
		chia.TestNodeEvent()
		return
	}

	//区块链对象
	blockChain := chia.BlockChain{
		BaseUrl:  cfg.Coin.BlockChainRpcUrl,
		CertPath: cfg.FullNodeCertPath.CertPath,
		KeyPath:  cfg.FullNodeCertPath.KeyPath,
		WalletId: 1,
	}
	//监控区块链状态
	go chia.MonitorBlockState(blockChain)

	//钱包对象
	wallet := chia.Wallet{
		BaseUrl:  cfg.Coin.WalletRpcUrl,
		CertPath: cfg.WalletCertPath.CertPath,
		KeyPath:  cfg.WalletCertPath.KeyPath,
		WalletId: 1,
	}
	//监控钱包状态
	go chia.MonitorWallet(wallet)

	//农民对象
	farmer := chia.Farmer{
		BaseUrl:               cfg.Coin.FarmerRpcUrl,
		CertPath:              cfg.WalletCertPath.CertPath,
		KeyPath:               cfg.WalletCertPath.KeyPath,
		IsSearchForPrivateKey: false,
	}
	//监控耕种状态
	go chia.MonitorFarmer(farmer)

	//新协议，支持矿池
	if cfg.Monitor.IsSupportPool {
		//监控矿池状态
		go chia.MonitorPool(farmer)

		//监控矿池收益
		go chia.MonitorPoolEarning(cfg.PoolName)
	}

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
