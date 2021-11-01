package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"edgenesis.io/shifu/deviceshifu/pkg/mockdevice/mockdevice"
)

func main() {
	available_funcs := []string{
		"get_measurement",
		"get_status",
		"measure_plate",
		"get_output",
	}
	mockdevice.StartMockDevice(available_funcs, instructionHandler)
}

func instructionHandler(functionName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Handling: %v", functionName)
		switch functionName {
		case "get_measurement":
			output_matrix := [8][12]float32{}
			rand.Seed(time.Now().UnixNano())
			reading_range := float32(3.0)
			for i := 0; i < len(output_matrix); i++ {
				for j := 0; j < len(output_matrix[i]); j++ {
					num := fmt.Sprintf("%.2f", rand.Float32()*reading_range)
					fmt.Fprintf(w, num+" ")
				}
				fmt.Fprintf(w, "\n")
			}
		case "get_status":
			rand.Seed(time.Now().UnixNano())
			fmt.Fprintf(w, mockdevice.STATUS_STR_LIST[(rand.Intn(len(mockdevice.STATUS_STR_LIST)))])
		case "measure_plate":
			values := r.URL.Query()
			for parameterName, parameterValues := range values {
				log.Printf("paramname is: %v, value is: %v\n", parameterName, parameterValues[0])
			}

			plateType, ok := r.URL.Query()["plateType"]

			if !ok || len(plateType[0]) < 1 {
				log.Println("Url Param 'plateType' is missing")
				return
			}

			waveLength, ok := r.URL.Query()["waveLength"]

			if !ok || len(waveLength[0]) < 1 {
				log.Println("Url Param 'waveLength' is missing")
				return
			}

			rows, ok := r.URL.Query()["rows"]

			if !ok || len(rows[0]) < 1 {
				log.Println("Url Param 'rows' is missing")
				return
			}

			columns, ok := r.URL.Query()["columns"]

			if !ok || len(columns[0]) < 1 {
				log.Println("Url Param 'columns' is missing")
				return
			}

			log.Printf("Start scanning with plateType: %v, waveLength: %v, on %v rows and %v columns",
				string(plateType[0]), string(waveLength[0]), string(rows[0]), string(columns[0]))
			fmt.Fprintf(w, "Start scanning with plateType: %v, waveLength: %v, on %v rows and %v columns",
				string(plateType[0]), string(waveLength[0]), string(rows[0]), string(columns[0]))
		case "get_output":
			fileLocationString := "/root/Method 1_20211025_180406.xlsx"
			if _, err := os.Stat(fileLocationString); err == nil {
				fileBytes, err := ioutil.ReadFile(fileLocationString)
				if err != nil {
					panic(err)
				}
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/octet-stream")
				w.Write(fileBytes)
				return
			} else if errors.Is(err, os.ErrNotExist) {
				log.Printf("File does not exist: %v", fileLocationString)
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, "File does not exist: "+fileLocationString+"\n")
			} else {
				log.Printf("File may not exist: %v", fileLocationString)
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, "File may not exist: "+fileLocationString+"\n")
			}
		}
	}
}
