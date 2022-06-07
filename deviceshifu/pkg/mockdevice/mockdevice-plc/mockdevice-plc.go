package main

import (
	"edgenesis.io/shifu/deviceshifu/pkg/mockdevice/mockdevice"
	"fmt"
	"k8s.io/apimachinery/pkg/util/rand"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	dataStorage       map[string]string
	MemoryArea        = []string{"M", "Q", "T", "C"}
	originalCharacter = "0b0000000000000000"
)

const (
	rootAddress = "rootaddress"
	address     = "address"
	digit       = "digit"
	value       = "value"
)

func main() {
	dataStorage = make(map[string]string)
	for _, v := range MemoryArea {
		dataStorage[v] = originalCharacter
	}

	available_funcs := []string{
		"getcontent",
		"sendsinglebit",
		"get_status",
	}
	mockdevice.StartMockDevice(available_funcs, instructionHandler)
}

func instructionHandler(functionName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Handling: %v", functionName)
		switch functionName {
		case "getcontent":
			query := r.URL.Query()
			rootaddress := query.Get(rootAddress)
			if _, ok := dataStorage[rootaddress]; !ok {
				log.Println("Nonexistent memory area:", rootaddress)
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "Nonexistent memory area")
				return
			}

			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, dataStorage[rootaddress])
		case "sendsinglebit":
			query := r.URL.Query()
			rootaddress := query.Get(rootAddress)
			addressValue, err := strconv.Atoi(query.Get(address))
			if err != nil {
				log.Fatalln(err)
			}

			digitsValue, err := strconv.Atoi(query.Get(digit))
			if err != nil {
				log.Fatalln(err)
			}

			if _, ok := dataStorage[rootaddress]; !ok {
				log.Println("Nonexistent memory area:", rootaddress)
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "Nonexistent memory area")
				return
			}

			valueValue := query.Get(value)
			responseValue := []byte(dataStorage[rootaddress])
			valueModificator := []byte(valueValue)
			responseValue[len(dataStorage[rootaddress])-1-
				addressValue-digitsValue] = valueModificator[0]
			dataStorage[rootaddress] = string(responseValue)
			log.Println(responseValue)
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, dataStorage[rootaddress])
		case "get_status":
			rand.Seed(time.Now().UnixNano())
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, mockdevice.STATUS_STR_LIST[(rand.Intn(len(mockdevice.STATUS_STR_LIST)))])
		}
	}
}
