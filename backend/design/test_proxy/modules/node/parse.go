package main

import "net/http"

type ParseResult struct {
	DeviceID string
}

func ParsePath(reqURL *http.Request) ParseResult {
	resp := ParseResult{
		DeviceID: "device_id#123456",
	}

	return resp
}
