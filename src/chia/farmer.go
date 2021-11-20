package chia

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"

	"chia_monitor/src/config"
	"chia_monitor/src/utils"
	"chia_monitor/src/wechat"
)

type SearchForPrivateKey struct {
	SearchForPrivateKey bool `json:"search_for_private_key"`
}

type Farmer struct {
	BaseUrl               string
	CertPath              string
	KeyPath               string
	IsSearchForPrivateKey bool
}

type HarvestersRpcResult struct {
	Harvesters []struct {
		Connection struct {
			Host   string `json:"host"`
			NodeID string `json:"node_id"`
			Port   int    `json:"port"`
		} `json:"connection"`
		FailedToOpenFilenames []interface{} `json:"failed_to_open_filenames"`
		NoKeyFilenames        []interface{} `json:"no_key_filenames"`
		Plots                 []interface{} `json:"plots"`
	} `json:"harvesters"`
	Error   string `json:"error"`
	Success bool   `json:"success"`
}

// PoolStateRpcResult 获取矿池状态
type PoolStateRpcResult struct {
	PoolState []struct {
		AuthenticationTokenTimeout   int         `json:"authentication_token_timeout"`
		CurrentDifficulty            int         `json:"current_difficulty"`
		CurrentPoints                int         `json:"current_points"`
		NextFarmerUpdate             float64     `json:"next_farmer_update"`
		NextPoolInfoUpdate           float64     `json:"next_pool_info_update"`
		P2SingletonPuzzleHash        string      `json:"p2_singleton_puzzle_hash"`
		PointsAcknowledged24H        [][]float64 `json:"points_acknowledged_24h"`
		PointsAcknowledgedSinceStart int         `json:"points_acknowledged_since_start"`
		PointsFound24H               [][]float64 `json:"points_found_24h"`
		PointsFoundSinceStart        int         `json:"points_found_since_start"`
		PoolConfig                   struct {
			AuthenticationPublicKey string `json:"authentication_public_key"`
			LauncherID              string `json:"launcher_id"`
			OwnerPublicKey          string `json:"owner_public_key"`
			P2SingletonPuzzleHash   string `json:"p2_singleton_puzzle_hash"`
			PayoutInstructions      string `json:"payout_instructions"`
			PoolURL                 string `json:"pool_url"`
			TargetPuzzleHash        string `json:"target_puzzle_hash"`
		} `json:"pool_config"`
		PoolErrors24H []struct {
			ErrorCode    int    `json:"error_code"`
			ErrorMessage string `json:"error_message"`
		} `json:"pool_errors_24h"`
	} `json:"pool_state"`
	Error   string `json:"error"`
	Success bool   `json:"success"`
}

//MonitorFarmer 监控耕种状态
func MonitorFarmer(farmer Farmer) {
	var event string
	var detail string
	var remark string
	var isFarming bool
	var harvesterOfflineCount int
	var host string

	//获取配置文件
	cfg := config.GetConfig()
	machineName := cfg.Monitor.MachineName
	event = "耕种状态监控"
	log.Info("Start to monitor farmer...")

	for {
		//获取收割机状态
		harvestersRpcResult, err := farmer.GetHarvesters()
		if err != nil {
			log.Error("Get blockchain state failed: ", err)
			//发送错误通知
			detail = err.Error()
			remark = "获取收割机列表错误"
			wechat.SendChiaMonitorNoticeToWechat(machineName, event, detail, remark)
			//等待间隔时间后重新查询
			time.Sleep(time.Duration(cfg.BockChainInterval) * time.Minute)
			continue
		}
		//获取成功
		if harvestersRpcResult.Success {
			log.Info("Get harvest list rpc result success!")
			harvesterOfflineCount = 0
			detailTmp := "设备掉线："
			remark = "收割机掉线，请及时登陆设备处理"
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
			detail = harvestersRpcResult.Error
			remark = "获取收割机列表失败"
			wechat.SendChiaMonitorNoticeToWechat(machineName, event, detail, remark)
		}

		time.Sleep(time.Duration(cfg.Monitor.FarmerInterval) * time.Minute)
	}

}

