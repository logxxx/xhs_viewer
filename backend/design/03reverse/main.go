package main

import (
	"github.com/logxxx/utils/fileutil"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

func main() {
	count := 0
	fileutil.ScanFiles("G:\\weibo_download\\date", true, func(filePath string, fileInfo os.FileInfo) error {
		if !strings.HasPrefix(fileInfo.Name(), "1") {
			return nil
		}
		count++
		if count > 1000 {
			os.Exit(1)
		}
		log.Printf("%v %v", count, filePath)
		return nil
	})
}
