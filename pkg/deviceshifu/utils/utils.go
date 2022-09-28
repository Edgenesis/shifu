package utils

import (
	"net/url"
)

func ParseHTTPGetParams(urlStr string) (map[string]string, error) {
	urlInfo, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	values, err := url.ParseQuery(urlInfo.RawQuery)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(values))

	for key, value := range values {
		if len(value) == 0 {
			result[key] = ""
		} else {
			result[key] = value[0]
		}
	}

	return result, nil
}
