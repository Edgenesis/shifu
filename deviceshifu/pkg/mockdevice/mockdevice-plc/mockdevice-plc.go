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
				log.Println("error getting", rootaddress)
				fmt.Fprintf(w, "error getting")
				return
			}
			fmt.Fprintf(w, dataStorage[rootaddress])
		case "sendsinglebit":
			query := r.URL.Query()
			rootaddress := query.Get(rootAddress)
			addressValue, err := strconv.Atoi(query.Get(address))
			if err != nil {
				panic(err)
			}
			digitsValue, err := strconv.Atoi(query.Get(digit))
			if err != nil {
				panic(err)
			}
			valueValue := query.Get(value)
			sendsed := []byte(dataStorage[rootaddress])
			send := []byte(valueValue)
			sendsed[len(dataStorage[rootaddress])-1-addressValue-digitsValue] = send[0]
			dataStorage[rootaddress] = string(sendsed)
			log.Println(sendsed)
			fmt.Fprintf(w, dataStorage[rootaddress])
		case "get_status":
			rand.Seed(time.Now().UnixNano())
			fmt.Fprintf(w, mockdevice.STATUS_STR_LIST[(rand.Intn(len(mockdevice.STATUS_STR_LIST)))])
		}
	}
}
