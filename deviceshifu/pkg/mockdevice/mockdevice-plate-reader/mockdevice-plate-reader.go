package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"edgenesis.io/shifu/deviceshifu/pkg/mockdevice/mockdevice"
)

func main() {
	available_funcs := []string{
		"get_measurement",
		"get_status",
	}
	mockdevice.StartMockDevice(available_funcs, instructionHandler)
}

func instructionHandler(functionName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Handling: %v", functionName)
		switch functionName {
		case "get_measurement":
			matrix_row, matrix_col := 8, 12
			rand.Seed(time.Now().UnixNano())
			reading_range := float32(3.0)
			for i := 0; i < matrix_row; i++ {
				for j := 0; j < matrix_col; j++ {
					num := fmt.Sprintf("%.2f", rand.Float32()*reading_range)
					fmt.Fprintf(w, num+" ")
				}
				fmt.Fprintf(w, "\n")
			}
		case "get_status":
			rand.Seed(time.Now().UnixNano())
			fmt.Fprintf(w, mockdevice.STATUS_STR_LIST[(rand.Intn(len(mockdevice.STATUS_STR_LIST)))])
		}
	}
}
