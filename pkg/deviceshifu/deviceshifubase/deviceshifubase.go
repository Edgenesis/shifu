package deviceshifubase

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"

	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

// DeviceShifuBase deviceshifu Basic Info
type DeviceShifuBase struct {
	Name              string
	Server            *http.Server
	DeviceShifuConfig *DeviceShifuConfig
	EdgeDevice        *v1alpha1.EdgeDevice
	RestClient        *rest.RESTClient
}

// DeviceShifuMetaData Deviceshifu MetaData
type DeviceShifuMetaData struct {
	Name           string
	ConfigFilePath string
	KubeConfigPath string
	Namespace      string
}

// collectTelemetry struct of collectTelemetry
type collectTelemetry func() (bool, error)

// DeviceShifu interface of Deviceshifu include start telemetry and stop http server
type DeviceShifu interface {
	Start(stopCh <-chan struct{}) error
	Stop() error
}

// Str and default value
const (
	ConfigmapDriverPropertiesStr                    = "driverProperties"
	ConfigmapInstructionsStr                        = "instructions"
	ConfigmapTelemetriesStr                         = "telemetries"
	EdgedeviceResourceStr                           = "edgedevices"
	TelemetryCollectionServiceResourceStr           = "telemetryservices"
	DeviceTelemetryTimeoutInMS               int64  = 3000
	DeviceTelemetryUpdateIntervalInMS        int64  = 3000
	DeviceTelemetryInitialDelayInMS          int64  = 3000
	DeviceDefaultConnectionTimeoutInMS       int64  = 3000
	DeviceDefaultPortStr                     string = ":8080"
	DeviceDefaultRequestTimeoutInMS          int64  = 1000
	DeviceDefaultTelemetryUpdateIntervalInMS int64  = 1000
	DeviceIsHealthyStr                       string = "Device is healthy"
	DeviceConfigmapFolderPath                string = "/etc/edgedevice/config"
	DeviceKubeconfigDoNotLoadStr             string = "NULL"
	DeviceNameSpaceDefault                   string = "default"
	KubernetesConfigDefault                  string = ""
	DevuceInstructionTimeoutURIQueryStr      string = "timeout"
	DeviceDefaultGolbalTimeoutInSeconds      int    = 3
	DefaultHTTPServerTimeoutInSeconds        int    = 0
)

var (
	// TelemetryCollectionServiceMap Telemetry Collection Service Map
	TelemetryCollectionServiceMap map[string]string
)

// New new deviceshifu base
func New(deviceShifuMetadata *DeviceShifuMetaData) (*DeviceShifuBase, *http.ServeMux, error) {
	if deviceShifuMetadata.Name == "" {
		return nil, nil, fmt.Errorf("DeviceShifu's name can't be empty")
	}

	if deviceShifuMetadata.ConfigFilePath == "" {
		deviceShifuMetadata.ConfigFilePath = DeviceConfigmapFolderPath
	}

	deviceShifuConfig, err := NewDeviceShifuConfig(deviceShifuMetadata.ConfigFilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("Error parsing ConfigMap at %v", deviceShifuMetadata.ConfigFilePath)
	}

	mux := http.NewServeMux()
	edgeDevice := &v1alpha1.EdgeDevice{}
	client := &rest.RESTClient{}

	if deviceShifuMetadata.KubeConfigPath != DeviceKubeconfigDoNotLoadStr {
		edgeDeviceConfig := &EdgeDeviceConfig{
			NameSpace:      deviceShifuMetadata.Namespace,
			DeviceName:     deviceShifuMetadata.Name,
			KubeconfigPath: deviceShifuMetadata.KubeConfigPath,
		}

		edgeDevice, client, err = NewEdgeDevice(edgeDeviceConfig)
		if err != nil {
			klog.Fatalf("Error retrieving EdgeDevice")
			return nil, nil, err
		}
	}

	base := &DeviceShifuBase{
		Name: deviceShifuMetadata.Name,
		Server: &http.Server{
			Addr:         DeviceDefaultPortStr,
			Handler:      mux,
			ReadTimeout:  time.Duration(DefaultHTTPServerTimeoutInSeconds) * time.Second,
			WriteTimeout: time.Duration(DefaultHTTPServerTimeoutInSeconds) * time.Second,
		},
		DeviceShifuConfig: deviceShifuConfig,
		EdgeDevice:        edgeDevice,
		RestClient:        client,
	}

	return base, mux, nil
}

func BindDefaultHandler(mux *http.ServeMux) {
	mux.HandleFunc("/health", deviceHealthHandler)
	mux.HandleFunc("/", instructionNotFoundHandler)
}

func deviceHealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, DeviceIsHealthyStr)
}

func instructionNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	klog.Errorf("Error: Device instruction does not exist!")
	http.Error(w, "Error: Device instruction does not exist!", http.StatusNotFound)
}

