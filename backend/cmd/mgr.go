package main

import (
	"errors"
	"github.com/logxxx/utils"
	"github.com/logxxx/utils/fileutil"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
)

var (
	enoughVideoCount = errors.New("enough video count")
)

type VideoMgr struct {
	FromDirs []string
	ToDir    string
	Videos   []string
	MaxCount int
}

func NewVideoMgr(from []string, to string, maxCount int) *VideoMgr {

	if maxCount <= 0 {
		maxCount = 1000
	}
	if maxCount >= 10000 {
		maxCount = 1000
	}

	resp := &VideoMgr{
		FromDirs: from,
		ToDir:    to,
		MaxCount: maxCount,
	}

	RunBinlog()

	err := resp.PreloadVideos()
	if err != nil {
		panic(err)
	}
	return resp
}

func (m *VideoMgr) PreloadVideos() error {

	videos, err := findAllVideos(m.FromDirs, m.ToDir, m.MaxCount)
	if err != nil && err != enoughVideoCount {
		return err
	}
	if len(videos) == 0 {
		return errors.New("no video find")
	}
	log.Infof("VideoMgr.PreloadVideos get %v videos", len(videos))
	m.Videos = videos
	return nil
}

func (m *VideoMgr) RemoveVideo(path string) {
	for i := range m.Videos {
		if m.Videos[i] == path {
			m.Videos[i] = ""
			log.Infof("......remove video succ. idx=%v path=%v", i, path)
			return
		}
	}
	log.Infof("!!!!!!!!!!!!!!remove video fAILED!!! path=%v", path)
}

func (m *VideoMgr) GetVideos(limit int, tokenStr string) (total int, resp []string, nextToken string, err error) {

	defer func() {
		log.Infof("GetVideos limit=%v token=%v total=%v return[len=%v next=%v err=%v]", limit, tokenStr, len(m.Videos), len(resp), nextToken, err)
	}()

	if limit <= 0 {
		nextToken = tokenStr
		return
	}

	if len(m.Videos) == 0 {
		err = errors.New("no videos")
		return
	}

	total = len(m.Videos)

	lastIdx, _ := strconv.Atoi(tokenStr)

	if lastIdx >= len(m.Videos) {
		err = errors.New("no more videos")
		return
	}

	nextToken = strconv.Itoa(lastIdx + limit)

	if lastIdx+limit < len(m.Videos) {
		resp = m.Videos[lastIdx : lastIdx+limit]
	} else {
		resp = m.Videos[lastIdx:]
		nextToken = ""
	}

	log.Printf("here1 [%v~%v] resp(%v):%v", lastIdx, lastIdx+limit, len(resp), resp)

	resp = utils.RemoveEmpty(resp)
	//log.Printf("here2 resp(%v):%v", len(resp), resp)

	return
}

// 预加载视频
func findAllVideos(dirs []string, filterPath string, maxCount int) (videos []string, err error) {

	currCount := 0
	for _, dir := range dirs {
		err = fileutil.ScanFiles(dir, func(filePath string, fileInfo os.FileInfo) error {

			if currCount >= maxCount {
				return enoughVideoCount
			}

			if !fileutil.IsVideo(fileInfo.Name()) {
				return nil
			}
			if filePath == "" {
				return nil
			}
			if strings.HasPrefix(filePath, filterPath) {
				return nil
			}
			videos = append(videos, filePath)
			currCount++
			return nil
		})
		if err != nil {
			return
		}
	}

	return
}
