package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"edgenesis.io/shifu/deviceshifu/pkg/mockdevice/mockdevice"
)

func main() {
	available_funcs := []string{
		"read_value",
		"get_status",
	}
	mockdevice.StartMockDevice(available_funcs, instructionHandler)
}

func instructionHandler(functionName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Handling: %v", functionName)
		switch functionName {
		case "read_value":
			rand.Seed(time.Now().UnixNano())
			min := 10
			max := 30
			fmt.Fprint(w, strconv.Itoa(rand.Intn(max-min+1)+min))
		case "get_status":
			rand.Seed(time.Now().UnixNano())
			fmt.Fprint(w, mockdevice.STATUS_STR_LIST[(rand.Intn(len(mockdevice.STATUS_STR_LIST)))])
		}
	}
}
