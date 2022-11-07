package unitest

import (
	"net/http"
	"time"

	"k8s.io/klog/v2"
)

// SO FAR ONLY FOR UNIT TESTING USAGE
// DO NOT USE IN NONE UNIT TESTING CODE
// if there are common use case, let's move it to an much common package or even another code repo

// RetryAndGetHTTP Send Http Get pre Second success or untill retries is reached
func RetryAndGetHTTP(url string, retries int) (*http.Response, error) {
	var (
		err      error
		response *http.Response
	)

	for retries > 0 {
		response, err = http.Get(url)
		if err != nil {
			klog.Errorf("%v", err)
			retries--
			time.Sleep(time.Millisecond * 100)
			continue
		}

		if response.StatusCode == http.StatusOK {
			return response, nil
		}
	}

	return nil, err
}
