package util

import (
	"net/url"
	"strings"
)

func WriteString(strs ...string) string {
	var b strings.Builder
	for _, str := range strs {
		b.WriteString(str)
	}
	return b.String()
}

func SplitLines(s string) []string {
	if strings.Contains(s, "\r\n") {
		return strings.Split(s, "\r\n")
	}
	return strings.Split(s, "\n")
}

func PercentEncode(value string) string {
	if value == "" {
		return ""
	}
	encoded := url.QueryEscape(value)
	encoded = strings.ReplaceAll(encoded, "+", "%20")
	encoded = strings.ReplaceAll(encoded, "*", "%2A")
	encoded = strings.ReplaceAll(encoded, "%7E", "~")
	return encoded
}
