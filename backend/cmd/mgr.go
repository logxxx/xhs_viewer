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

type VideoMgr struct {
	FromDir string
	ToDir   string
	Videos  []string
}

func NewVideoMgr(from, to string) *VideoMgr {

	if from == to {
		panic("from dir and to dir cannot totally same!")
	}

	resp := &VideoMgr{
		FromDir: from,
		ToDir:   to,
	}

	RunBinlog()

	err := resp.PreloadVideos()
	if err != nil {
		panic(err)
	}
	return resp
}

func (m *VideoMgr) PreloadVideos() error {
	videos, err := findAllVideos(m.FromDir, m.ToDir)
	if err != nil {
		return err
	}
	if len(videos) == 0 {
		return errors.New("no video find")
	}
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

func (m *VideoMgr) GetVideos(limit int, tokenStr string) (resp []string, nextToken string, err error) {

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
func findAllVideos(dir string, filterPath string) (videos []string, err error) {

	err = fileutil.ScanFiles(dir, func(filePath string, fileInfo os.FileInfo) error {

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
		return nil
	})
	if err != nil {
		return
	}
	return
}
