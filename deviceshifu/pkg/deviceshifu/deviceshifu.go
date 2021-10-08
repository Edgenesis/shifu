package deviceshifu

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	v1alpha1 "edgenesis.io/shifu/k8s/crd/api/v1alpha1"
	"k8s.io/client-go/rest"
)

type DeviceShifu struct {
	Name              string
	server            *http.Server
	deviceShifuConfig *DeviceShifuConfig
	edgeDevice        *v1alpha1.EdgeDevice
	restClient        *rest.RESTClient
}

type DeviceShifuMetaData struct {
	Name           string
	ConfigFilePath string
	KubeConfigPath string
	Namespace      string
}

type DeviceShifuHTTPHandlerMetaData struct {
	httpClient     *http.Client
	edgeDeviceSpec v1alpha1.EdgeDeviceSpec
	instruction    string
	properties     *DeviceShifuInstruction
}

type DeviceShifuUSBHandlerMetaData struct {
	edgeDeviceSpec v1alpha1.EdgeDeviceSpec
	instruction    string
	properties     *DeviceShifuInstruction
}

const (
	DEVICE_IS_HEALTHY_STR             string = "Device is healthy"
	DEVICE_CONFIGMAP_FOLDER_PATH      string = "/etc/edgedevice/config"
	DEVICE_KUBECONFIG_DO_NOT_LOAD_STR string = "NULL"
	DEVICE_NAMESPACE_DEFAULT          string = "default"
	DEVICE_DEFAULT_PORT_STR           string = ":8080"
	KUBERNETES_CONFIG_DEFAULT         string = ""
)

func New(deviceShifuMetadata *DeviceShifuMetaData) (*DeviceShifu, error) {
	if deviceShifuMetadata.Name == "" {
		return nil, fmt.Errorf("DeviceShifu's name can't be empty\n")
	}

	if deviceShifuMetadata.ConfigFilePath == "" {
		deviceShifuMetadata.ConfigFilePath = DEVICE_CONFIGMAP_FOLDER_PATH
	}

	deviceShifuConfig, err := NewDeviceShifuConfig(deviceShifuMetadata.ConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf("Error parsing ConfigMap at %v\n", deviceShifuMetadata.ConfigFilePath)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", deviceHealthHandler)

	edgeDevice := &v1alpha1.EdgeDevice{}
	client := &rest.RESTClient{}

	if deviceShifuMetadata.KubeConfigPath != DEVICE_KUBECONFIG_DO_NOT_LOAD_STR {
		edgeDeviceConfig := &EdgeDeviceConfig{
			deviceShifuMetadata.Namespace,
			deviceShifuMetadata.Name,
			deviceShifuMetadata.KubeConfigPath,
		}

		edgeDevice, client, err = NewEdgeDevice(edgeDeviceConfig)
		if err != nil {
			log.Fatalf("Error retrieving EdgeDevice")
			return nil, err
		}

		if &edgeDevice.Spec == nil {
			log.Fatalf("edgeDeviceConfig.Spec is nil")
			return nil, err
		}

		switch protocol := *edgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolHTTP:
			for instruction, properties := range deviceShifuConfig.Instructions {
				deviceShifuHTTPHandlerMetaData := &DeviceShifuHTTPHandlerMetaData{
					client.Client,
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
			Addr:         DEVICE_DEFAULT_PORT_STR,
			Handler:      mux,
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 60 * time.Second,
		},
		deviceShifuConfig: deviceShifuConfig,
		edgeDevice:        edgeDevice,
		restClient:        client,
	}

	ds.updateEdgeDeviceResourceStatus(v1alpha1.EdgeDevicePending)
	return ds, nil
}

func deviceHealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, DEVICE_IS_HEALTHY_STR)
}

func deviceCommandHandlerHTTP(deviceShifuHTTPHandlerMetaData *DeviceShifuHTTPHandlerMetaData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handlerProperties := deviceShifuHTTPHandlerMetaData.properties
		handlerInstruction := deviceShifuHTTPHandlerMetaData.instruction
		handlerHTTPClient := deviceShifuHTTPHandlerMetaData.httpClient
		handlerEdgeDevice := deviceShifuHTTPHandlerMetaData.edgeDeviceSpec

		if handlerProperties != nil {
			// TODO: handle validation compile
			for _, instructionProperty := range handlerProperties.DeviceShifuInstructionProperties {
				log.Printf("Properties of command: %v %v\n", handlerInstruction, instructionProperty)
			}
		}

		log.Printf("handling instruction '%v' to '%v'", handlerInstruction, *handlerEdgeDevice.Address)
		resp, err := handlerHTTPClient.Get("http://" + *handlerEdgeDevice.Address + "/" + handlerInstruction)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}

		if resp != nil {
			copyHeader(w.Header(), resp.Header)
			w.WriteHeader(resp.StatusCode)
			io.Copy(w, resp.Body)
			return
		}

		// TODO: For now, just write tht instruction to the response
		log.Println("resp is nil")
		w.Write([]byte(handlerInstruction))
	}
}

