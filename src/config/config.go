package config

import (
	"flag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var (
	cfgFile = flag.String("config", "./config.yaml", "配置文件路径")

	cfg *Config
)

// LogConfig 日志相关配置
type LogConfig struct {
	LogDir       string `yaml:"logDir"`       //日志文件夹
	AppName      string `yaml:"appName"`      //应用名称
	LogSaveDay   uint32 `yaml:"logSaveDay"`   //日志保存天数
	IsProduction bool   `yaml:"isProduction"` //是不是生产环境，生产环境日志级别：info
}

// FullNodeCertPath 全节点证书路径
type FullNodeCertPath struct {
	CertPath string `yaml:"certPath"`
	KeyPath  string `yaml:"keyPath"`
}

// WalletCertPath 钱包证书路径
type WalletCertPath struct {
	CertPath string `yaml:"certPath"`
	KeyPath  string `yaml:"keyPath"`
}

// Config 配置文件结构体
type Config struct {
	Listen            string `yaml:"listen"` //监听本地的端口
	*LogConfig        `yaml:"logConfig"`
	*FullNodeCertPath `yaml:"fullNodeCertPath"`
	*WalletCertPath   `yaml:"walletCertPath"`
}

//GetConfig 获取配置
func GetConfig() *Config {
	if cfg != nil {
		return cfg
	}
	bytes, err := ioutil.ReadFile(*cfgFile)
	if err != nil {
		panic(err)
	}

	cfgData := &Config{}
	err = yaml.Unmarshal(bytes, cfgData)
	if err != nil {
		panic(err)
	}
	cfg = cfgData
	return cfg
}