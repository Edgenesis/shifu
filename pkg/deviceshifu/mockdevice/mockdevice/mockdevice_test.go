package mockdevice

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestStartMockDevice(t *testing.T) {
	os.Setenv("MOCKDEVICE_NAME", "mockdevice_test")
	os.Setenv("MOCKDEVICE_PORT", "12345")
	availableFuncs := []string{
		"get_position",
		"get_status",
	}

	instructionHandler := func(functionName string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Handling: %v", functionName)
			switch functionName {
			case "get_status":
				fmt.Fprintf(w, "Running")
			}
		}
	}

	go StartMockDevice(availableFuncs, instructionHandler)

	time.Sleep(1 * time.Second)
	resp, err := http.Get("http://localhost:12345/get_status")
	if err != nil {
		t.Errorf("HTTP GET returns an error %v", err.Error())
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("cannot read body from response, %+v", err)
	}

	if string(body) != "Running" {
		t.Errorf("Body is not running: %+v", string(body))
	}
}
