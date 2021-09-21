package deviceshifu

import (
	"context"
	"errors"
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
	edgeDeviceConfig  *v1alpha1.EdgeDevice
}

type DeviceShifuMetaData struct {
	Name           string
	ConfigFilePath string
	KubeConfigPath string
	Namespace      string
}

type DeviceShifuHTTPHandlerMetaData struct {
	httpClient       *http.Client
	edgeDeviceConfig v1alpha1.EdgeDeviceSpec
	instruction      string
	properties       *DeviceShifuInstruction
}

type DeviceShifuUSBHandlerMetaData struct {
	edgeDeviceConfig v1alpha1.EdgeDeviceSpec
	instruction      string
	properties       *DeviceShifuInstruction
}

const (
	DEVICE_IS_HEALTHY_STR             string = "Device is healthy"
	DEVICE_CONFIGMAP_FOLDER_PATH      string = "/etc/edgedevice/config"
	DEVICE_KUBECONFIG_DO_NOT_LOAD_STR string = "NULL"
	DEVICE_NAMESPACE_DEFAULT          string = "default"
	KUBERNETES_CONFIG_DEFAULT         string = ""
)

// func New(name string, config_file_dir string, kube_config_location string, namespace string) *DeviceShifu {
func New(deviceShifuMetadata *DeviceShifuMetaData) (*DeviceShifu, error) {
	if deviceShifuMetadata.Name == "" {
		return nil, errors.New("DeviceShifu's name can't be empty\n")
	}

	if deviceShifuMetadata.ConfigFilePath == "" {
		deviceShifuMetadata.ConfigFilePath = DEVICE_CONFIGMAP_FOLDER_PATH
	}

	deviceShifuConfig, err := NewDeviceShifuConfig(deviceShifuMetadata.ConfigFilePath)
	if err != nil {
		fmt.Errorf("Error parsing ConfigMap at %v", deviceShifuMetadata.ConfigFilePath)
		return nil, err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", deviceHealthHandler)

	edgeDevice := &v1alpha1.EdgeDevice{}

	if deviceShifuMetadata.KubeConfigPath != DEVICE_KUBECONFIG_DO_NOT_LOAD_STR {
		edgeDeviceConfigMetaData := &EdgeDeviceConfigMetaData{
			deviceShifuMetadata.Namespace,
			deviceShifuMetadata.Name,
			deviceShifuMetadata.KubeConfigPath,
		}

		edgeDevice, err = NewEdgeDeviceConfig(edgeDeviceConfigMetaData)
		if err != nil {
			log.Fatalf("Error parsing EdgeDevice Resource")
			return nil, err
		}

		if &edgeDevice.Spec == nil {
			log.Fatalf("edgeDeviceConfig.Spec is nil")
			return nil, err
		}

		switch protocol := *edgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolHTTP:
			httpClient := &http.Client{Timeout: 3 * time.Second} // TODO: read timeout from EdgeDeviceConfig
			for instruction, properties := range deviceShifuConfig.Instructions {
				deviceShifuHTTPHandlerMetaData := &DeviceShifuHTTPHandlerMetaData{
					httpClient,
					edgeDevice.Spec,
					instruction,
					properties,
				}

				mux.HandleFunc("/"+instruction, deviceCommandHandlerHTTP(deviceShifuHTTPHandlerMetaData))
			}
		case v1alpha1.ProtocolUSB:
			for instruction, properties := range deviceShifuConfig.Instructions {
				deviceShifuUSBHandlerMetaData := &DeviceShifuUSBHandlerMetaData{
					edgeDevice.Spec,
					instruction,
					properties,
				}

				mux.HandleFunc("/"+instruction, deviceCommandHandlerUSB(deviceShifuUSBHandlerMetaData))
			}
		}
	}

	ds := &DeviceShifu{
		Name: deviceShifuMetadata.Name,
		server: &http.Server{
			Addr:         ":8080",
			Handler:      mux,
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 60 * time.Second,
		},
		deviceShifuConfig: deviceShifuConfig,
		edgeDeviceConfig:  edgeDevice,
	}

	return ds, nil
}

func deviceHealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, DEVICE_IS_HEALTHY_STR)
}

// func deviceCommandHandlerHTTP(httpClient *http.Client, edgeDeviceConfig v1alpha1.EdgeDeviceSpec, instruction string, properties *DeviceShifuInstruction) http.HandlerFunc {
func deviceCommandHandlerHTTP(deviceShifuHTTPHandlerMetaData *DeviceShifuHTTPHandlerMetaData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		properties := deviceShifuHTTPHandlerMetaData.properties
		instruction := deviceShifuHTTPHandlerMetaData.instruction
		httpClient := deviceShifuHTTPHandlerMetaData.httpClient
		edgeDeviceConfig := deviceShifuHTTPHandlerMetaData.edgeDeviceConfig

		if properties != nil {
			// TODO: handle validation compile
			for _, property := range properties.DeviceShifuInstructionProperties {
				log.Printf("Properties of command: %v %v\n", instruction, property)
			}
		}

		log.Printf("handling instruction '%v' to '%v'", instruction, edgeDeviceConfig.Address)
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

func deviceCommandHandlerUSB(deviceShifuUSBHandlerMetaData *DeviceShifuUSBHandlerMetaData) http.HandlerFunc {
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
	if err := ds.server.Shutdown(context.TODO()); err != nil {
		return err
	}

	fmt.Printf("deviceShifu %s's http server stopped\n", ds.Name)
	return nil
}
