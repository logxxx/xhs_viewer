package main

import "net/http"

type ParseResult struct {
	DeviceID string
}

func ParsePath(reqURL *http.Request) ParseResult {
	resp := ParseResult{}

	return resp
}
