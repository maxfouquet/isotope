package main

import (
	"net/http"
)

var forwardableHeadersSet = map[string]bool{
	"X-Request-Id":      true,
	"X-B3-Traceid":      true,
	"X-B3-Spanid":       true,
	"X-B3-Parentspanid": true,
	"X-B3-Sampled":      true,
	"X-B3-Flags":        true,
	"X-Ot-Span-Context": true,
}

func extractForwardableHeader(
	header http.Header) (forwardableHeader http.Header) {
	for key := range forwardableHeadersSet {
		if values, ok := header[key]; ok {
			forwardableHeader[key] = values
		}
	}
	return
}
