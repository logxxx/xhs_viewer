package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/logxxx/utils"
	"github.com/logxxx/utils/fileutil"
	"github.com/logxxx/utils/logger"
	"github.com/logxxx/utils/reqresp"
	"github.com/logxxx/utils/runutil"
	log "github.com/sirupsen/logrus"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	fromDir = flag.String("from_dir", "", "")
	toDir   = flag.String("to_dir", "", "")
)

type GetVideosResp struct {
	Videos    []GetVideosRespElem `json:"videos,omitempty"`
	NextToken string              `json:"next_token,omitempty"`
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

	if *fromDir == "" {
		panic("empty from dir")
	}

	if *toDir == "" {
		panic("empty to dir")
	}

	mgr := NewVideoMgr(*fromDir, *toDir)

	g := gin.New()
	g.Use(reqresp.Cors())

	g.GET("/ping", func(c *gin.Context) {
		c.String(200, fmt.Sprintf("pong %v", utils.FormatTimeSafe(time.Now())))
	})

	g.GET("/viewer/act", func(c *gin.Context) {
		reqID := c.Query("id")
		reqAction := c.Query("action")
		if reqID == "" || reqAction == "" {
			log.Infof("handle act err: empty req")
			reqresp.MakeErrMsg(c, errors.New("empty req"))
			return
		}
		log.Infof("handle act id=%v action=%v", reqID, reqAction)

		if reqAction != "" {
			dstDir := *toDir
			err := fileutil.MoveFileToDir(utils.B64To(reqID), filepath.Join(dstDir, reqAction))
			log.Infof("moveto:%v=>%v err:%v", utils.B64To(reqID), filepath.Join(dstDir, reqAction), err)
			if err != nil {
				runutil.GoRunSafe(func() {
					time.Sleep(10 * time.Second)
					err := fileutil.MoveFileToDir(utils.B64To(reqID), filepath.Join(dstDir, reqAction))
					log.Infof("try move TWICE:%v=>%v err:%v", utils.B64To(reqID), filepath.Join(dstDir, reqAction), err)
					if err != nil && strings.Contains(err.Error(), "used") {
						AddToErrBinlog(utils.B64To(reqID), filepath.Join(dstDir, reqAction), err)
					} else {
						log.Infof("no need add to binlog")
					}
				})
			}
		}

		mgr.RemoveVideo(utils.B64To(reqID))

		reqresp.MakeRespOk(c)
	})

	g.GET("/viewer/videos", func(c *gin.Context) {

		reqLimit := c.Query("limit")
		reqToken := c.Query("next_token")
		limit, _ := strconv.Atoi(reqLimit)
		if limit <= 0 {
			limit = 10
		}

		vs := []string{}
		var err error
		var nextToken = ""
		for i := 0; i < 1000; i++ {
			roundVs := []string{}
			roundVs, nextToken, err = mgr.GetVideos(limit, reqToken)
			if err != nil {
				reqresp.MakeErrMsg(c, err)
				return
			}
			vs = append(vs, roundVs...)
			if nextToken != "" && len(vs) >= limit {
				break
			}
			reqToken = nextToken
			limit -= len(vs)
		}

		resp := GetVideosResp{
			NextToken: nextToken,
		}
		for _, v := range vs {
			f, _ := os.Stat(v)
			resp.Videos = append(resp.Videos, GetVideosRespElem{
				ID:   utils.B64(v),
				Name: filepath.Base(v),
				Size: utils.GetShowSize(f.Size()),
			})
		}
		reqresp.MakeResp(c, resp)
	})

	fileLimiter := map[string]int{}
	g.GET("/viewer/file", func(c *gin.Context) {

		//X-Forwarded-For 和 X-Real-IP
		//log.Infof("video X-Forwarded-For:%v X-Real-IP:%v", c.GetHeader("X-Forwarded-For"), c.GetHeader("X-Real-IP"))

		id := c.Query("id")
		//log.Infof("get file:%v", id)

		if id == "" {
			reqresp.MakeErrMsg(c, errors.New("empty id"))
			return
		}

		filePath := getFilePathByID(id)

		if fileLimiter[filePath] > 20*1024*1024 {
			//c.String(200, "")
			//return
		}

		c.File(filePath)

		allTotal := 0
		for _, s := range fileLimiter {
			allTotal += s
		}

		log.Debugf("GetFile return:%v ALL:[count=%v total=%v] path:%v", utils.GetShowSize(int64(c.Writer.Size())), len(fileLimiter), utils.GetShowSize(int64(allTotal)), filePath)
		fileLimiter[filePath] += c.Writer.Size()
		return

		f, err := os.Open(filePath)
		if err != nil {
			reqresp.MakeErrMsg(c, err)
			return
		}
		to := make([]byte, 10*1024*1024)
		succ, _ := f.Read(to)
		if succ > 0 {
			to = to[:succ]
		}
		log.Debugf("GetFile return:%v path:%v", utils.GetShowSize(int64(len(to))), filePath)
		c.Writer.Write(to)

	})
	g.StaticFile("/", `D:\mytest\mywork\xhs_viewer\frontend\dist\index.html`)
	g.StaticFS("/dist", gin.Dir(`D:\mytest\mywork\xhs_viewer\frontend\dist`, true))
	g.Run(":9887")
}

func getFilePathByID(id string) string {
	return utils.B64To(id)
}