// HTTP header type:
// type Header map[string][]string
func copyHeader(dst, src http.Header) {
	for header, headerValueList := range src {
		for _, value := range headerValueList {
			dst.Add(header, value)
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

func (ds *DeviceShifu) collectHTTPTelemetry(telemetry string, telemetryProperties DeviceShifuTelemetryProperties) (bool, error) {
	if ds.edgeDevice.Spec.Address == nil {
		return false, fmt.Errorf("Device %v does not have an address", ds.Name)
	}

	if telemetryProperties.DeviceInstructionName == nil {
		return false, fmt.Errorf("Device %v telemetry %v does not have an instruction name", ds.Name, telemetry)
	}

	address := *ds.edgeDevice.Spec.Address
	instruction := *telemetryProperties.DeviceInstructionName
	resp, err := ds.restClient.Client.Get("http://" + address + "/" + instruction)
	if err != nil {
		log.Printf("error checking telemetry: %v, error: %v", telemetry, err.Error())
		return false, err
	}

	if resp != nil {
		if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
			return true, nil
		}
	}

	return false, nil
}

func (ds *DeviceShifu) collectHTTPTelemetries() error {
	telemetryOK := true
	telemetries := ds.deviceShifuConfig.Telemetries
	for telemetry, telemetryProperties := range telemetries {
		status, err := ds.collectHTTPTelemetry(telemetry, telemetryProperties.DeviceShifuTelemetryProperties)
		if err != nil {
			if !status && telemetryOK {
				telemetryOK = false
			}
		}
	}

	if telemetryOK {
		ds.updateEdgeDeviceResourceStatus(v1alpha1.EdgeDeviceRunning)
	} else {
		ds.updateEdgeDeviceResourceStatus(v1alpha1.EdgeDeviceFailed)
	}

	return nil
}

func (ds *DeviceShifu) telemetryCollection() error {
	// TODO: handle interval for different telemetries
	log.Printf("deviceShifu %s's telemetry collection started\n", ds.Name)

	if ds.edgeDevice.Spec.Protocol != nil {
		switch protocol := *ds.edgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolHTTP:
			ds.collectHTTPTelemetries()
		default:
			log.Printf("EdgeDevice protocol %v not supported in deviceShifu\n", protocol)
			ds.updateEdgeDeviceResourceStatus(v1alpha1.EdgeDeviceFailed)
		}

		return nil
	}

	return fmt.Errorf("EdgeDevice %v has no telemetry field in configuration\n", ds.Name)
}

func (ds *DeviceShifu) updateEdgeDeviceResourceStatus(status v1alpha1.EdgeDevicePhase) {
	log.Printf("updating device %v status to: %v\n", ds.Name, status)
	currEdgeDevice := &v1alpha1.EdgeDevice{}
	err := ds.restClient.Get().
		Namespace(ds.edgeDevice.Namespace).
		Resource(EDGEDEVICE_RESOURCE_STR).
		Name(ds.Name).
		Do(context.TODO()).
		Into(currEdgeDevice)

	if err != nil {
		log.Printf("Unable to update status, error: %v", err.Error())
		return
	}

	if currEdgeDevice.Status.EdgeDevicePhase == nil {
		edgeDeviceStatus := v1alpha1.EdgeDevicePending
		currEdgeDevice.Status.EdgeDevicePhase = &edgeDeviceStatus
	}

	*currEdgeDevice.Status.EdgeDevicePhase = status

	putResult := &v1alpha1.EdgeDevice{}
	err = ds.restClient.Put().
		Namespace(ds.edgeDevice.Namespace).
		Resource(EDGEDEVICE_RESOURCE_STR).
		Name(ds.Name).
		Body(currEdgeDevice).
		Do(context.TODO()).
		Into(putResult)

	if err != nil {
		log.Printf("Unable to update status, error: %v", err)
	}
}

func (ds *DeviceShifu) StartTelemetryCollection() error {
	log.Println("Wait 5 seconds before updating status")
	time.Sleep(5 * time.Second)
	for {
		ds.telemetryCollection()
		time.Sleep(5 * time.Second)
	}
}

func (ds *DeviceShifu) Start(stopCh <-chan struct{}) error {
	fmt.Printf("deviceShifu %s started\n", ds.Name)

	go ds.startHttpServer(stopCh)
	go ds.StartTelemetryCollection()

	return nil
}

func (ds *DeviceShifu) Stop() error {
	if err := ds.server.Shutdown(context.TODO()); err != nil {
		return err
	}

	fmt.Printf("deviceShifu %s's http server stopped\n", ds.Name)
	return nil
}
