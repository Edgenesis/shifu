package mockdevice

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

type MockDevice struct {
	Name   string
	server *http.Server
}

type MockDeviceDriver interface {
	main()
	instructionHandler(string) func(http.ResponseWriter, *http.Request)
}

type instructionHandlerFunc func(string) http.HandlerFunc

var STATUS_STR_LIST = []string{
	"Running",
	"Idle",
	"Busy",
	"Error",
}

func (md *MockDevice) Start(stopCh <-chan struct{}) error {
	log.Printf("mockDevice %s started\n", md.Name)

	go md.startHttpServer(stopCh)

	return nil
}

func (md *MockDevice) startHttpServer(stopCh <-chan struct{}) error {
	log.Printf("mockDevice %s's http server started\n", md.Name)
	return md.server.ListenAndServe()
}

func deviceHealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Healthy")
}

func New(deviceName string, devicePort string, available_funcs []string, instructionHandler instructionHandlerFunc) (*MockDevice, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", deviceHealthHandler)
	for _, function := range available_funcs {
		mux.HandleFunc("/"+function, instructionHandler(function))
	}

	md := &MockDevice{
		Name: deviceName,
		server: &http.Server{
			Addr:         ":" + devicePort,
			Handler:      mux,
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 60 * time.Second,
		},
	}
	return md, nil
}

func StartMockDevice(available_funcs []string, instructionHandler instructionHandlerFunc) {
	deviceName := os.Getenv("MOCKDEVICE_NAME")
	devicePort := os.Getenv("MOCKDEVICE_PORT")
	// available_funcs := []string{"read_value", "get_status"}
	md, err := New(deviceName, devicePort, available_funcs, instructionHandler)
	if err != nil {
		log.Printf("Error starting device %v", deviceName)
	}

	md.Start(wait.NeverStop)

	select {}
}
