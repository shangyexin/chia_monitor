package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"chia_monitor/src/config"
)

//初始化日志设置
func InitLog(LogDir string, appName string, isProduction bool) {
	//获取配置文件
	cfg := config.GetConfig()

	baseLogPath := path.Join(LogDir, appName)
	/*
		日志轮转相关函数
		WithLinkName为最新的日志建立软连接
		WithRotationTime` 设置日志分割的时间，隔多久分割一次
		WithMaxAge 和 WithRotationCount二者只能设置一个
		WithMaxAge 设置文件清理前的最长保存时间
		WithRotationCount 设置文件清理前最多保存的个数
	*/
	//下面配置日志每隔24小时轮转一个新文件，保留最近30天的日志文件，多余的自动清理掉。
	file_writer, err := rotatelogs.New(
		baseLogPath+".%Y%m%d.log",
		rotatelogs.WithLinkName(baseLogPath+".log"),
		rotatelogs.WithMaxAge(time.Duration(cfg.LogConfig.LogSaveDay)*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		log.Errorf("Config local file system middleware error. %v", errors.WithStack(err))
	}

	log.SetReportCaller(true) //打印文件和行号

	log.SetOutput(io.MultiWriter(file_writer, os.Stdout)) //同时输出到文件和屏幕

	//设置日志格式和级别
	if isProduction {
		//logger.SetFormatter(&logger.JSONFormatter{})
		log.SetFormatter(new(LogFormatter))
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetFormatter(new(LogFormatter))
		log.SetLevel(log.DebugLevel)
	}

	log.Info("=====================Log init success=====================")
}

//日志自定义格式
type LogFormatter struct{}

//格式详情
func (s *LogFormatter) Format(entry *log.Entry) ([]byte, error) {
	timestamp := time.Now().Local().Format("2006-01-02 15:04:05")
	var file string
	var lenth int
	if entry.Caller != nil {
		file = filepath.Base(entry.Caller.File)
		lenth = entry.Caller.Line
	}
	//fmt.Println(entry.Data)
	msg := fmt.Sprintf("%s [%s:%d][GOID:%d][%s] %s\n",
		timestamp, file, lenth, getGID(), strings.ToUpper(entry.Level.String()), entry.Message)
	return []byte(msg), nil
}

//获取goroutine id
func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}