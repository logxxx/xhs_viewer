package main

import (
	"github.com/logxxx/utils"
	"github.com/logxxx/utils/ffmpeg"
	"github.com/logxxx/utils/fileutil"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

func main() {
	scanDirAndMakeThumb("G:/weibo_download")
}

func main1() {

	log.SetLevel(log.InfoLevel)

	rootDir := "G:\\weibo_download\\date"
	dirs, err := os.ReadDir(rootDir)
	if err != nil {
		panic(err)
	}

	for _, dir := range dirs {

		//videoDir := filepath.Join(rootDir, dir.Name(), "视频")
		videoDir := filepath.Join(rootDir, dir.Name())

		log.Infof("dir:%v", videoDir)

		scanDirAndMakeThumb(videoDir)

	}
}

func scanDirAndMakeThumb(dir string) {
	count := 0

	err := fileutil.ScanFiles(dir, true, func(filePath string, fileInfo os.FileInfo) (noErr error) {

		//log.Infof("scan %v", filePath)

		if !utils.HasFile(filePath) {
			return
		}

		if !fileutil.IsVideo(filePath) {
			return
		}

		/*
			newName, hit := fileutil.ReplaceInvalidChar(filePath, "RANDOM")
			if hit {
				count++
				log.Printf("RENAME%v: %v => %v", count, filePath, newName)
				err = os.Rename(filePath, newName)
				if err != nil {
					panic(err)
				}
			}

			return
		*/
		if fileInfo.Size() < 3*1024*1024 {
			return
		}

		if utils.GetFileSize(filePath+".thumb.mp4") > 0 {
			return
		}

		count++

		os.Remove(filePath + ".thumb.mp4")

		err1 := makeThumb(fileInfo, filePath)
		if err1 != nil {
			log.Errorf("make thumb err:%v filePath:%v", err1, filePath)
			os.Remove(filePath + ".thumb.mp4")
			return
		}

		f, err2 := os.Stat(filePath + ".thumb.mp4")
		if err2 != nil {
			log.Errorf("read thumb file err:%v", err2)
			return
		}

		log.Infof("make thumbsize(%v) before:%v after:%v path:%v", count, utils.GetShowSize(fileInfo.Size()), utils.GetShowSize(f.Size()), filePath)

		return
	})
	if err != nil {
		log.Errorf("ScanFiles err:%v", err)
	}
}

func makeThumb(fileInfo os.FileInfo, filePath string) error {
	segNum := 3
	segDur := 5
	skipStart := 0
	skipEnd := 0
	if fileInfo.Size() > 100*1024*1024 {
		skipStart = 60
		skipEnd = 60
		segNum = 5
		segDur = 5

	}
	log.Infof("make thumb %v", filePath)
	_, err := ffmpeg.GenePreviewVideoSlice(ffmpeg.GenePreviewVideoSliceOpt{
		FilePath:    filePath,
		ToPath:      filePath + ".thumb.mp4",
		SegNum:      segNum,
		SegDuration: segDur,
		SkipStart:   skipStart,
		SkipEnd:     skipEnd,
	})
	return err
}
