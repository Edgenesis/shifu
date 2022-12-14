package deviceshifubase

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"

	"k8s.io/client-go/rest"
)

var zlog *zap.SugaredLogger

func init() {
	logger, _ := zap.NewProduction()
	zlog = logger.Sugar()
}

// DeviceShifuBase deviceshifu Basic Info
type DeviceShifuBase struct {
	Name              string
	Server            *http.Server
	DeviceShifuConfig *DeviceShifuConfig
	DeviceShifuSecret DeviceShifuSecret
	EdgeDevice        *v1alpha1.EdgeDevice
	RestClient        *rest.RESTClient
}

type DeviceShifuSecret map[string]string

// DeviceShifuMetaData Deviceshifu MetaData
type DeviceShifuMetaData struct {
	Name           string
	ConfigFilePath string
	SecretFilePath string
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
	ConfigmapDriverPropertiesStr                 = "driverProperties"
	ConfigmapInstructionsStr                     = "instructions"
	ConfigmapTelemetriesStr                      = "telemetries"
	ConfigmapCustomizedInstructionsStr           = "customInstructionsPython"
	EdgedeviceResourceStr                        = "edgedevices"
	TelemetryCollectionServiceResourceStr        = "telemetryservices"
	DeviceDefaultPortStr                  string = ":8080"
	DeviceIsHealthyStr                    string = "Device is healthy"
	DeviceConfigmapFolderPath             string = "/etc/edgedevice/config"
	DeviceSecretFolderPath                string = "/etc/edgedevice/secret"
	DeviceKubeconfigDoNotLoadStr          string = "NULL"
	DeviceNameSpaceDefault                string = "default"
	KubernetesConfigDefault               string = ""
	DeviceInstructionTimeoutURIQueryStr   string = "timeout"
	DeviceDefaultCMDDoNotExec             string = "issue_cmd"
	DeviceDefaultCMDStubHealth            string = "stub_health"
	PowerShellStubTimeoutStr              string = "cmdTimeout"
	PowerShellStubTimeoutTolerationStr    string = "stub_toleration"
	PythonHandlersModuleName                     = "customized_handlers"
	PythonScriptDir                              = "pythoncustomizedhandlers"
	SQLSettingSecret                             = "telemetry_service_sql_pwd"
	HTTPSettingSecret                            = "telemetry_service_http_pwd"
)

var (
	// TelemetryCollectionServiceMap Telemetry Collection Service Map
	TelemetryCollectionServiceMap map[string]v1alpha1.TelemetryServiceSpec
	CustomInstructionsPython      map[string]string
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

	if deviceShifuMetadata.SecretFilePath == "" {
		deviceShifuMetadata.SecretFilePath = DeviceSecretFolderPath
	}

	deviceShifuSecret, err := NewDeviceShifuSecret(deviceShifuMetadata.SecretFilePath)
	if err != nil {
		klog.Infof("error: %v, when parsing Secret at %v, use the default plaintext password", err, deviceShifuMetadata.SecretFilePath)
	}

	mux := http.NewServeMux()
	edgeDevice := &v1alpha1.EdgeDevice{}
	client := &rest.RESTClient{}

	CustomInstructionsPython = deviceShifuConfig.CustomInstructionsPython
	zlog.Infof("configured custom instruction: %v\n", deviceShifuConfig.CustomInstructionsPython)
	zlog.Infof("read custom instruction: %v\n", CustomInstructionsPython)

	if deviceShifuMetadata.KubeConfigPath != DeviceKubeconfigDoNotLoadStr {
		edgeDeviceConfig := &EdgeDeviceConfig{
			NameSpace:      deviceShifuMetadata.Namespace,
			DeviceName:     deviceShifuMetadata.Name,
			KubeconfigPath: deviceShifuMetadata.KubeConfigPath,
		}

		edgeDevice, client, err = NewEdgeDevice(edgeDeviceConfig)
		if err != nil {
			zlog.Errorf("Error retrieving EdgeDevice")
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
		DeviceShifuSecret: deviceShifuSecret,
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
	zlog.Errorf("Error: Device instruction does not exist!")
	http.Error(w, "Error: Device instruction does not exist!", http.StatusNotFound)
}

// UpdateEdgeDeviceResourcePhase Update device status
func (ds *DeviceShifuBase) UpdateEdgeDeviceResourcePhase(edPhase v1alpha1.EdgeDevicePhase) {
	zlog.Infof("updating device %v status to: %v", ds.Name, edPhase)
	currEdgeDevice := &v1alpha1.EdgeDevice{}
	err := ds.RestClient.Get().
		Namespace(ds.EdgeDevice.Namespace).
		Resource(EdgedeviceResourceStr).
		Name(ds.Name).
		Do(context.TODO()).
		Into(currEdgeDevice)

	if err != nil {
		zlog.Errorf("Unable to update status, error: %v", err.Error())
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
		zlog.Errorf("Unable to update status, error: %v", err)
	}
}

func (ds *DeviceShifuBase) telemetryCollection(fn collectTelemetry) error {
	telemetryOK := true
	status, err := fn()
	zlog.Infof("Status is: %v", status)
	if err != nil {
		zlog.Errorf("Error is: %v", err.Error())
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
	zlog.Infof("Wait 5 seconds before updating status")
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
			zlog.Errorf("error when telemetry collection, error: %v", err)
			return err
		}
		time.Sleep(time.Duration(telemetryUpdateIntervalInMilliseconds) * time.Millisecond)
	}
}

func (ds *DeviceShifuBase) startHTTPServer(stopCh <-chan struct{}) error {
	zlog.Infof("deviceshifu %s's http server started", ds.Name)
	return ds.Server.ListenAndServe()
}

// Start HTTP server and telemetryCollection
func (ds *DeviceShifuBase) Start(stopCh <-chan struct{}, fn collectTelemetry) error {
	zlog.Infof("deviceshifu %s started", ds.Name)

	go func() {
		err := ds.startHTTPServer(stopCh)
		if err != nil {
			zlog.Errorf("error during Http Server is up, error: %v", err)
		}
	}()
	go func() {
		err := ds.StartTelemetryCollection(fn)
		if err != nil {
			zlog.Errorf("error during Telemetry is running, error: %v", err)
		}
	}()
	return nil
}

// Stop Stop http server
func (ds *DeviceShifuBase) Stop() error {
	if err := ds.Server.Shutdown(context.TODO()); err != nil {
		return err
	}

	zlog.Infof("deviceshifu %s's http server stopped", ds.Name)
	return nil
}
