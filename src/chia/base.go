package chia

import (
	"os/exec"

	log "github.com/sirupsen/logrus"
)

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
