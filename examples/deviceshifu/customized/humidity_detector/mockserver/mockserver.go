package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	var data []byte
	var customData []byte

	http.HandleFunc("/data/save", func(writer http.ResponseWriter, request *http.Request) {
		data, _ = io.ReadAll(request.Body)
		log.Println("save data from telemetry service", string(data))
	})
	http.HandleFunc("/data/read", func(writer http.ResponseWriter, request *http.Request) {
		_, _ = fmt.Fprint(writer, string(data))
		log.Println("read data")
	})
	http.HandleFunc("/custom_data/save", func(writer http.ResponseWriter, request *http.Request) {
		customData, _ = io.ReadAll(request.Body)
		log.Println("save customData from telemetry service", string(customData))
	})
	http.HandleFunc("/custom_data/read", func(writer http.ResponseWriter, request *http.Request) {
		_, _ = fmt.Fprint(writer, string(customData))
		log.Println("read customData")
	})

	_ = http.ListenAndServe(":11111", nil)
}
