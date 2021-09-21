package deviceshifu

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	v1alpha1 "github.com/Edgenesis/shifu/k8s/crd/api/v1alpha1"
)

type DeviceShifu struct {
	Name              string
	server            *http.Server
	deviceShifuConfig *DeviceShifuConfig
	edgeDeviceConfig  v1alpha1.EdgeDevice
}

const (
	DEVICE_IS_HEALTHY_STR             string = "Device is healthy"
	DEVICE_CONFIGMAP_FOLDER_STR       string = "/etc/edgedevice/config"
	DEVICE_KUBECONFIG_DO_NOT_LOAD_STR string = "NULL"
	DEVICE_NAMESPACE_DEFAULT          string = "default"
	KUBERNETES_CONFIG_DEFAULT         string = ""
)

func New(name string, config_file_dir string, kube_config_location string, namespace string) *DeviceShifu {
	if name == "" {
		fmt.Errorf("DeviceShifu's name can't be empty\n")
		return nil
	}

	if config_file_dir == "" {
		config_file_dir = DEVICE_CONFIGMAP_FOLDER_STR
	}

	deviceShifuConfig, err := NewDeviceShifuConfig(config_file_dir)
	if err != nil {
		fmt.Errorf("Error parsing ConfigMap at %v", config_file_dir)
		return nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", deviceHealthHandler)

	edgeDeviceConfig := &v1alpha1.EdgeDevice{}

	if kube_config_location != DEVICE_KUBECONFIG_DO_NOT_LOAD_STR {
		edgeDeviceConfig, err = NewEdgeDeviceConfig(namespace, name, kube_config_location)
		if err != nil {
			log.Fatalf("Error parsing EdgeDevice Resource")
			return nil
		}

		switch protocol := *edgeDeviceConfig.Spec.Protocol; protocol {
		case v1alpha1.ProtocolHTTP:
			httpClient := &http.Client{Timeout: 3 * time.Second} // TODO: read timeout from EdgeDeviceConfig
			for instruction, properties := range deviceShifuConfig.Instructions {
				mux.HandleFunc("/"+instruction, deviceCommandHandlerHTTP(httpClient, edgeDeviceConfig.Spec, instruction, properties))
			}
		case v1alpha1.ProtocolUSB:
			for instruction, properties := range deviceShifuConfig.Instructions {
				mux.HandleFunc("/"+instruction, deviceCommandHandlerUSB(edgeDeviceConfig.Spec, instruction, properties))
			}
		}
	}

	ds := &DeviceShifu{
		Name: name,
		server: &http.Server{
			Addr:         ":8080",
			Handler:      mux,
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 60 * time.Second,
		},
		deviceShifuConfig: deviceShifuConfig,
		edgeDeviceConfig:  *edgeDeviceConfig,
	}

	return ds
}

func deviceHealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, DEVICE_IS_HEALTHY_STR)
}

func deviceCommandHandlerHTTP(httpClient *http.Client, edgeDeviceConfig v1alpha1.EdgeDeviceSpec, instruction string, properties *DeviceShifuInstruction) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if properties != nil {
			// TODO: handle validation compile
			for _, property := range properties.DeviceShifuInstructionProperties {
				log.Printf("Properties of command: %v %v\n", instruction, property)
			}
		}

		log.Printf("handling instruction '%v' to '%v'", instruction, *edgeDeviceConfig.Address)
		resp, err := httpClient.Get("http://" + *edgeDeviceConfig.Address + "/" + instruction)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}

		if resp != nil {
			copyHeader(w.Header(), resp.Header)
			w.WriteHeader(resp.StatusCode)
			io.Copy(w, resp.Body)
		} else {
			// TODO: For now, just write tht instruction to the response
			log.Println("resp is nil")
			w.Write([]byte(instruction))
		}
	}
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func deviceCommandHandlerUSB(edgeDeviceConfig v1alpha1.EdgeDeviceSpec, instruction string, properties *DeviceShifuInstruction) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: handle commands for USB devices
	}
}

func (ds *DeviceShifu) startHttpServer(stopCh <-chan struct{}) error {
	fmt.Printf("deviceShifu %s's http server started\n", ds.Name)
	return ds.server.ListenAndServe()
}

// TODO: update configs
// TODO: update status based on telemetry

func (ds *DeviceShifu) Start(stopCh <-chan struct{}) error {
	fmt.Printf("deviceShifu %s started\n", ds.Name)

	go ds.startHttpServer(stopCh)

	return nil
}

func (ds *DeviceShifu) Stop() error {
	err := ds.server.Shutdown(context.TODO())
	if err != nil {
		return err
	}

	fmt.Printf("deviceShifu %s's http server stopped\n", ds.Name)
	return nil
}
