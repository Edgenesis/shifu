package main

import (
	"github.com/edgenesis/shifu/pkg/rtspRecord"
	"net/http"
	"os"
)

var serverListenPort = os.Getenv("SERVER_LISTEN_PORT")

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/register", rtspRecord.Register)
	mux.HandleFunc("/unregister", rtspRecord.Unregister)
	mux.HandleFunc("/update", rtspRecord.Update)
	err := http.ListenAndServe(serverListenPort, mux)
	if err != nil {
		return
	}
}
