package utils

import (
	"log"
	"net/http"
	"time"
)

func RetryAndGetHTTP(url string, retries int) (*http.Response, error) {
	var (
		err      error
		response *http.Response
	)

	for retries > 0 {
		response, err = http.Get(url)
		if err != nil {
			log.Println(err)
			retries -= 1
			time.Sleep(time.Second * 1)
			continue
		}

		if response.StatusCode == http.StatusOK {
			return response, nil
		}
	}

	return nil, err
}
