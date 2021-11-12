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
	MessageTitle  string `json:"message_title"`
	MessageDetail string `json:"message_detail"`
	UpdateTime    string `json:"update_time"`
	Remark        string `json:"remark"`
	WechatAccount string `json:"wechat_account"`
}

// SendChiaMonitorNoticeToWechat 发送Chia监控消息给微信
func SendChiaMonitorNoticeToWechat(messageTitle, messageDetail, remark string) {
	chiaMonitorMessage := ChiaMonitorMessage{
		MessageTitle:  messageTitle,
		MessageDetail: messageDetail,
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
