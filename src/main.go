package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"
	"time"

	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"

	"chia_monitor/src/chia"
	"chia_monitor/src/config"
	"chia_monitor/src/logger"
	"chia_monitor/src/utils"
	"chia_monitor/src/wechat"
)

const blockChainUrl = "https://127.0.0.1:6755/"
const walletUrl = "https://127.0.0.1:6761/"
const farmerUrl = "https://127.0.0.1:6759/"

const restartChiaCmd = "/root/restart.sh"
const syncingCountMax = 6

var (
	nodeEventTest   bool
	walletEventTest bool
	farmerEventTest bool
)

func main() {
	log.Info("Start flax monitor...")
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

	//农民对象
	farmer := chia.Farmer{
		BaseUrl:               farmerUrl,
		CertPath:              cfg.WalletCertPath.CertPath,
		KeyPath:               cfg.WalletCertPath.KeyPath,
		IsSearchForPrivateKey: false,
	}
	//监控农民状态
	go monitorFarmer(farmer)

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
	var timestamp int
	var blockRecordRpcResult chia.BlockRecordRpcResult

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
			remark = "已自动重启Flax"
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
					detail = "重启Flax后恢复"
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
				if blockchainStateRpcResult.BlockchainState.Peak.Timestamp != 0 {
					timestamp = blockchainStateRpcResult.BlockchainState.Peak.Timestamp
				} else {
					peakHash := blockchainStateRpcResult.BlockchainState.Peak.HeaderHash
					log.Debugf("peakHash: %+v", peakHash)
					blockRecordRpcResult, err = blockChain.GetBlockRecord(peakHash)
					if err != nil {
						log.Error("Get block record error: ", err)
						//发送错误通知
						event = "RPC获取区块记录错误"
						detail = err.Error()
						wechat.SendChiaMonitorNoticeToWechat(machineName, event, detail, remark)
						//等待间隔时间后重新查询
						time.Sleep(time.Duration(cfg.BockChainInterval) * time.Minute)
						continue
					}
					if !blockRecordRpcResult.Success {
						log.Error("Get block record failed: ", err)
						//发送错误通知
						event = "RPC获取区块记录失败"
						detail = blockRecordRpcResult.Error
						wechat.SendChiaMonitorNoticeToWechat(machineName, event, detail, remark)
						//等待间隔时间后重新查询
						time.Sleep(time.Duration(cfg.BockChainInterval) * time.Minute)
						continue
					}
					getBlockRecordCount := 0
					for {
						if blockRecordRpcResult.BlockRecord.Timestamp != 0 {
							timestamp = blockRecordRpcResult.BlockRecord.Timestamp
							log.Info("Get block timestamp success, getBlockRecordCount: ", getBlockRecordCount)
							break
						}
						blockRecordRpcResult, err = blockChain.GetBlockRecord(blockRecordRpcResult.BlockRecord.PrevHash)
						getBlockRecordCount = getBlockRecordCount + 1
						if err != nil {
							log.Error("Get block record error: ", err)
							//发送错误通知
							event = "RPC获取区块记录错误"
							detail = err.Error()
							wechat.SendChiaMonitorNoticeToWechat(machineName, event, detail, remark)
							break
						}
						if !blockRecordRpcResult.Success {
							log.Error("Get block record failed: ", blockRecordRpcResult.Error)
							//发送错误通知
							event = "RPC获取区块记录失败"
							detail = blockRecordRpcResult.Error
							wechat.SendChiaMonitorNoticeToWechat(machineName, event, detail, remark)
							break
						}
					}
				}
				//重试次数+1
				if syncingCount < syncingCountMax {
					log.Debugf("Retry count: %d", syncingCount)
					syncingCount = syncingCount + 1
					//需要等待自动恢复
					isNeedAutoRecover = true
					//发送区块链未同步，正等待自动恢复微信通知
					event = "区块链未同步"
					currentBlockTime := time.Unix(int64(timestamp), 0).Format("2006-01-02 15:04:05")
					detail = fmt.Sprintf("第%d次等待自动恢复，当前最新区块时间：%s",
						syncingCount,
						currentBlockTime)
					remark = ""
					wechat.SendChiaMonitorNoticeToWechat(machineName, event, detail, remark)
				} else {
					//发送区块链未同步，已经重新启动微信通知
					event = "区块链未同步"
					detail = "已经达到最大等待次数，立即重启Flax"
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
			log.Error("Get blockchain state rpc result failed: ", blockchainStateRpcResult.Error)
			//发送获取rpc失败微信通知
			event = "RPC获取区块链状态失败"
			detail = blockchainStateRpcResult.Error
			remark = "已自动重启Flax"
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
	var event string
	var detail string
	var remark string
	//获取配置文件
	cfg := config.GetConfig()
	machineName := cfg.Monitor.MachineName

	//创建定时任务
	c := cron.New()
	err := c.AddFunc(cfg.Monitor.WalletCron, func() {
		//获取钱包余额
		walletRpcResult, err := wallet.GetWalletBalance()
		if err != nil {
			log.Error("Get wallet balance err: ", err)
			//发送获取rpc失败微信通知
			event = "RPC获取钱包状态失败"
			detail = "RPC返回失败结果"
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
	select {}
}

//监控农民状态
func monitorFarmer(farmer chia.Farmer) {
	var event string
	var detail string
	var remark string
	var isFarming bool
	var harvesterOfflineCount int
	var host string

	//获取配置文件
	cfg := config.GetConfig()
	machineName := cfg.Monitor.MachineName

	for {
		//获取收割机状态
		harvestersRpcResult, err := farmer.GetHarvesters()
		if err != nil {
			log.Error("Get blockchain state failed: ", err)
			//发送错误通知
			event = "RPC获取收割机列表错误"
			detail = err.Error()
			wechat.SendChiaMonitorNoticeToWechat(machineName, event, detail, remark)
			//等待间隔时间后重新查询
			time.Sleep(time.Duration(cfg.BockChainInterval) * time.Minute)
			continue
		}
		//获取成功
		if harvestersRpcResult.Success {
			log.Info("Get harvest list rpc result success!")
			harvesterOfflineCount = 0
			event = "收割机掉线"
			detailTmp := "设备掉线："
			remark = "请及时登陆设备处理"
			//查找配置里面的本地+固定IP三台
			for _, harvesterMonitor := range cfg.Monitor.HarvesterList {
				isFarming = false
				address := net.ParseIP(harvesterMonitor)
				if address == nil {
					// 没有匹配上，实际为域名，需要解析ip地址
					addr, err := net.ResolveIPAddr("ip", harvesterMonitor)
					if err != nil {
						log.Errorf("%s resolve failed", harvesterMonitor)
						continue
					}
					host = addr.String()
					log.Debugf("%s resolve to ip is %s", harvesterMonitor, addr)
				} else {
					// 匹配成功，为IP地址
					host = address.String()
				}
				for _, harvester := range harvestersRpcResult.Harvesters {
					if host == harvester.Connection.Host {
						log.Debugf("%s is farming, ok", harvesterMonitor)
						isFarming = true
					}
				}
				if isFarming == false {
					log.Errorf("%s is not farming", harvesterMonitor)
					harvesterOfflineCount = harvesterOfflineCount + 1
					detailTmp = detailTmp + "\n" + harvesterMonitor
				}
			}
			//查找lj.yasin.store
			if harvesterOfflineCount > 0 {
				detail = fmt.Sprintf("%d台%s", harvesterOfflineCount, detailTmp)
				//不存在标识位，直接发送通知
				if !utils.Exists(cfg.Monitor.HarvesterOfflineFlag) {
					//发送错误通知
					wechat.SendChiaMonitorNoticeToWechat(machineName, event, detail, remark)
					//写入标识位文件
					err = ioutil.WriteFile(cfg.Monitor.HarvesterOfflineFlag, []byte(detail), 0644)
					if err != nil {
						log.Errorf("Write file [%s] failed: %s", cfg.Monitor.HarvesterOfflineFlag, err)
					} else {
						log.Infof("Write file [%s] success", cfg.Monitor.HarvesterOfflineFlag)
					}
				} else {
					//存在标识位，判断与当前记录的flag内容是否一致，有变化时才重新发送通知
					file, err := ioutil.ReadFile(cfg.Monitor.HarvesterOfflineFlag)
					if err != nil {
						log.Errorf("Open file [%s] failed: %s", cfg.Monitor.HarvesterOfflineFlag, err)
						continue
					}
					if detail == string(file) {
						log.Debug("Same harvester offline info, do not send notice to wechat")
					} else {
						log.Info("Different harvester offline info, send notice to wechat")
						//发送错误通知
						wechat.SendChiaMonitorNoticeToWechat(machineName, event, detail, remark)
						//写入标识位文件
						err = ioutil.WriteFile(cfg.Monitor.HarvesterOfflineFlag, []byte(detail), 0644)
						if err != nil {
							log.Errorf("Write file [%s] failed: %s", cfg.Monitor.HarvesterOfflineFlag, err)
						} else {
							log.Infof("Write file [%s] success", cfg.Monitor.HarvesterOfflineFlag)
						}
					}
				}
			} else {
				log.Info("All harvesters are online!")
			}
		} else {
			log.Error("Get blockchain state rpc result failed: ", harvestersRpcResult.Error)
			//发送获取rpc失败微信通知
			event = "RPC获取收割机列表失败"
			detail = harvestersRpcResult.Error
			wechat.SendChiaMonitorNoticeToWechat(machineName, event, detail, remark)
		}

		time.Sleep(time.Duration(cfg.Monitor.FarmerInterval) * time.Minute)
	}

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