// GetRewardTargets 获取奖励地址
func (f Farmer) GetRewardTargets() {
	url := f.BaseUrl + "get_reward_targets"
	searchForPrivateKey := &SearchForPrivateKey{SearchForPrivateKey: f.IsSearchForPrivateKey}
	//发起请求
	resp, err := utils.PostHttps(url, searchForPrivateKey, "application/json", f.CertPath, f.KeyPath)
	if err != nil {
		log.Error(err)
		return
	}
	log.Info(string(resp))
}

// GetPoolState 获取矿池状态
func (f Farmer) GetPoolState() (poolStateRpcResult PoolStateRpcResult, err error) {
	url := f.BaseUrl + "get_pool_state"
	searchForPrivateKey := &SearchForPrivateKey{SearchForPrivateKey: f.IsSearchForPrivateKey}
	//发起请求
	resp, err := utils.PostHttps(url, searchForPrivateKey, "application/json", f.CertPath, f.KeyPath)
	if err != nil {
		log.Error(err)
		return
	}
	log.Debug(string(resp))

	err = json.Unmarshal(resp, &poolStateRpcResult)

	return poolStateRpcResult, err
}

// GetHarvesters 获取收割机状态
func (f Farmer) GetHarvesters() (harvestersRpcResult HarvestersRpcResult, err error) {
	url := f.BaseUrl + "get_harvesters"
	searchForPrivateKey := &SearchForPrivateKey{SearchForPrivateKey: f.IsSearchForPrivateKey}
	//发起请求
	resp, err := utils.PostHttps(url, searchForPrivateKey, "application/json", f.CertPath, f.KeyPath)
	if err != nil {
		log.Error(err)
		return
	}
	//log.Debug(string(resp))

	err = json.Unmarshal(resp, &harvestersRpcResult)

	return harvestersRpcResult, err
}

//MonitorPool 监控矿池状态
func MonitorPool(farmer Farmer) {
	var event string
	var detail string
	var remark string

	//获取配置文件
	cfg := config.GetConfig()
	machineName := cfg.Monitor.MachineName
	event = "监控矿池状态"

	//创建定时任务
	c := cron.New()
	err := c.AddFunc(cfg.Monitor.DailyCron, func() {
		//获取矿池状态
		poolStateRpcResult, err := farmer.GetPoolState()
		if err != nil {
			log.Error("Get pool state err: ", err)
			//发送获取rpc失败微信通知
			detail = err.Error()
			remark = "矿池状态获取错误"
		} else {
			//获取成功
			if poolStateRpcResult.Success {
				log.Info("Get pool state rpc result success!")
				//发送矿池状态微信通知
				successPercent := float64(len(poolStateRpcResult.PoolState[0].PointsAcknowledged24H)) / float64(len(poolStateRpcResult.PoolState[0].PointsFound24H)) * 100
				detail = fmt.Sprintf("%s，当前难度：%d，当前积分：%d，24h积分获取成功率：%.2f%%",
					poolStateRpcResult.PoolState[0].PoolConfig.PoolURL,
					poolStateRpcResult.PoolState[0].CurrentDifficulty,
					poolStateRpcResult.PoolState[0].CurrentPoints,
					successPercent,
				)
				remark = "获取矿池状态成功"
			} else {
				log.Info("Get pool state rpc result failed: ", poolStateRpcResult.Error)
				//发送获取钱包余额微信通知
				detail = poolStateRpcResult.Error
				remark = "获取矿池状态失败"
			}
		}
		wechat.SendChiaMonitorNoticeToWechat(machineName, event, detail, remark)
	})

	if err != nil {
		log.Fatal("Start pool monitor cron task err: ", err)
		return
	}

	c.Start()
	log.Info("Start pool monitor cron task success!")
}
