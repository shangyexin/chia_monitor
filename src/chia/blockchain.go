package chia

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"chia_monitor/src/config"
	"chia_monitor/src/utils"
	"chia_monitor/src/wechat"
)

const restartChiaCmd = "/root/restart.sh"
const syncingCountMax = 6

type BlockChain struct {
	BaseUrl  string
	CertPath string
	KeyPath  string
	WalletId int
}

type HeaderHash struct {
	HeaderHash string `json:"header_hash"`
}

type BlockchainStateRpcResult struct {
	BlockchainState struct {
		Difficulty                  int  `json:"difficulty"`
		GenesisChallengeInitialized bool `json:"genesis_challenge_initialized"`
		MempoolSize                 int  `json:"mempool_size"`
		Peak                        struct {
			ChallengeBlockInfoHash string `json:"challenge_block_info_hash"`
			ChallengeVdfOutput     struct {
				Data string `json:"data"`
			} `json:"challenge_vdf_output"`
			Deficit                            int         `json:"deficit"`
			FarmerPuzzleHash                   string      `json:"farmer_puzzle_hash"`
			Fees                               int         `json:"fees"`
			FinishedChallengeSlotHashes        interface{} `json:"finished_challenge_slot_hashes"`
			FinishedInfusedChallengeSlotHashes interface{} `json:"finished_infused_challenge_slot_hashes"`
			FinishedRewardSlotHashes           interface{} `json:"finished_reward_slot_hashes"`
			HeaderHash                         string      `json:"header_hash"`
			Height                             int         `json:"height"`
			InfusedChallengeVdfOutput          struct {
				Data string `json:"data"`
			} `json:"infused_challenge_vdf_output"`
			Overflow                   bool   `json:"overflow"`
			PoolPuzzleHash             string `json:"pool_puzzle_hash"`
			PrevHash                   string `json:"prev_hash"`
			PrevTransactionBlockHash   string `json:"prev_transaction_block_hash"`
			PrevTransactionBlockHeight int    `json:"prev_transaction_block_height"`
			RequiredIters              int    `json:"required_iters"`
			RewardClaimsIncorporated   []struct {
				Amount         int64  `json:"amount"`
				ParentCoinInfo string `json:"parent_coin_info"`
				PuzzleHash     string `json:"puzzle_hash"`
			} `json:"reward_claims_incorporated"`
			RewardInfusionNewChallenge string      `json:"reward_infusion_new_challenge"`
			SignagePointIndex          int         `json:"signage_point_index"`
			SubEpochSummaryIncluded    interface{} `json:"sub_epoch_summary_included"`
			SubSlotIters               int         `json:"sub_slot_iters"`
			Timestamp                  int         `json:"timestamp"`
			TotalIters                 int64       `json:"total_iters"`
			Weight                     int         `json:"weight"`
		} `json:"peak"`
		Space        interface{} `json:"space"`
		SubSlotIters int         `json:"sub_slot_iters"`
		Sync         struct {
			SyncMode           bool `json:"sync_mode"`
			SyncProgressHeight int  `json:"sync_progress_height"`
			SyncTipHeight      int  `json:"sync_tip_height"`
			Synced             bool `json:"synced"`
		} `json:"sync"`
	} `json:"blockchain_state"`
	Error   string `json:"error"`
	Success bool   `json:"success"`
}

