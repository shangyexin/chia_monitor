package chia

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"

	"chia_monitor/src/config"
	"chia_monitor/src/utils"
	"chia_monitor/src/wechat"
)

const XCHPoolDailyEarningUrl = "https://farmer.xchpool.io/api/xchpool/farmer/earnings/daily"

type XCHPoolEarning struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    int    `json:"code"`
	Result  []struct {
		Date   string  `json:"date"`
		Amount float64 `json:"amount"`
	} `json:"result"`
	Timestamp int64 `json:"timestamp"`
}

// getXCHPoolEarning 获取XCHPool收益
func getXCHPoolEarning() (xchPoolEarning XCHPoolEarning, err error) {
	params := url.Values{}
	Url, err := url.Parse(XCHPoolDailyEarningUrl)
	if err != nil {
		log.Fatal(err)
		return
	}
	//获取配置文件
	cfg := config.GetConfig()

	params.Set("launcherId", cfg.Monitor.LauncherId)
	//如果参数中有中文参数,这个方法会进行URLEncode
	Url.RawQuery = params.Encode()
	urlPath := Url.String()
	log.Info("XCHPool earning url: ", urlPath)
	resp, err := utils.Get(urlPath)
	if err != nil {
		return
	}
	err = json.Unmarshal(resp, &xchPoolEarning)

	return xchPoolEarning, err
}

//MonitorPoolEarning 监控矿池收益
func MonitorPoolEarning(poolName string) {
	var event string
	var detail string
	var remark string
	//获取配置文件
	cfg := config.GetConfig()
	machineName := cfg.Monitor.MachineName

	event = "获取矿池收益"

	//创建定时任务
	c := cron.New()
	err := c.AddFunc(cfg.Monitor.DailyCron, func() {
		//获取矿池收益
		switch poolName {
		case "XCHPool":
			xchPoolEarning, err := getXCHPoolEarning()
			log.Infof("xchPoolEarning: %+v", xchPoolEarning)
			if err != nil {
				detail = fmt.Sprintf("获取XCHPool收益错误：%s", err.Error())
				remark = "获取矿池收益错误"
			} else if !xchPoolEarning.Success {
				detail = fmt.Sprintf("获取XCHPool收益失败：%s", xchPoolEarning.Message)
				remark = "获取矿池收益失败"
			} else {
				yesterdayEarning := xchPoolEarning.Result[len(xchPoolEarning.Result)-2].Amount
				todayEarning := xchPoolEarning.Result[len(xchPoolEarning.Result)-1].Amount
				detail = fmt.Sprintf("%s，昨日收益：%.5f，今日当前收益：%.5f", poolName, yesterdayEarning, todayEarning)
				remark = "获取矿池收益成功"
			}
		default:
			log.Error("Unknown pool name: ", poolName)
			detail = fmt.Sprintf("未知的矿池名称：%s", poolName)
			remark = "获取矿池收益错误"
		}
		//发送微信消息成功
		wechat.SendChiaMonitorNoticeToWechat(machineName, event, detail, remark)
	})

	if err != nil {
		log.Fatal("Start pool earning cron task err: ", err)
		return
	}

	c.Start()
	log.Info("Start pool earning cron task success!")
}
