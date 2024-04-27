package main

import (
	"github.com/logxxx/utils"
	"github.com/logxxx/utils/ffmpeg"
	"github.com/logxxx/utils/fileutil"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

var thumbDirs = []string{}

func main() {

	log.SetLevel(log.InfoLevel)

	rootDir := "H:/output_xhs"
	dirs, err := os.ReadDir(rootDir)
	if err != nil {
		panic(err)
	}
	for _, dir := range dirs {

		videoDir := filepath.Join(rootDir, dir.Name(), "视频")

		log.Infof("dir:%v", videoDir)

		fileutil.ScanFiles(videoDir, func(filePath string, fileInfo os.FileInfo) (err error) {
			if !utils.HasFile(filePath) {
				return
			}

			if !fileutil.IsVideo(filePath) {
				return
			}
			if fileInfo.Size() < 3*1024*1024 {
				return
			}

			if utils.HasFile(filePath + ".thumb.mp4") {
				return
			}
			if utils.GetFileSize(filePath+".thumb.mp4") > 0 {
				return
			}
			os.Remove(filePath + ".thumb.mp4")

			err1 := makeThumb(filePath)
			if err1 != nil {
				log.Errorf("make thumb err:%v filePath:%v", err1, filePath)
				os.Remove(filePath + ".thumb.mp4")
				return
			}

			f, err2 := os.Stat(filePath + ".thumb.mp4")
			if err2 != nil {
				log.Errorf("read thumb file err:%v", err)
				return
			}

			log.Infof("make thumbsize before:%v after:%v path:%v", utils.GetShowSize(fileInfo.Size()), utils.GetShowSize(f.Size()), filePath)

			return
		})

	}
}

func makeThumb(filePath string) error {
	return ffmpeg.GenePreviewVideo(filePath, filePath+".thumb.mp4")
}
