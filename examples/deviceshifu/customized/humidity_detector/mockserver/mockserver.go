package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

var data []byte

func main() {
	http.HandleFunc("/save_data", saveData)
	http.HandleFunc("/read_data", readData)

	http.ListenAndServe(":11111", nil)
}

func saveData(w http.ResponseWriter, r *http.Request) {
	data, _ = io.ReadAll(r.Body)
	log.Println("save data from telemetry service")
}

func readData(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, string(data))
	log.Println("read data")
}
