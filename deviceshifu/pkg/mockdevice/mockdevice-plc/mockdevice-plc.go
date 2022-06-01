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
	resp              map[string]string
	MemoryArea        = []string{"M", "Q", "T", "C"}
	originalCharacter = "00000000"
)

func main() {
	resp = make(map[string]string)
	for _, v := range MemoryArea {
		resp[v] = "00000000"
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
			rootaddress := query.Get("rootaddress")
			if _, ok := resp[rootaddress]; !ok {
				fmt.Errorf("error getting", rootaddress)
				return
			}
			fmt.Fprintf(w, "0b00000000"+resp[rootaddress])
		case "sendsinglebit":
			query := r.URL.Query()
			rootaddress := query.Get("rootaddress")
			address, err := strconv.Atoi(query.Get("address"))
			if err != nil {
				panic(err)
			}
			digits, err := strconv.Atoi(query.Get("digit"))
			if err != nil {
				panic(err)
			}
			value := query.Get("value")
			sendsed := []byte(resp[rootaddress])
			send := []byte(value)
			sendsed[len(resp[rootaddress])-1-address-digits] = send[0]
			resp[rootaddress] = string(sendsed)
			log.Println(sendsed)
			fmt.Fprintf(w, "0b00000000"+resp[rootaddress])
		case "get_status":
			rand.Seed(time.Now().UnixNano())
			fmt.Fprintf(w, mockdevice.STATUS_STR_LIST[(rand.Intn(len(mockdevice.STATUS_STR_LIST)))])
		}
	}
}