type BlockRecordRpcResult struct {
	BlockRecord struct {
		ChallengeBlockInfoHash string `json:"challenge_block_info_hash"`
		ChallengeVdfOutput     struct {
			Data string `json:"data"`
		} `json:"challenge_vdf_output"`
		Deficit                            int           `json:"deficit"`
		FarmerPuzzleHash                   string        `json:"farmer_puzzle_hash"`
		Fees                               int           `json:"fees"`
		FinishedChallengeSlotHashes        []string      `json:"finished_challenge_slot_hashes"`
		FinishedInfusedChallengeSlotHashes interface{}   `json:"finished_infused_challenge_slot_hashes"`
		FinishedRewardSlotHashes           []string      `json:"finished_reward_slot_hashes"`
		HeaderHash                         string        `json:"header_hash"`
		Height                             int           `json:"height"`
		InfusedChallengeVdfOutput          interface{}   `json:"infused_challenge_vdf_output"`
		Overflow                           bool          `json:"overflow"`
		PoolPuzzleHash                     string        `json:"pool_puzzle_hash"`
		PrevHash                           string        `json:"prev_hash"`
		PrevTransactionBlockHash           string        `json:"prev_transaction_block_hash"`
		PrevTransactionBlockHeight         int           `json:"prev_transaction_block_height"`
		RequiredIters                      int           `json:"required_iters"`
		RewardClaimsIncorporated           []interface{} `json:"reward_claims_incorporated"`
		RewardInfusionNewChallenge         string        `json:"reward_infusion_new_challenge"`
		SignagePointIndex                  int           `json:"signage_point_index"`
		SubEpochSummaryIncluded            interface{}   `json:"sub_epoch_summary_included"`
		SubSlotIters                       int           `json:"sub_slot_iters"`
		Timestamp                          int           `json:"timestamp"`
		TotalIters                         int           `json:"total_iters"`
		Weight                             int           `json:"weight"`
	} `json:"block_record"`
	Error   string `json:"error"`
	Success bool   `json:"success"`
}

// GetBlockchainState 获取区块链状态
func (b BlockChain) GetBlockchainState() (blockchainStateRpcResult BlockchainStateRpcResult, err error) {
	url := b.BaseUrl + "get_blockchain_state"
	walletId := WalletId{WalletId: b.WalletId}
	//发起请求
	resp, err := utils.PostHttps(url, walletId, "application/json", b.CertPath, b.KeyPath)
	if err != nil {
		return
	}
	log.Debug(string(resp))

	err = json.Unmarshal(resp, &blockchainStateRpcResult)

	return blockchainStateRpcResult, err
}

// GetBlockRecord 获取区块记录
func (b BlockChain) GetBlockRecord(headerHashStr string) (blockRecordRpcResult BlockRecordRpcResult, err error) {
	url := b.BaseUrl + "get_block_record"
	headerHash := HeaderHash{HeaderHash: headerHashStr}
	//发起请求
	resp, err := utils.PostHttps(url, headerHash, "application/json", b.CertPath, b.KeyPath)
	if err != nil {
		return
	}
	log.Debug(string(resp))

	err = json.Unmarshal(resp, &blockRecordRpcResult)

	return blockRecordRpcResult, err
}

//MonitorBlockState 监控区块链状态
func MonitorBlockState(blockChain BlockChain) {
	var iSRestarted bool
	var isNeedAutoRecover bool
	var syncingCount int
	var event string
	var detail string
	var remark string
	var timestamp int
	var blockRecordRpcResult BlockRecordRpcResult

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
			log.Error("Get blockchain state rpc result failed: ", blockchainStateRpcResult.Error)
			//发送获取rpc失败微信通知
			event = "RPC获取区块链状态失败"
			detail = blockchainStateRpcResult.Error
			remark = "已自动重启Chia"
			wechat.SendChiaMonitorNoticeToWechat(machineName, event, detail, remark)
			//重启Chia
			restartChia()
			iSRestarted = true
		}

		time.Sleep(time.Duration(cfg.Monitor.BockChainInterval) * time.Minute)
	}
}

//TestNodeEvent 测试节点事件
func TestNodeEvent() {
	//获取配置文件
	cfg := config.GetConfig()

	machineName := cfg.Monitor.MachineName
	event := "node测试事件"
	detail := "node测试详情"
	remark := "node测试备注"
	//发送测试通知
	wechat.SendChiaMonitorNoticeToWechat(machineName, event, detail, remark)
}
