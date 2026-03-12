package mockdevice

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/edgenesis/shifu/pkg/logger"

	"k8s.io/apimachinery/pkg/util/wait"
)

// MockDevice basic info
type MockDevice struct {
	Name   string
	server *http.Server
	addr   string
}

// Driver MockDevice Driver interface include main function and instruction handler
type Driver interface {
	main()
	instructionHandler(string) func(http.ResponseWriter, *http.Request)
}

type instructionHandlerFunc func(string) http.HandlerFunc

// StatusSetList Status Set List
var StatusSetList = []string{
	"Running",
	"Idle",
	"Busy",
	"Error",
}

// Start start http server
func (md *MockDevice) Start(stopCh <-chan struct{}) error {
	logger.Infof("mockDevice %s started", md.Name)

	listener, err := net.Listen("tcp", md.server.Addr)
	if err != nil {
		return err
	}

	md.addr = listener.Addr().String()
	md.server.Addr = md.addr

	if stopCh != nil {
		go func() {
			<-stopCh

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			if err := md.server.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Errorf("error shutting down mockDevice %s http server: %v", md.Name, err)
			}
		}()
	}

	go func() {
		err := md.startHTTPServer(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("mockDevice %s http server error: %v", md.Name, err)
		}
	}()
	return nil
}

func (md *MockDevice) startHTTPServer(listener net.Listener) error {
	logger.Infof("mockDevice %s's http server started", md.Name)
	return md.server.Serve(listener)
}

func (md *MockDevice) URL() string {
	return "http://127.0.0.1:" + md.Port()
}

func (md *MockDevice) Port() string {
	addr := md.addr
	if addr == "" {
		addr = md.server.Addr
	}

	_, port, err := net.SplitHostPort(addr)
	if err == nil {
		return port
	}

	return addr
}

func deviceHealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Healthy")
}

// New new mock device
func New(deviceName string, devicePort string, availableFuncs []string, instructionHandler instructionHandlerFunc) (*MockDevice, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", deviceHealthHandler)
	for _, function := range availableFuncs {
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

// StartMockDevice Start MockDevice
func StartMockDevice(availableFuncs []string, instructionHandler instructionHandlerFunc) {
	deviceName := os.Getenv("MOCKDEVICE_NAME")
	devicePort := os.Getenv("MOCKDEVICE_PORT")
	md, err := New(deviceName, devicePort, availableFuncs, instructionHandler)
	if err != nil {
		logger.Errorf("Error starting device %v", deviceName)
	}

	err = md.Start(wait.NeverStop)
	if err != nil {
		logger.Errorf("Error start MockDevice %#v", err)
	}

	select {}
}
