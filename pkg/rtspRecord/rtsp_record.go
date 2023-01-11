package rtspRecord

import (
	"fmt"
	"github.com/edgenesis/shifu/pkg/logger"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"net/http"
	"path/filepath"
	"strconv"
)

var VideoSavePath string

func Register(w http.ResponseWriter, r *http.Request) {
	request, err := trans[RegisterRequest](r)
	if err != nil {
		logger.Errorf("Error to Unmarshal request body to struct: %v", err)
		http.Error(w, "Error to Unmarshal request body", http.StatusBadRequest)
		return
	}
	username, password, err := getCredential(request.SecretName)
	if err != nil {
		logger.Errorf("unable to get username and password, error: %v", err)
		http.Error(w, "unable to get username and password", http.StatusBadRequest)
		return
	}
	d := &Device{
		in:      fmt.Sprintf("rtsp://%v:%v@%v", username, password, request.ServerAddress),
		running: false,
		clip:    0,
	}
	out := filepath.Join(VideoSavePath, request.DeviceName+"_"+strconv.Itoa(d.clip)+".mp4")
	d.cmd = ffmpeg.Input(d.in, ffmpeg.KwArgs{"rtsp_transport": "tcp"}).
		Output(out, ffmpeg.KwArgs{"c": "copy"}).
		OverWriteOutput().ErrorToStdOut().Compile()
	if request.Recoding {
		startRecord(d)
		d.clip += 1
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	store.m[request.DeviceName] = d
	err = store.save()
	if err != nil {
		logger.Errorf("can't save map: %v", err)
		return
	}
}

func Unregister(w http.ResponseWriter, r *http.Request) {
	request, err := trans[UnregisterRequest](r)
	if err != nil {
		logger.Errorf("Error to Unmarshal request body to struct: %v", err)
		http.Error(w, "Error to Unmarshal request body", http.StatusBadRequest)
		return
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	d, exist := store.m[request.DeviceName]
	if !exist {
		logger.Error("device %v not found", request.DeviceName)
		http.Error(w, "device not found", http.StatusBadRequest)
		return
	}
	err = stopRecord(d)
	if err != nil {
		logger.Errorf("can't stop record of device %v: %v", request.DeviceName, err)
		http.Error(w, "can't stop record", http.StatusBadRequest)
		return
	}
	delete(store.m, request.DeviceName)
	err = store.save()
	if err != nil {
		logger.Errorf("can't save map: %v", err)
		return
	}
}

func Update(w http.ResponseWriter, r *http.Request) {
	request, err := trans[UpdateRequest](r)
	if err != nil {
		logger.Errorf("Error to Unmarshal request body to struct: %v", err)
		http.Error(w, "Error to Unmarshal request body", http.StatusBadRequest)
		return
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	d, exist := store.m[request.DeviceName]
	if !exist {
		logger.Error("device %v not found", request.DeviceName)
		http.Error(w, "device not found", http.StatusBadRequest)
		return
	}
	if request.Record {
		if d.running {
			logger.Warnf("try to start a already started device %v", request.DeviceName)
			return
		}
		out := filepath.Join(VideoSavePath, request.DeviceName+"_"+strconv.Itoa(d.clip)+".mp4")
		d.cmd = ffmpeg.Input(d.in, ffmpeg.KwArgs{"rtsp_transport": "tcp"}).
			Output(out, ffmpeg.KwArgs{"c": "copy"}).
			OverWriteOutput().ErrorToStdOut().Compile()
		startRecord(d)
		d.clip += 1
	} else {
		if !d.running {
			logger.Warnf("try to stop a already stopped device %v", request.DeviceName)
			return
		}
		err := stopRecord(d)
		if err != nil {
			logger.Errorf("can't stop record of device %v: %v", request.DeviceName, err)
			http.Error(w, "can't stop record", http.StatusBadRequest)
			return
		}
	}
	err = store.save()
	if err != nil {
		logger.Errorf("can't save map: %v", err)
		return
	}
}
