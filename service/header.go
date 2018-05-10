package main

import (
	"net/http"
)

var forwardableHeadersSet = map[string]bool{
	"x-request-id":      true,
	"x-b3-traceid":      true,
	"x-b3-spanid":       true,
	"x-b3-parentspanid": true,
	"x-b3-sampled":      true,
	"x-b3-flags":        true,
	"x-ot-span-context": true,
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
