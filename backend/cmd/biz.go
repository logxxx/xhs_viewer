package main

import (
	"github.com/gin-gonic/gin"
	"github.com/logxxx/utils"
	"github.com/logxxx/utils/reqresp"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func GetImages(c *gin.Context) {
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
}
