package main

import (
	"github.com/logxxx/utils"
	"github.com/logxxx/utils/fileutil"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

type RemoveBinlog struct {
	Row []RemoveRow
}

type RemoveRow struct {
	Time  string
	From  string
	To    string
	Error string
}

func AddToErrBinlog(from, to string, err error) {
	logs := &RemoveBinlog{}
	fileutil.ReadJsonFile("remove_binlog.json", logs)
	newRow := RemoveRow{
		Time:  utils.FormatTimeSafe(time.Now()),
		From:  from,
		To:    to,
		Error: err.Error(),
	}
	logs.Row = append(logs.Row, newRow)
	log.Debugf("AddToErrBinlog(%v) add:%+v", len(logs.Row), newRow)
	fileutil.WriteJsonToFile(logs, "remove_binlog.json")
}

func RunBinlog() {
	logs := &RemoveBinlog{}
	fileutil.ReadJsonFile("remove_binlog.json", logs)
	resp := &RemoveBinlog{}
	for _, row := range logs.Row {

		log.Infof("binlog.delete start. %v=>%v", row.From, row.To)

		if row.Error != "" && !strings.Contains(row.Error, "used by") {
			log.Infof("binlog.delete skip:%v", row.Error)
			continue
		}
		err := fileutil.MoveFileToDir(row.From, row.To)
		if err == nil {
			log.Infof("binlog.delete succ!")
			continue
		}
		if !strings.Contains(err.Error(), "used by") {
			log.Infof("binlog.delete err:%v", err)
			continue
		}
		log.Infof("binlog.delete failed:%v", row)
		resp.Row = append(resp.Row, row)
	}
	fileutil.WriteJsonToFile(resp, "remove_binlog.json")
}
