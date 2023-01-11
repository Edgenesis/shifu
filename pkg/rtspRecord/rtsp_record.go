package rtspRecord

import (
	"fmt"
	"github.com/edgenesis/shifu/pkg/logger"
	"net/http"
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
		DeviceName: request.DeviceName,
		In:         fmt.Sprintf("rtsp://%v:%v@%v", username, password, request.ServerAddress),
		Running:    false,
		Clip:       0,
	}
	if request.Record {
		d.startRecord()
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
		logger.Errorf("device %v not found", request.DeviceName)
		http.Error(w, "device not found", http.StatusBadRequest)
		return
	}
	err = d.stopRecord()
	if err != nil {
		logger.Errorf("when stop device %v error: %v", d.DeviceName, err)
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
		logger.Errorf("device %v not found", request.DeviceName)
		http.Error(w, "device not found", http.StatusBadRequest)
		return
	}
	if request.Record {
		if d.Running {
			logger.Warnf("try to start a already started device %v", request.DeviceName)
			return
		}
		d.startRecord()
	} else {
		if !d.Running {
			logger.Warnf("try to stop a already stopped device %v", request.DeviceName)
			return
		}
		err = d.stopRecord()
		if err != nil {
			logger.Errorf("when stop device %v error: %v", d.DeviceName, err)
		}
	}
	err = store.save()
	if err != nil {
		logger.Errorf("can't save map: %v", err)
		return
	}
}
