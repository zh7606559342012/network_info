package conf

import (
	"os"
	"path"
	"runtime"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

func LoggerInit() (err error) {
	logPath := CmnConf.LogConf.LogPath + "agent.log"
	//log.Formatter = &logrus.JSONFormatter{}                                       // 设置为json格式的日志
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644) // 创建一个log日志文件
	if err != nil {
		return
	}
	Log.Out = f                  // 设置log的默认文件输出
	gin.SetMode(gin.ReleaseMode) // 线上模式，控制台不会打印信息
	gin.DefaultWriter = Log.Out  // gin框架自己记录的日志也会输出

	Log.SetReportCaller(true)
	l, err := logrus.ParseLevel(CmnConf.LogConf.LogLevel)
	if err != nil {
		l = logrus.DebugLevel
	}
	Log.SetLevel(l)

	//时间戳
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	customFormatter.CallerPrettyfier = func(frame *runtime.Frame) (function string, file string) {
		fileName := path.Base(frame.File) + ":" + strconv.Itoa(frame.Line)
		//return frame.Function, fileName
		return "", fileName
	}

	Log.SetFormatter(customFormatter)

	return
}
