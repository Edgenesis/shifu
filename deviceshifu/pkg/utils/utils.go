package utils

import "strings"

//ParseAllParams get all params on GET
func ParseAllParams(url string) map[string]string {
	var result = make(map[string]string)

	paramBody := strings.Split(url, "?")
	if len(paramBody) < 2 {
		return nil
	}

	params := strings.Split(paramBody[1], "&")

	for _, item := range params {
		info := strings.Split(item, "=")
		if len(info) == 1 {
			result[info[0]] = ""
		} else {
			result[info[0]] = info[1]
		}
	}

	return result
}
