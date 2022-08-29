package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/edgenesis/shifu/pkg/deviceshifu/mockdevice/mockdevice"
)

func main() {
	available_funcs := []string{
		"get_position",
		"get_status",
	}
	mockdevice.StartMockDevice(available_funcs, instructionHandler)
}

func instructionHandler(functionName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Handling: %v", functionName)
		switch functionName {
		case "get_position":
			rand.Seed(time.Now().UnixNano())
			xrange := 100
			yrange := 200
			xpos := strconv.Itoa(rand.Intn(xrange))
			ypos := strconv.Itoa(rand.Intn(yrange))
			fmt.Fprintf(w, "xpos: %v, ypos: %v", xpos, ypos)
		case "get_status":
			rand.Seed(time.Now().UnixNano())
			fmt.Fprintf(w, mockdevice.STATUS_STR_LIST[(rand.Intn(len(mockdevice.STATUS_STR_LIST)))])
		}
	}
}
