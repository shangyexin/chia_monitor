package wechat

import (
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"chia_monitor/src/utils"
)

const chiaMonitorPostUrl = "https://test.wechat.yasin.store/api/v1/online/send_chia_monitor_message"
const myWechatAccount = "oOsegjnJh_Org9KilAs4CQ7pDjjE"

// ChiaMonitorMessage Chia监控消息结构体
type ChiaMonitorMessage struct {
	MachineName   string `json:"machine_name"`
	Event         string `json:"event"`
	Detail        string `json:"detail"`
	UpdateTime    string `json:"update_time"`
	Remark        string `json:"remark"`
	WechatAccount string `json:"wechat_account"`
}

// SendChiaMonitorNoticeToWechat 发送Chia监控消息给微信
func SendChiaMonitorNoticeToWechat(machineName, event, detail, remark string) {
	chiaMonitorMessage := ChiaMonitorMessage{
		MachineName:   machineName,
		Event:         event,
		Detail:        detail,
		UpdateTime:    time.Now().Format("2006-01-02 15:04:05"),
		Remark:        remark,
		WechatAccount: myWechatAccount,
	}
	log.Infof("chiaMonitorMessage: %+v", chiaMonitorMessage)
	resp, err := utils.Post(chiaMonitorPostUrl, chiaMonitorMessage, "application/json")
	result := strings.Trim(string(resp), "\"")
	if err != nil {
		log.Errorf("Send chia monitor notice failed: %+v", err)
	} else if result != "success" {
		log.Errorf("Send chia monitor notice failed: %s", result)
	} else {
		log.Info("Send chia monitor notice success")
	}
}
