package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/logxxx/utils"
	"github.com/logxxx/utils/fileutil"
	"github.com/logxxx/utils/reqresp"
	"github.com/logxxx/utils/runutil"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func initWeb(g *gin.Engine, mgr *VideoMgr) {
	g.GET("/ping", func(c *gin.Context) {
		c.String(200, fmt.Sprintf("pong %v", utils.FormatTimeSafe(time.Now())))
	})

	g.GET("/viewer/reload_video", func(c *gin.Context) {
		log.Infof("reload video!")
		mgr.PreloadVideos()
		reqresp.MakeRespOk(c)
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

		filePath := utils.B64To(reqID)

		if reqAction != "" {
			dstDir := *toDir
			err := fileutil.MoveFileToDir(filePath, filepath.Join(dstDir, reqAction))
			log.Infof("moveto:%v=>%v err:%v", filePath, filepath.Join(dstDir, reqAction), err)
			if err != nil {
				runutil.GoRunSafe(func() {
					time.Sleep(10 * time.Second)
					err := fileutil.MoveFileToDir(filePath, filepath.Join(dstDir, reqAction))
					log.Infof("try move TWICE:%v=>%v err:%v", utils.B64To(reqID), filepath.Join(dstDir, reqAction), err)
					if err != nil && strings.Contains(err.Error(), "used") {
						AddToErrBinlog(filePath, filepath.Join(dstDir, reqAction), err)
					} else {
						log.Infof("no need add to binlog")
					}
				})
			}
		}

		mgr.RemoveVideo(filePath)
		os.Remove(filePath + ".thumb.mp4")

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
		var total = 0
		for i := 0; i < 1000; i++ {
			roundVs := []string{}
			total, roundVs, nextToken, err = mgr.GetVideos(limit, reqToken)
			if err != nil {
				reqresp.MakeErrMsg(c, err)
				return
			}
			for _, roundV := range roundVs {
				if !utils.HasFile(roundV) {
					continue
				}
				vs = append(vs, roundV)
			}

			if nextToken != "" && len(vs) >= limit {
				break
			}
			reqToken = nextToken
			limit -= len(vs)
		}

		resp := GetVideosResp{
			NextToken: nextToken,
			Total:     total,
		}
		for _, v := range vs {
			f, _ := os.Stat(v)
			if f == nil {
				continue
			}
			resp.Videos = append(resp.Videos, GetVideosRespElem{
				ID:   utils.B64(v),
				Name: filepath.Base(strings.ReplaceAll(v, "#", "%23")),
				Size: utils.GetShowSize(f.Size()),
			})
		}
		reqresp.MakeResp(c, resp)
	})

	g.GET("/viewer/file", func(c *gin.Context) {

		//X-Forwarded-For 和 X-Real-IP
		//log.Infof("video X-Forwarded-For:%v X-Real-IP:%v", c.GetHeader("X-Forwarded-For"), c.GetHeader("X-Real-IP"))

		id := c.Query("id")
		//log.Infof("get file:%v", id)
		isPreview := c.Query("is_preview")

		if id == "" {
			reqresp.MakeErrMsg(c, errors.New("empty id"))
			return
		}

		filePath := getFilePathByID(id)

		if isPreview == "true" && utils.HasFile(filePath+".thumb.mp4") {
			log.Infof("return thumb video:%v", filePath+".thumb.mp4")
			c.File(filePath + ".thumb.mp4")
		} else {
			log.Infof("return real video:%v", filePath)
			c.File(filePath)
		}

	})
	g.StaticFile("/", `D:\mytest\mywork\xhs_viewer\frontend\dist\index.html`)
	g.StaticFS("/dist", gin.Dir(`D:\mytest\mywork\xhs_viewer\frontend\dist`, true))
}
