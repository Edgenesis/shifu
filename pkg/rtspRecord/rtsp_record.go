package rtspRecord

import (
	"encoding/json"
	"fmt"
	"github.com/edgenesis/shifu/pkg/logger"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"io"
	"net/http"
	"sync"
	"syscall"
)

var m sync.Map

func trans[T Request](r *http.Request) (T, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var request T
	err = json.Unmarshal(body, &request)
	if err != nil {
		return nil, err
	}
	logger.Infof("request: %v", request)
	return request, nil
}

func startRecord(name string) {
	di, exist := m.Load(name)
	if !exist {
		logger.Errorf("device %v not found.", name)
		return
	}
	d := di.(*Device)
	if d.running {
		logger.Warnf("try to start a already started device :%v", name)
		return
	}
	d.running = true
	go func() {
		err := d.cmd.Run()
		if err != nil {
			logger.Error(err)
			return
		}
	}()
}

func stopRecord(name string) {
	di, exist := m.Load(name)
	if !exist {
		logger.Error("device %v not found.")
		return
	}
	d := di.(*Device)
	if !d.running {
		logger.Warnf("try to stop a already stopped device :%v", name)
		return
	}
	d.running = false
	err := d.cmd.Process.Signal(syscall.SIGINT)
	if err != nil {
		logger.Errorf("Can't stop the process: ", err)
		return
	}
}

func Register(w http.ResponseWriter, r *http.Request) {
	request, err := trans[RegisterRequest](r)
	if err != nil {
		logger.Errorf("Error to Unmarshal request body to struct: %v", err)
		http.Error(w, "Error to Unmarshal request body", http.StatusBadRequest)
		return
	}
	cmd := ffmpeg.Input(fmt.Sprintf("rtsp://%v:%v@%v", request.Username, request.Password, request.ServerAddress),
		ffmpeg.KwArgs{"rtsp_transport": "tcp"}).
		Output(request.OutputPath, ffmpeg.KwArgs{"c": "copy"}).
		OverWriteOutput().ErrorToStdOut().Compile()
	m.Store(request.DeviceName, &Device{
		cmd:     cmd,
		running: false,
	})
	if request.Recoding {
		startRecord(request.DeviceName)
	}
}

func Unregister(w http.ResponseWriter, r *http.Request) {
	request, err := trans[UnregisterRequest](r)
	if err != nil {
		logger.Errorf("Error to Unmarshal request body to struct: %v", err)
		http.Error(w, "Error to Unmarshal request body", http.StatusBadRequest)
		return
	}
	stopRecord(request.DeviceName)
	m.Delete(request.DeviceName)
}

func Update(w http.ResponseWriter, r *http.Request) {
	request, err := trans[UpdateRequest](r)
	if err != nil {
		logger.Errorf("Error to Unmarshal request body to struct: %v", err)
		http.Error(w, "Error to Unmarshal request body", http.StatusBadRequest)
		return
	}
	if request.Record {
		startRecord(request.DeviceName)
	} else {
		stopRecord(request.DeviceName)
	}
}
