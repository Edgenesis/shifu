package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/edgenesis/shifu/pkg/deviceshifu/mockdevice/mockdevice"
	"k8s.io/klog/v2"
)

func main() {
	availableFuncs := []string{
		"get_measurement",
		"get_status",
	}
	mockdevice.StartMockDevice(availableFuncs, instructionHandler)
}

func instructionHandler(functionName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		klog.Infof("Handling: %v", functionName)
		switch functionName {
		case "get_measurement":
			rand.Seed(time.Now().UnixNano())
			readingRange := float32(3.0)
			for i := 0; i < 8; i++ {
				for j := 0; j < 12; j++ {
					num := fmt.Sprintf("%.2f", rand.Float32()*readingRange)
					fmt.Fprintf(w, num+" ")
				}
				fmt.Fprintf(w, "\n")
			}
		case "get_status":
			rand.Seed(time.Now().UnixNano())
			fmt.Fprintf(w, mockdevice.StatusSetList[(rand.Intn(len(mockdevice.StatusSetList)))])
		}
	}
}
