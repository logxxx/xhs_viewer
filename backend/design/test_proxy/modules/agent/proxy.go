package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
)

type PortProxy struct {
}

func (p *PortProxy) ServeHTTP(w http.ResponseWriter, httpReq *http.Request) {
	logger := log.WithField("func_name", "PortProxy.ServeHTTP").WithField("reqPath", httpReq.URL.Path)

	logger.Debugf("func start")

	pathElems := strings.Split(httpReq.URL.Path, "/")
	if len(pathElems) <= 2 {
		logger.Debugf("err1")
		write(w, 400, "错误的路径1")
		return
	}

	if pathElems[1] != "port_forward" {
		logger.Debugf("err2")
		write(w, 400, "错误的路径2")
		return
	}
	port, err := strconv.Atoi(pathElems[2])
	if err != nil {
		logger.Debugf("err3:%v", pathElems[2])
		write(w, 400, err.Error())
		return
	}

	uiURL, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%v", port))
	if err != nil {
		log.Errorf("RegisterReverseProxy url.Parse err:%v port:%v", err, port)
		return
	}
	logger.Debugf("uiURL:%v", uiURL)

	proxy := httputil.NewSingleHostReverseProxy(uiURL)

	httpReq.URL.Path = httpReq.URL.Path[len(fmt.Sprintf("/port_forward/%v", port)):]

	if httpReq.URL.Path == "" {
		httpReq.URL.Path = "/"
	}

	logger.Debugf("PortHandler new httpReq.URL:%v", httpReq.URL.String())

	proxy.ServeHTTP(w, httpReq)
}

func write(w http.ResponseWriter, code int, input string) {
	w.Write([]byte(input))
	w.WriteHeader(code)
}
