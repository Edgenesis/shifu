package rtspRecord

import (
	"encoding/json"
	"fmt"
	"github.com/edgenesis/shifu/pkg/logger"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"io"
	"net/http"
)

func Register(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Errorf("Error when Read Data From Body, error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	request := RegisterRequest{}
	err = json.Unmarshal(body, &request)
	if err != nil {
		logger.Errorf("Error to Unmarshal request body to struct")
		http.Error(w, "unexpected end of JSON input", http.StatusBadRequest)
		return
	}
	logger.Infof("request: %v", request)
	err = ffmpeg.Input(fmt.Sprintf("rtsp://%v:%v@%v", request.Username, request.Password, request.ServerAddress)).
		Output(".output2.mp4", ffmpeg.KwArgs{"c:v": "libx265"}).
		OverWriteOutput().ErrorToStdOut().Run()
}

func Unregister(w http.ResponseWriter, r *http.Request) {

}

func Update(w http.ResponseWriter, r *http.Request) {

}
