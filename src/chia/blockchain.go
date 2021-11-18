package chia

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"chia_monitor/src/config"
	"chia_monitor/src/utils"
	"chia_monitor/src/wechat"
)

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
	Success bool `json:"success"`
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
