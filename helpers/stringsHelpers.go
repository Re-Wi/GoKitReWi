package helpers

import (
	"net/url"
	"regexp"
	"strings"
)

func CutPrefix(s, prefix string) (string, bool) {
	if strings.HasPrefix(s, prefix) {
		return strings.TrimPrefix(s, prefix), true
	}
	return s, false
}

func EitherCutPrefix(s string, prefix ...string) (string, bool) {
	// 任一前缀匹配则返回剩余部分
	for _, p := range prefix {
		if strings.HasPrefix(s, p) {
			return strings.TrimPrefix(s, p), true
		}
	}
	return s, false
}

// trim space and equal
func TrimEqual(s, prefix string) (string, bool) {
	if strings.TrimSpace(s) == prefix {
		return "", true
	}
	return s, false
}

func EitherTrimEqual(s string, prefix ...string) (string, bool) {
	// 任一前缀匹配则返回剩余部分
	for _, p := range prefix {
		if strings.TrimSpace(s) == p {
			return "", true
		}
	}
	return s, false
}

func GetLarkbitableFromURL(url_str string) (string, string, string) {
	r := regexp.MustCompile(`base/(\w+)\?table=(\w+)&view=(\w+)`)
	matches := r.FindStringSubmatch(url_str)
	if len(matches) < 4 {
		return "", "", ""
	}
	return matches[1], matches[2], matches[3]
}

func CheckURL(u string) bool {
	parsedURL, err := url.ParseRequestURI(u)
	if err != nil {
		return false
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}
	return true
}
