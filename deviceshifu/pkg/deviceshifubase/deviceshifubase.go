package deviceshifubase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"edgenesis.io/shifu/k8s/crd/api/v1alpha1"
	"k8s.io/client-go/rest"
)

type DeviceShifuBase struct {
	Name              string
	Server            *http.Server
	DeviceShifuConfig *DeviceShifuConfig
	EdgeDevice        *v1alpha1.EdgeDevice
	RestClient        *rest.RESTClient
}

type DeviceShifuMetaData struct {
	Name           string
	ConfigFilePath string
	KubeConfigPath string
	Namespace      string
}

type collectTelemetry func() (bool, error)

type DeviceShifu interface {
	Start(stopCh <-chan struct{}) error
	Stop() error
}

const (
	CM_DRIVERPROPERTIES_STR                            = "driverProperties"
	CM_INSTRUCTIONS_STR                                = "instructions"
	CM_TELEMETRIES_STR                                 = "telemetries"
	EDGEDEVICE_RESOURCE_STR                            = "edgedevices"
	TELEMETRYCOLLECTIONSERVICE_RESOURCE_STR            = "telemetryservices"
	DEVICE_TELEMETRY_TIMEOUT_MS                 int64  = 3000
	DEVICE_TELEMETRY_UPDATE_INTERVAL_MS         int64  = 3000
	DEVICE_TELEMETRY_INITIAL_DELAY_MS           int64  = 3000
	DEVICE_DEFAULT_CONNECTION_TIMEOUT_MS        int64  = 3000
	DEVICE_DEFAULT_PORT_STR                     string = ":8080"
	DEVICE_DEFAULT_REQUEST_TIMEOUT_MS           int64  = 1000
	DEVICE_DEFAULT_TELEMETRY_UPDATE_INTERVAL_MS int64  = 1000
	DEVICE_IS_HEALTHY_STR                       string = "Device is healthy"
	DEVICE_CONFIGMAP_FOLDER_PATH                string = "/etc/edgedevice/config"
	DEVICE_KUBECONFIG_DO_NOT_LOAD_STR           string = "NULL"
	DEVICE_NAMESPACE_DEFAULT                    string = "default"
	KUBERNETES_CONFIG_DEFAULT                   string = ""
	DEVICE_INSTRUCTION_TIMEOUT_URI_QUERY_STR    string = "timeout"
	DEVICE_DEFAULT_GLOBAL_TIMEOUT_SECONDS       int    = 3
	DEFAULT_HTTP_SERVER_TIMEOUT_SECONDS         int    = 0
)

var (
	TelemetryCollectionServiceMap map[string]string
)

func New(deviceShifuMetadata *DeviceShifuMetaData) (*DeviceShifuBase, *http.ServeMux, error) {
	if deviceShifuMetadata.Name == "" {
		return nil, nil, fmt.Errorf("DeviceShifu's name can't be empty\n")
	}

	if deviceShifuMetadata.ConfigFilePath == "" {
		deviceShifuMetadata.ConfigFilePath = DEVICE_CONFIGMAP_FOLDER_PATH
	}

	deviceShifuConfig, err := NewDeviceShifuConfig(deviceShifuMetadata.ConfigFilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("Error parsing ConfigMap at %v\n", deviceShifuMetadata.ConfigFilePath)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", deviceHealthHandler)
	mux.HandleFunc("/", instructionNotFoundHandler)

	edgeDevice := &v1alpha1.EdgeDevice{}
	client := &rest.RESTClient{}

	if deviceShifuMetadata.KubeConfigPath != DEVICE_KUBECONFIG_DO_NOT_LOAD_STR {
		edgeDeviceConfig := &EdgeDeviceConfig{
			NameSpace:      deviceShifuMetadata.Namespace,
			DeviceName:     deviceShifuMetadata.Name,
			KubeconfigPath: deviceShifuMetadata.KubeConfigPath,
		}

		edgeDevice, client, err = NewEdgeDevice(edgeDeviceConfig)
		if err != nil {
			log.Fatalf("Error retrieving EdgeDevice")
			return nil, nil, err
		}

		if &edgeDevice.Spec == nil {
			log.Fatalf("edgeDeviceConfig.Spec is nil")
			return nil, nil, err
		}
	}

	base := &DeviceShifuBase{
		Name: deviceShifuMetadata.Name,
		Server: &http.Server{
			Addr:         DEVICE_DEFAULT_PORT_STR,
			Handler:      mux,
			ReadTimeout:  time.Duration(DEFAULT_HTTP_SERVER_TIMEOUT_SECONDS) * time.Second,
			WriteTimeout: time.Duration(DEFAULT_HTTP_SERVER_TIMEOUT_SECONDS) * time.Second,
		},
		DeviceShifuConfig: deviceShifuConfig,
		EdgeDevice:        edgeDevice,
		RestClient:        client,
	}

	return base, mux, nil
}

func deviceHealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, DEVICE_IS_HEALTHY_STR)
}

func instructionNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Error: Device instruction does not exist!")
	http.Error(w, "Error: Device instruction does not exist!", http.StatusNotFound)
}

