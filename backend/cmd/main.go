package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/logxxx/utils"
	"github.com/logxxx/utils/fileutil"
	"github.com/logxxx/utils/logger"
	"github.com/logxxx/utils/randutil"
	"github.com/logxxx/utils/reqresp"
	"github.com/logxxx/utils/runutil"
	log "github.com/sirupsen/logrus"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	fromDir            = flag.String("from_dir", "", "")
	toDir              = flag.String("to_dir", "", "")
	repeatContentLarge = strings.Repeat("hello world", 1000000)
	repeatContentSmall = strings.Repeat("hello world", 10000)
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

		var err error
		if reqAction != "" {
			dstDir := *toDir
			err = os.MkdirAll(filepath.Join(dstDir, reqAction), 0755)
			if err != nil {
				log.Errorf("MkdirAll err:%v dir:%v", err, filepath.Join(dstDir, reqAction))
			}
			err = fileutil.MoveFileToDir(utils.B64To(reqID), filepath.Join(dstDir, reqAction))
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

		if err != nil {
			if strings.Contains(err.Error(), "permission denied") {
				reqresp.MakeErrMsg(c, errors.New("permission denied"))
				return
			}
			reqresp.MakeErrMsg(c, err)
			return
		}

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
		var total int
		for i := 0; i < 1000; i++ {
			roundVs := []string{}
			total, roundVs, nextToken, err = mgr.GetVideos(limit, reqToken)
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
			Total:     total,
			NextToken: nextToken,
			Time:      utils.FormatTimeSafe(time.Now()),
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

	g.GET("/viewer/test_stream/:id", func(c *gin.Context) {
		streamIDStr := c.Param("id")
		streamID, _ := strconv.Atoi(streamIDStr)
		if streamID > 0 {
			//time.Sleep(time.Second * time.Duration(streamID%10))
		}
		content := repeatContentSmall
		if streamID%10 == 0 {
			//content = repeatContentLarge
		}
		c.String(200, fmt.Sprintf("%v: %v", streamIDStr, randutil.RandStr(10)+content+randutil.RandStr(10)))
	})
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

		if fileLimiterGet(filePath) > 20*1024*1024 {
			//c.String(200, "")
			//return
		}

		/*
			f, err := os.Open(filePath)
			if err != nil {
				reqresp.MakeErrMsg(c, err)
				return
			}
			io.Copy(c.Writer, f)

			return

		*/

		c.File(filePath)

		allTotal := fileLimiterTotal()
		//req: Range resp:Content-Range
		log.Debugf("GetFile return:%v req_range=%v resp_range=%v ALL:[count=%v total=%v] path:%v",
			utils.GetShowSize(int64(c.Writer.Size())),
			getReqRangeSize(c.Request.Header.Get("Range")),
			getRespRangeSize(c.Writer.Header().Get("Content-Range")),
			fileLimiterLen(), utils.GetShowSize(int64(allTotal)), filePath)
		fileLimiterAdd(filePath, c.Writer.Size())
		log.Debugf("reqRange:%v=>%v", c.Request.Header.Get("Range"), getReqRangeSize(c.Request.Header.Get("Range")))
		log.Debugf("respRange:%v=>%v", c.Writer.Header().Get("Content-Range"), getRespRangeSize(c.Writer.Header().Get("Content-Range")))
		return

		/*
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

		*/

	})
	//g.StaticFile("/", `../../frontend/dist/index.html`)
	g.StaticFile("/", `/data/hehanyang/mytest/xhs_viewer/frontend/dist/index.html`)
	//g.StaticFS("/dist", gin.Dir(`../../frontend/dist/`, true))
	g.StaticFS("/dist", gin.Dir(`/data/hehanyang/mytest/xhs_viewer/frontend/dist/`, true))
	g.Run(":9887")
}

var (
	fileLimiter     = map[string]int{}
	fileLimiterLock sync.Mutex
)

func fileLimiterTotal() (resp int) {
	for _, v := range fileLimiter {
		resp += v
	}
	return
}

func fileLimiterLen() int {
	fileLimiterLock.Lock()
	defer fileLimiterLock.Unlock()
	return len(fileLimiter)
}

func fileLimiterGet(path string) int {
	fileLimiterLock.Lock()
	defer fileLimiterLock.Unlock()
	return fileLimiter[path]
}

func fileLimiterAdd(path string, delta int) {
	fileLimiterLock.Lock()
	defer fileLimiterLock.Unlock()
	fileLimiter[path] += delta
}

func getReqRangeSize(input string) string {
	//bytes=xxxxxxx-
	sizeStr := utils.Extract(input, "bytes=", "-")
	size, _ := strconv.Atoi(sizeStr)
	return utils.GetShowSize(int64(size)) + "-"
}

func getRespRangeSize(input string) string {
	//bytes xxx-xxx/xxx
	range1Str := utils.Extract(input, "bytes ", "-")
	range2Str := utils.Extract(input, "/", "")
	range1, _ := strconv.Atoi(range1Str)
	range2, _ := strconv.Atoi(range2Str)
	return utils.GetShowSize(int64(range1)) + "-" + utils.GetShowSize(int64(range2))
}

func getFilePathByID(id string) string {
	return utils.B64To(id)
}
