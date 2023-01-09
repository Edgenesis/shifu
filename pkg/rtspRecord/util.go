package rtspRecord

import (
	"encoding/json"
	"github.com/edgenesis/shifu/pkg/logger"
	"io"
	"net/http"
	"syscall"
)

func trans[T Request](r *http.Request) (*T, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	request := new(T)
	err = json.Unmarshal(body, request)
	if err != nil {
		return nil, err
	}
	logger.Infof("request: %v", *request)
	return request, nil
}

func startRecord(d *Device) {
	d.running = true
	go func() {
		err := d.cmd.Run()
		if err != nil {
			logger.Error(err)
			return
		}
	}()
}

func stopRecord(d *Device) error {
	d.running = false
	return d.cmd.Process.Signal(syscall.SIGINT)
}
