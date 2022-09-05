package utils

import (
	"log"
	"net/http"
	"time"
)

// RetryAndGetHTTP Send Http Get pre Second success or untill retries is reached
func RetryAndGetHTTP(url string, retries int) (*http.Response, error) {
	var (
		err      error
		response *http.Response
	)

	for retries > 0 {
		response, err = http.Get(url)
		if err != nil {
			log.Println(err)
			retries--
			time.Sleep(time.Second * 1)
			continue
		}

		if response.StatusCode == http.StatusOK {
			return response, nil
		}
	}

	return nil, err
}
