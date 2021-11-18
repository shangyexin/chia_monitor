package chia

import (
	"chia_monitor/src/utils"
	"encoding/json"
	log "github.com/sirupsen/logrus"
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
	Success bool `json:"success"`
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
func (f Farmer) GetPoolState() {
	url := f.BaseUrl + "get_pool_state"
	searchForPrivateKey := &SearchForPrivateKey{SearchForPrivateKey: f.IsSearchForPrivateKey}
	//发起请求
	resp, err := utils.PostHttps(url, searchForPrivateKey, "application/json", f.CertPath, f.KeyPath)
	if err != nil {
		log.Error(err)
		return
	}
	log.Info(string(resp))
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
