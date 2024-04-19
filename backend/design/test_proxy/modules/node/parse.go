package main

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"path/filepath"
	"strings"
)

type ParseResult struct {
	DeviceID string
}

func ParsePath(req *http.Request) (resp ParseResult, err error) {

	// /drive-reverse-proxy-test/$device_id/port_forward/9887/
	//0/-----------1-----------/----2------/-----3-----/--4--/
	pathElems := strings.Split(req.URL.Path, "/")
	if len(pathElems) <= 4 {
		log.Debugf("err: len(pathElems) <= 3: %v", pathElems)
		return
	}

	if pathElems[1] != routePrefix {
		log.Debugf("pathElems[1] != routePrefix elem:%v prefix:%v", pathElems[1], routePrefix)
		return
	}

	log.Debugf("before parse path, req.URL.Path:%v", req.URL.Path)
	req.URL.Path = "/" + filepath.Join(pathElems[3:]...)
	log.Debugf("after parse path, req.URL.Path:%v", req.URL.Path)

	resp = ParseResult{
		DeviceID: pathElems[2],
	}

	return
}
