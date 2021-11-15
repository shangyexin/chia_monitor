package test

import (
	"chia_monitor/src/config"
	"chia_monitor/src/wechat"
)

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
