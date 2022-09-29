package utils

import (
	"fmt"
	"log"
	"strings"
)

func ParseHTTPGetParams(urlStr string) (map[string]string, error) {
	var paramStr string
	log.Println(urlStr)
	url := strings.Split(urlStr, "?")

	if len(url) <= 0 {
		return nil, fmt.Errorf("empty Query")
	} else if len(url) == 1 {
		paramStr = url[0]
	} else {
		paramStr = url[1]
	}

	params := strings.Split(paramStr, "&")

	result := make(map[string]string, len(params))

	for _, item := range params {
		info := strings.Split(item, "=")
		if len(info) == 2 {
			result[info[0]] = info[1]
		} else if len(info) == 1 {
			result[info[0]] = ""
		}
	}

	return result, nil
}
