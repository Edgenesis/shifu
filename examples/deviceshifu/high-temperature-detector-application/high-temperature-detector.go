package main

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"k8s.io/klog/v2"
)

func main() {
	targetURL := "http://edgedevice-thermometer/read_value"
	req, _ := http.NewRequest("GET", targetURL, nil)
	for {
		res, _ := http.DefaultClient.Do(req)
		body, _ := ioutil.ReadAll(res.Body)
		temperature, _ := strconv.Atoi(string(body))
		if temperature > 20 {
			klog.Infoln("High temperature:", temperature)
		} else if temperature > 15 {
			klog.Infoln("Normal temperature:", temperature)
		} else {
			klog.Infoln("Low temperature:", temperature)
		}
		res.Body.Close()
		time.Sleep(2 * time.Second)
	}
}
