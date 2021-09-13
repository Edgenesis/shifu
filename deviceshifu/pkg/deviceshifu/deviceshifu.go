package deviceshifu

import (
	"fmt"
	"net/http"
)

type DeviceShifu struct {
	Name string
}

func (ds *DeviceShifu) deviceShifuHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, ds.Name)
}

func (ds *DeviceShifu) startHttpServer(stopCh <-chan struct{}) error {
	http.HandleFunc("/", ds.deviceShifuHandler)
	http.ListenAndServe(":8000", nil)

	return nil
}

func (ds *DeviceShifu) Start(stopCh <-chan struct{}) error {
	fmt.Printf("deviceShifu %s started\n", ds.Name)

	go ds.startHttpServer(stopCh)

	return nil
}
