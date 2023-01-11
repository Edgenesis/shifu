package main

import (
	"github.com/edgenesis/shifu/pkg/logger"
	"github.com/edgenesis/shifu/pkg/rtspRecord"
	"net/http"
	"os"
)

var serverListenPort = os.Getenv("SERVER_LISTEN_PORT")

const storePersistFilePath = "/data/mapStore"
const videoPersistDirectory = "/data/video"

func main() {
	rtspRecord.InitPersistMap(storePersistFilePath)
	os.Mkdir(videoPersistDirectory, os.ModePerm)
	rtspRecord.VideoSavePath = videoPersistDirectory
	mux := http.NewServeMux()
	mux.HandleFunc("/register", rtspRecord.Register)
	mux.HandleFunc("/unregister", rtspRecord.Unregister)
	mux.HandleFunc("/update", rtspRecord.Update)
	err := http.ListenAndServe(serverListenPort, mux)
	logger.Infof("Listening at %#v", serverListenPort)
	if err != nil {
		logger.Error(err)
		return
	}
}
