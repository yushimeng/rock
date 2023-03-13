package sip

import (
	"strconv"
	"strings"
)

func parseContentLength(headerName string, headerText string) (
	header Header, err error) {
	var contentLength ContentLengthHeader
	var value uint64
	value, err = strconv.ParseUint(strings.TrimSpace(headerText), 10, 32)
	contentLength = ContentLengthHeader(value)
	return &contentLength, err
}

func parseContentType(headerName string, headerText string) (headers Header, err error) {
	// var contentType ContentType
	headerText = strings.TrimSpace(headerText)
	contentType := ContentTypeHeader(headerText)
	return &contentType, nil
}
