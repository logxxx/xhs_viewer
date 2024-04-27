package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/logxxx/utils"
	"github.com/logxxx/utils/logger"
	"github.com/logxxx/utils/reqresp"
	log "github.com/sirupsen/logrus"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
	"os"
	"strings"
)

var (
	flagFromDirs = flag.String("from_dirs", "", "")
	toDir        = flag.String("to_dir", "", "")
	maxCount     = flag.Int("max_count", 1000, "")
)

type GetVideosResp struct {
	Total     int                 `json:"total"`
	Videos    []GetVideosRespElem `json:"videos,omitempty"`
	NextToken string              `json:"next_token,omitempty"`
	Time      string              `json:"time"`
}

type GetVideosRespElem struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Size string `json:"size,omitempty"`
}

func main() {

	log.SetLevel(log.DebugLevel)
	log.SetOutput(os.Stdout)

	log.SetFormatter(&logger.MyLogFormatter{})
	logger := &lumberjack.Logger{
		Filename:   "cli.log",
		MaxSize:    50, // 日志文件大小，单位是 MB
		MaxBackups: 3,  // 最大过期日志保留个数
		MaxAge:     30, // 保留过期文件最大时间，单位 天
	}

	log.SetOutput(logger) // logrus 设置日志的输出方式

	wd, _ := os.Getwd()
	log.Infof("wd:%v", wd)
	flag.Parse()

	if *flagFromDirs == "" {
		panic("empty from dir")
	}

	fromDirs := strings.Split(*flagFromDirs, ",")
	if len(fromDirs) == 0 {
		panic("empty from dirs")
	}

	if *toDir == "" {
		panic("empty to dir")
	}

	mgr := NewVideoMgr(fromDirs, *toDir, *maxCount)

	g := gin.New()
	g.Use(reqresp.Cors())

	initWeb(g, mgr)
	g.Run(":9887")
}

func getFilePathByID(id string) string {
	return utils.B64To(id)
}