// UpdateEdgeDeviceResourcePhase Update device status
func (ds *DeviceShifuBase) UpdateEdgeDeviceResourcePhase(edPhase v1alpha1.EdgeDevicePhase) {
	klog.Infof("updating device %v status to: %v", ds.Name, edPhase)
	currEdgeDevice := &v1alpha1.EdgeDevice{}
	err := ds.RestClient.Get().
		Namespace(ds.EdgeDevice.Namespace).
		Resource(EdgedeviceResourceStr).
		Name(ds.Name).
		Do(context.TODO()).
		Into(currEdgeDevice)

	if err != nil {
		klog.Errorf("Unable to update status, error: %v", err.Error())
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
		Resource(EdgedeviceResourceStr).
		Name(ds.Name).
		Body(currEdgeDevice).
		Do(context.TODO()).
		Into(putResult)

	if err != nil {
		klog.Errorf("Unable to update status, error: %v", err)
	}
}

// ValidateTelemetryConfig Validate TelemetryConfig
func (ds *DeviceShifuBase) ValidateTelemetryConfig() error {
	if ds.DeviceShifuConfig.Telemetries.DeviceShifuTelemetrySettings == nil {
		ds.DeviceShifuConfig.Telemetries.DeviceShifuTelemetrySettings = &DeviceShifuTelemetrySettings{}
	}

	var dsTelemetrySettings = ds.DeviceShifuConfig.Telemetries.DeviceShifuTelemetrySettings
	if initial := dsTelemetrySettings.DeviceShifuTelemetryInitialDelayInMilliseconds; initial == nil {
		var telemetryInitialDelayInMilliseconds = DeviceTelemetryInitialDelayInMS
		dsTelemetrySettings.DeviceShifuTelemetryInitialDelayInMilliseconds = &telemetryInitialDelayInMilliseconds
	} else if *initial < 0 {
		return errors.New("error deviceShifuTelemetryInitialDelay mustn't be negative number")
	}

	if timeout := dsTelemetrySettings.DeviceShifuTelemetryTimeoutInMilliseconds; timeout == nil {
		var telemetryTimeoutInMilliseconds = DeviceTelemetryTimeoutInMS
		dsTelemetrySettings.DeviceShifuTelemetryTimeoutInMilliseconds = &telemetryTimeoutInMilliseconds
	} else if *timeout < 0 {
		return errors.New("error deviceShifuTelemetryTimeout mustn't be negative number")
	}

	if interval := dsTelemetrySettings.DeviceShifuTelemetryUpdateIntervalInMilliseconds; interval == nil {
		var telemetryUpdateIntervalInMilliseconds = DeviceTelemetryUpdateIntervalInMS
		dsTelemetrySettings.DeviceShifuTelemetryUpdateIntervalInMilliseconds = &telemetryUpdateIntervalInMilliseconds
	} else if *interval < 0 {
		return errors.New("error deviceShifuTelemetryInterval mustn't be negative number")
	}

	return nil
}

func (ds *DeviceShifuBase) telemetryCollection(fn collectTelemetry) error {
	telemetryOK := true
	status, err := fn()
	klog.Infof("Status is: %v", status)
	if err != nil {
		klog.Errorf("Error is: %v", err.Error())
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

// StartTelemetryCollection Start TelemetryCollection
func (ds *DeviceShifuBase) StartTelemetryCollection(fn collectTelemetry) error {
	klog.Infof("Wait 5 seconds before updating status")
	time.Sleep(5 * time.Second)
	telemetryUpdateIntervalInMilliseconds := DeviceDefaultTelemetryUpdateIntervalInMS
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
		err := ds.telemetryCollection(fn)
		if err != nil {
			klog.Errorf("error when telemetry collection, error: %v", err)
			return err
		}
		time.Sleep(time.Duration(telemetryUpdateIntervalInMilliseconds) * time.Millisecond)
	}
}

func (ds *DeviceShifuBase) startHTTPServer(stopCh <-chan struct{}) error {
	klog.Infof("deviceshifu %s's http server started", ds.Name)
	return ds.Server.ListenAndServe()
}

// Start HTTP server and telemetryCollection
func (ds *DeviceShifuBase) Start(stopCh <-chan struct{}, fn collectTelemetry) error {
	klog.Infof("deviceshifu %s started", ds.Name)

	go func() {
		err := ds.startHTTPServer(stopCh)
		if err != nil {
			klog.Errorf("error during Http Server is up, error: %v", err)
		}
	}()
	go func() {
		err := ds.StartTelemetryCollection(fn)
		if err != nil {
			klog.Errorf("error during Telemetry is running, error: %v", err)
		}
	}()
	return nil
}

// Stop Stop http server
func (ds *DeviceShifuBase) Stop() error {
	if err := ds.Server.Shutdown(context.TODO()); err != nil {
		return err
	}

	klog.Infof("deviceshifu %s's http server stopped", ds.Name)
	return nil
}
