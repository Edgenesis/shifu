package deviceshifu

import (
	"fmt"
	"net/http"
	"time"
)

type DeviceShifu struct {
	Name   string
	server *http.Server
}

const (
	DEVICE_IS_HEALTHY_STR string = "Device is healthy"
)

func New(name string) *DeviceShifu {
	if name == "" {
		fmt.Errorf("DeviceShifu's name can't be empty\n")
		return nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", deviceHealthHandler)

	ds := &DeviceShifu{
		Name: name,
		server: &http.Server{
			Addr:         ":8080",
			Handler:      mux,
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 60 * time.Second,
		},
	}

	return ds
}

func deviceHealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, DEVICE_IS_HEALTHY_STR)
}

func (ds *DeviceShifu) startHttpServer(stopCh <-chan struct{}) error {
	fmt.Printf("deviceShifu %s's http server started\n", ds.Name)
	return ds.server.ListenAndServe()
}

func (ds *DeviceShifu) Start(stopCh <-chan struct{}) error {
	fmt.Printf("deviceShifu %s started\n", ds.Name)

	go ds.startHttpServer(stopCh)

	return nil
}
