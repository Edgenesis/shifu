package telemetryservice

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/edgenesis/shifu/pkg/telemetryservice/mqtt"
	"github.com/edgenesis/shifu/pkg/telemetryservice/sql"
	"k8s.io/klog"
)

var serverListenPort = os.Getenv("SERVER_LISTEN_PORT")

// TODO: need to modify path of mqtt.BindMQTTServicehandler after other servie implement
func New(stop <-chan struct{}) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", matchHandler)
	err := Start(stop, mux, serverListenPort)
	if err != nil {
		klog.Errorf("Error when telemetryService Running, error: %v", err)
	}
}

func Start(stop <-chan struct{}, mux *http.ServeMux, addr string) error {
	var errChan = make(chan error, 1)
	server := http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			klog.Errorf("Error when server running, error: %v", err)
			errChan <- err
		}
	}()

	klog.Infof("Listening at %#v", addr)
	select {
	case err := <-errChan:
		return err
	case <-stop:
		return server.Close()
	}
}

func matchHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		klog.Errorf("Error when Read Data From Body, error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	klog.Infof("requestBody: %s", string(body))
	telemetryRequest := v1alpha1.TelemetryRequest{}

	err = json.Unmarshal(body, &telemetryRequest)
	if err != nil {
		klog.Errorf("Error when unmarshal body to telemetryBody, error: %v", err.Error())
		http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
		return
	}

	if telemetryRequest.MQTTSetting != nil {
		err := mqtt.BindMQTTServicehandler(telemetryRequest)
		if err != nil {
			klog.Errorf("Handler MQTT Servuce handler, error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if telemetryRequest.SQLConnectionSetting != nil {
		err := sql.BindSQLServiceHandler(context.TODO(), telemetryRequest)
		if err != nil {
			klog.Errorf("Handler SQL Servuce handler, error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}
