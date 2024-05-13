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
	flagVideoFromDirs = flag.String("video_from_dirs", "", "")
	flagVideoToDir    = flag.String("video_to_dir", "", "")
	flagVideoMaxCount = flag.Int("video_max_count", 1000, "")

	flagImageFromDir  = flag.String("image_from_dir", "", "")
	flagImageToDir    = flag.String("image_to_dir", "", "")
	flagImageMaxCount = flag.Int("image_max_count", 1000, "")

	flagDistDir = flag.String("dist_dir", "", "")
)

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

	if !utils.HasFile(*flagDistDir) {
		panic("invalid distDir")
	}

	videoFromDirs := strings.Split(*flagVideoFromDirs, ",")

	InitMgr(videoFromDirs, *flagVideoToDir, *flagImageFromDir, *flagImageToDir)

	g := gin.New()
	g.Use(reqresp.Cors())

	initWeb(g)
	g.Run(":9887")
}

func getFilePathByID(id string) string {
	return utils.B64To(id)
}
