package main

import (
	"errors"
	"github.com/logxxx/utils"
	"github.com/logxxx/utils/fileutil"
	log "github.com/sirupsen/logrus"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

var (
	ErrEnoughCount = errors.New("enough count")

	videoMgr *VideoMgr

	imgMgr *ImageMgr
)

type VideoMgr struct {
	FromDirs []string
	ToDir    string
	Videos   []string
	MaxCount int
}

func InitMgr(vFromDirs []string, vToDir, iFromDir, iToDir string) {
	videoMgr = NewVideoMgr(vFromDirs, vToDir, 100)
	imgMgr = NewImageMgr(iFromDir, iToDir, 100)
}

func GetVideoMgr() *VideoMgr {
	return videoMgr
}

func GetImgMgr() *ImageMgr {
	return imgMgr
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
		log.Errorf("PreloadVideos err:%v", err)
	}
	return resp
}

func (m *VideoMgr) SetFromDir(fromDir string) {
	m.FromDirs = []string{fromDir}
}

func (m *VideoMgr) SetToDir(toDir string) {
	m.ToDir = toDir
}

func (m *VideoMgr) PreloadVideos() error {

	videos, err := findAllVideos(m.FromDirs, m.ToDir, m.MaxCount)
	if err != nil && err != ErrEnoughCount {
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
			//log.Infof("......remove video from mgr succ. idx=%v path=%v", i, path)
			return
		}
	}
	//log.Infof("!!!!!!!!!!!!!!remove video from mgr FAILED!!! path=%v", path)
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

	type VideoWithSize struct {
		Path string
		Size int64
	}

	vs := []VideoWithSize{}

	currCount := 0
	for _, dir := range dirs {
		err = fileutil.ScanFiles(dir, true, func(filePath string, fileInfo os.FileInfo) error {

			if strings.HasSuffix(filePath, "thumb.mp4") {
				return nil
			}

			if currCount >= maxCount {
				return ErrEnoughCount
			}

			//if fileInfo.Size() <= 100*1024*1024 {
			//	return nil
			//}

			if !fileutil.IsVideo(fileInfo.Name()) {
				return nil
			}
			if filePath == "" {
				return nil
			}
			if strings.HasPrefix(filePath, filterPath) {
				return nil
			}
			vs = append(vs, VideoWithSize{
				Path: filePath,
				Size: fileInfo.Size(),
			})
			currCount++
			log.Debugf("add video:%v", filePath)
			return nil
		})
		if err != nil && err != ErrEnoughCount {
			log.Errorf("findAllVideos ScanFiles err:%v", err)
			return
		}
	}

	//log.Infof("get %v videos", len(vs))

	sort.Slice(vs, func(i, j int) bool {
		return vs[i].Size > vs[j].Size
	})

	resp := []string{}
	for _, v := range vs {
		resp = append(resp, v.Path)
	}

	videos = resp

	return
}

func GetImageExt(input string) string {
	all := []string{".bmp", ".jpg", ".png", ".tif", ".gif", ".pcx", ".tga", ".exif", ".fpx", ".svg", ".psd", ".cdr", ".pcd", ".dxf", ".ufo", ".eps", ".ai", ".raw", ".WMF", ".webp", ".avif", ".apng"}
	for _, elem := range all {
		if strings.HasSuffix(input, elem) {
			return elem
		}
	}
	return ""
}

type ImageMgr struct {
	FromDir  string
	ToDir    string
	Images   [][]string
	MaxCount int
}

func NewImageMgr(from, to string, maxCount int) *ImageMgr {

	resp := &ImageMgr{
		FromDir:  from,
		ToDir:    to,
		MaxCount: 1000,
	}

	err := resp.PreloadImages()
	if err != nil {
		log.Errorf("NewImageMgr PreloadImages err:%v", err)
	}
	return resp

}

func (m *ImageMgr) PreloadImages() error {
	imgs, err := findAllImages(m.FromDir, m.ToDir, m.MaxCount)
	if err != nil && err != ErrEnoughCount {
		return err
	}
	if len(imgs) == 0 {
		return errors.New("no img find")
	}
	log.Infof("ImageMgr.PreloadImages get %v imgs", len(imgs))
	for _, elem := range imgs {
		m.Images = append(m.Images, elem)
	}

	return nil
}

// 预加载视频
func findAllImages(dir string, filterPath string, maxCount int) (images [][]string, err error) {

	modTime2Imgs := map[int64][]string{}

	count := 0
	err = fileutil.ScanFiles(dir, true, func(filePath string, fileInfo os.FileInfo) error {

		ext := GetImageExt(fileInfo.Name())
		if ext == "" {
			return nil
		}
		if filePath == "" {
			return nil
		}
		if filterPath != "" && strings.HasPrefix(filePath, filterPath) {
			return nil
		}

		count++
		if count >= maxCount {
			return ErrEnoughCount
		}

		appended := false
		for modTime, imgs := range modTime2Imgs {
			if math.Abs(float64(modTime-fileInfo.ModTime().Unix())) <= 3 {
				appended = true
				imgs = append(imgs, filePath)
				break
			}
		}
		if !appended {
			modTime2Imgs[fileInfo.ModTime().Unix()] = append(modTime2Imgs[fileInfo.ModTime().Unix()], filePath)
		}

		return nil
	})
	if err != nil {
		return
	}
	return
}

func (m *ImageMgr) GetImages(limit int, tokenStr string) (total int, resp [][]string, nextToken string, err error) {

	defer func() {
		log.Infof("GetImages limit=%v token=%v total=%v return[len=%v next=%v err=%v]", limit, tokenStr, len(m.Images), len(resp), nextToken, err)
	}()

	if limit <= 0 {
		nextToken = tokenStr
		return
	}

	if len(m.Images) == 0 {
		err = errors.New("no images")
		return
	}

	total = len(m.Images)

	lastIdx, _ := strconv.Atoi(tokenStr)

	if lastIdx >= len(m.Images) {
		err = errors.New("no more images")
		return
	}

	nextToken = strconv.Itoa(lastIdx + limit)

	if lastIdx+limit < len(m.Images) {
		resp = m.Images[lastIdx : lastIdx+limit]
	} else {
		resp = m.Images[lastIdx:]
		nextToken = ""
	}

	log.Printf("here1 [%v~%v] resp(%v):%v", lastIdx, lastIdx+limit, len(resp), resp)

	for i := range resp {
		resp[i] = utils.RemoveEmpty(resp[i])
	}

	newResp := [][]string{}
	for _, elem := range resp {
		if len(elem) == 0 {
			continue
		}
		newResp = append(newResp, elem)
	}

	resp = newResp

	//log.Printf("here2 resp(%v):%v", len(resp), resp)

	return
}