func (ds *DeviceShifuBase) UpdateEdgeDeviceResourcePhase(edPhase v1alpha1.EdgeDevicePhase) {
	log.Printf("updating device %v status to: %v\n", ds.Name, edPhase)
	currEdgeDevice := &v1alpha1.EdgeDevice{}
	err := ds.RestClient.Get().
		Namespace(ds.EdgeDevice.Namespace).
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
	} else {
		*currEdgeDevice.Status.EdgeDevicePhase = edPhase
	}

	putResult := &v1alpha1.EdgeDevice{}
	err = ds.RestClient.Put().
		Namespace(ds.EdgeDevice.Namespace).
		Resource(EDGEDEVICE_RESOURCE_STR).
		Name(ds.Name).
		Body(currEdgeDevice).
		Do(context.TODO()).
		Into(putResult)

	if err != nil {
		log.Printf("Unable to update status, error: %v", err)
	}
}

func (ds *DeviceShifuBase) ValidateTelemetryConfig() error {
	if ds.DeviceShifuConfig.Telemetries.DeviceShifuTelemetrySettings == nil {
		ds.DeviceShifuConfig.Telemetries.DeviceShifuTelemetrySettings = &DeviceShifuTelemetrySettings{}
	}

	var dsTelemetrySettings = ds.DeviceShifuConfig.Telemetries.DeviceShifuTelemetrySettings
	if initial := dsTelemetrySettings.DeviceShifuTelemetryInitialDelayInMilliseconds; initial == nil {
		var telemetryInitialDelayInMilliseconds = DEVICE_TELEMETRY_INITIAL_DELAY_MS
		dsTelemetrySettings.DeviceShifuTelemetryInitialDelayInMilliseconds = &telemetryInitialDelayInMilliseconds
	} else if *initial < 0 {
		return errors.New("error deviceShifuTelemetryInitialDelay mustn't be negative number")
	}

	if timeout := dsTelemetrySettings.DeviceShifuTelemetryTimeoutInMilliseconds; timeout == nil {
		var telemetryTimeoutInMilliseconds = DEVICE_TELEMETRY_TIMEOUT_MS
		dsTelemetrySettings.DeviceShifuTelemetryTimeoutInMilliseconds = &telemetryTimeoutInMilliseconds
	} else if *timeout < 0 {
		return errors.New("error deviceShifuTelemetryTimeout mustn't be negative number")
	}

	if interval := dsTelemetrySettings.DeviceShifuTelemetryUpdateIntervalInMilliseconds; interval == nil {
		var telemetryUpdateIntervalInMilliseconds = DEVICE_TELEMETRY_UPDATE_INTERVAL_MS
		dsTelemetrySettings.DeviceShifuTelemetryUpdateIntervalInMilliseconds = &telemetryUpdateIntervalInMilliseconds
	} else if *interval < 0 {
		return errors.New("error deviceShifuTelemetryInterval mustn't be negative number")
	}

	return nil
}

func (ds *DeviceShifuBase) telemetryCollection(fn collectTelemetry) error {
	telemetryOK := true
	status, err := fn()
	log.Printf("Status is: %v", status)
	if err != nil {
		log.Printf("Error is: %v", err.Error())
		telemetryOK = false
	}

	if !status && telemetryOK {
		telemetryOK = false
	}

	if telemetryOK {
		ds.UpdateEdgeDeviceResourcePhase(v1alpha1.EdgeDeviceRunning)
	} else {
		ds.UpdateEdgeDeviceResourcePhase(v1alpha1.EdgeDeviceFailed)
	}

	return nil
}

func (ds *DeviceShifuBase) StartTelemetryCollection(fn collectTelemetry) error {
	log.Println("Wait 5 seconds before updating status")
	time.Sleep(5 * time.Second)
	telemetryUpdateIntervalInMilliseconds := DEVICE_DEFAULT_TELEMETRY_UPDATE_INTERVAL_MS
	var err error
	TelemetryCollectionServiceMap, err = getTelemetryCollectionServiceMap(ds)
	if err != nil {
		return fmt.Errorf("error generating TelemetryCollectionServiceMap, error: %v", err.Error())
	}

	if ds.
		DeviceShifuConfig.
		Telemetries.
		DeviceShifuTelemetrySettings != nil &&
		ds.
			DeviceShifuConfig.
			Telemetries.
			DeviceShifuTelemetrySettings.
			DeviceShifuTelemetryUpdateIntervalInMilliseconds != nil {
		telemetryUpdateIntervalInMilliseconds = *ds.
			DeviceShifuConfig.
			Telemetries.
			DeviceShifuTelemetrySettings.
			DeviceShifuTelemetryUpdateIntervalInMilliseconds
	}

	for {
		ds.telemetryCollection(fn)
		time.Sleep(time.Duration(telemetryUpdateIntervalInMilliseconds) * time.Millisecond)
	}
}

func (ds *DeviceShifuBase) startHttpServer(stopCh <-chan struct{}) error {
	fmt.Printf("deviceShifu %s's http server started\n", ds.Name)
	return ds.Server.ListenAndServe()
}

func (ds *DeviceShifuBase) Start(stopCh <-chan struct{}, fn collectTelemetry) error {
	log.Printf("deviceShifu %s started\n", ds.Name)

	go ds.startHttpServer(stopCh)
	go ds.StartTelemetryCollection(fn)

	return nil
}

func (ds *DeviceShifuBase) Stop() error {
	if err := ds.Server.Shutdown(context.TODO()); err != nil {
		return err
	}

	log.Printf("deviceShifu %s's http server stopped\n", ds.Name)
	return nil
}
