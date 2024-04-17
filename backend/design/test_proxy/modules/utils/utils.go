package utils

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"time"
)

var (
	pid = os.Getpid()
)

func Log(ctx context.Context, fnName string) *log.Entry {
	resp := log.WithField("pid", pid).WithField("func_name", fnName).WithField("func_st", time.Now().Format("01/02 15:04:05"))

	traceID := ""
	if ctx != nil {
		rawTraceID := ctx.Value("trace_id")
		if rawTraceID != nil {
			traceID = fmt.Sprintf("%v", rawTraceID)
		}

	}
	if traceID == "" {
		traceID = "_" + RandStr(7)

	}

	resp = resp.WithField("trace_id", traceID)
	return resp
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func RandStr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
