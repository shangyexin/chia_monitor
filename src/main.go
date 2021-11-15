package main

import (
	"flag"
	"fmt"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"

	"chia_monitor/src/chia"
	"chia_monitor/src/config"
	"chia_monitor/src/logger"
	"chia_monitor/src/test"
	"chia_monitor/src/wechat"
)

const blockChainUrl = "https://127.0.0.1:8555/"
const walletUrl = "https://127.0.0.1:9256/"
const farmerUrl = "https://127.0.0.1:8559/"

const restartChiaCmd = "/root/restart.sh"
const syncingCountMax = 6

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
		test.TestNodeEvent()
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
	go monitorBlockState(blockChain)

	//钱包对象
	wallet := chia.Wallet{
		BaseUrl:  walletUrl,
		CertPath: cfg.WalletCertPath.CertPath,
		KeyPath:  cfg.WalletCertPath.KeyPath,
		WalletId: 1,
	}
	//监控钱包状态
	go monitorWallet(wallet)

	////农民对象
	//farmer := chia.Farmer{
	//	BaseUrl:               farmerUrl,
	//	CertPath:              cfg.WalletCertPath.CertPath,
	//	KeyPath:               cfg.WalletCertPath.KeyPath,
	//	IsSearchForPrivateKey: false,
	//}
	////监控农民状态
	//go monitorFarmer(farmer)

	select {}
}

//监控区块链状态
func monitorBlockState(blockChain chia.BlockChain) {
	var iSRestarted bool
	var isNeedAutoRecover bool
	var syncingCount int
	var event string
	var detail string
	var remark string

	//获取配置文件
	cfg := config.GetConfig()
	machineName := cfg.Monitor.MachineName

	for {
		//获取区块链状态
		blockchainStateRpcResult, err := blockChain.GetBlockchainState()
		if err != nil {
			log.Error("Get blockchain state failed: ", err)
			//发送错误通知
			event = "RPC获取区块链状态错误"
			detail = err.Error()
			remark = "已自动重启Chia"
			wechat.SendChiaMonitorNoticeToWechat(machineName, event, detail, remark)
			//重启Chia
			restartChia()
			iSRestarted = true
			//等待间隔时间后重新查询
			time.Sleep(time.Duration(cfg.BockChainInterval) * time.Minute)
			continue
		}
		//获取成功
		if blockchainStateRpcResult.Success {
			log.Info("Get blockchain state rpc result success!")
			//区块链已同步
			if blockchainStateRpcResult.BlockchainState.Sync.Synced {
				log.Info("Blockchain is synced!")
				//重启后恢复
				if iSRestarted {
					// 发送重启恢复微信通知
					event = "区块链同步成功"
					detail = "重启Chia后恢复"
					remark = ""
					wechat.SendChiaMonitorNoticeToWechat(machineName, event, detail, remark)
				} else if isNeedAutoRecover {
					//未同步计数清零
					syncingCount = 0
					// 发送自动恢复微信通知
					event = "区块链同步成功"
					detail = "等待间隔后自动恢复"
					remark = ""
					wechat.SendChiaMonitorNoticeToWechat(machineName, event, detail, remark)
				}
				iSRestarted = false
				isNeedAutoRecover = false
			} else {
				//区块链未同步
				log.Error("Blockchain is not synced!")
				log.Infof("Blockchain sync tip height: %d, sync progress height:%d",
					blockchainStateRpcResult.BlockchainState.Sync.SyncTipHeight,
					blockchainStateRpcResult.BlockchainState.Sync.SyncProgressHeight)
				//重试次数+1
				if syncingCount < syncingCountMax {
					log.Debugf("Retry count: %d", syncingCount)
					syncingCount = syncingCount + 1
					//需要等待自动恢复
					isNeedAutoRecover = true
					//发送区块链未同步，正等待自动恢复微信通知
					event = "区块链未同步"
					detail = fmt.Sprintf("第%d次等待自动恢复，目标区块高度：%d，当前区块高度: %d",
						syncingCount,
						blockchainStateRpcResult.BlockchainState.Sync.SyncTipHeight,
						blockchainStateRpcResult.BlockchainState.Sync.SyncProgressHeight)
					remark = ""
					wechat.SendChiaMonitorNoticeToWechat(machineName, event, detail, remark)
				} else {
					//发送区块链未同步，已经重新启动微信通知
					event = "区块链未同步"
					detail = "已经达到最大等待次数，立即重启Chia"
					remark = ""
					wechat.SendChiaMonitorNoticeToWechat(machineName, event, detail, remark)
					//等待syncingCountMax * blockChainInterval后都没有自动恢复，重启Chia
					restartChia()
					iSRestarted = true
					//重启后继续等待自动恢复，防止暂时未同步成功
					syncingCount = 0
				}
			}
		} else {
			//获取失败
			log.Debug("Get blockchain state rpc result failed!")
			//发送获取rpc失败微信通知
			event = "RPC获取区块链状态失败"
			detail = "RPC返回失败结果"
			remark = "已自动重启Chia"
			wechat.SendChiaMonitorNoticeToWechat(machineName, event, detail, remark)
			//重启Chia
			restartChia()
			iSRestarted = true
		}

		time.Sleep(time.Duration(cfg.Monitor.BockChainInterval) * time.Minute)
	}
}

//监控钱包状态
func monitorWallet(wallet chia.Wallet) {
	//获取钱包余额
	wallet.GetWalletBalance()
}

//监控农民状态
func monitorFarmer(farmer chia.Farmer) {
	//获取收割机状态
	farmer.GetHarvesters()
}

//重启chia
func restartChia() {
	cmd := exec.Command(restartChiaCmd)
	//将root目录作为工作目录
	cmd.Dir = "/root"
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Execute restart chia cmd failed with error:\n%s", err.Error())
	} else {
		log.Infof("Execute restart chia cmd succedd with output:\n%s", string(output))
	}
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
