package deviceshifuSocket

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
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
	socketConnection  *net.Conn
}

type DeviceShifuMetaData struct {
	Name           string
	ConfigFilePath string
	KubeConfigPath string
	Namespace      string
}

type DeviceShifuSocketHandlerMetaData struct {
	edgeDeviceSpec v1alpha1.EdgeDeviceSpec
	instruction    string
	properties     *DeviceShifuInstruction
	connection     *net.Conn
}

type deviceCommandHandler interface {
	commandHandleFunc(w http.ResponseWriter, r *http.Request) http.HandlerFunc
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
	mux.HandleFunc("/", instructionNotFoundHandler)

	edgeDevice := &v1alpha1.EdgeDevice{}
	client := &rest.RESTClient{}
	var socketConnection net.Conn

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
		case v1alpha1.ProtocolSocket:
			// Open the connection:
			connectionType := edgeDevice.Spec.ProtocolSettings.SocketSetting.NetworkType
			encoding := edgeDevice.Spec.ProtocolSettings.SocketSetting.Encoding
			if connectionType == nil || *connectionType != "tcp" {
				return nil, fmt.Errorf("Sorry!, Shifu currently only support TCP Socket")
			}

			if encoding == nil {
				log.Println("Socket encoding not specified, default to UTF-8")
				return nil, fmt.Errorf("Encoding error")
			}

			socketConnection, err := net.Dial(*connectionType, *edgeDevice.Spec.Address)
			if err != nil {
				return nil, fmt.Errorf("Cannot connect to %v", *edgeDevice.Spec.Address)
			}

			log.Printf("Connected to '%v'\n", *edgeDevice.Spec.Address)
			for instruction, properties := range deviceShifuConfig.Instructions {
				deviceShifuSocketHandlerMetaData := &DeviceShifuSocketHandlerMetaData{
					edgeDevice.Spec,
					instruction,
					properties,
					&socketConnection,
				}

				mux.HandleFunc("/"+instruction, deviceCommandHandlerSocket(deviceShifuSocketHandlerMetaData))
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
		socketConnection:  &socketConnection,
	}

	ds.updateEdgeDeviceResourcePhase(v1alpha1.EdgeDevicePending)
	return ds, nil
}

func deviceHealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, DEVICE_IS_HEALTHY_STR)
}

func instructionNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Error: Device instruction does not exist!")
	http.Error(w, "Error: Device instruction does not exist!", http.StatusNotFound)
}

func createUriFromRequest(address string, handlerInstruction string, r *http.Request) string {

	queryStr := "?"

	for queryName, queryValues := range r.URL.Query() {
		for _, queryValue := range queryValues {
			queryStr += queryName + "=" + queryValue + "&"
		}
	}

	queryStr = strings.TrimSuffix(queryStr, "&")

	if queryStr == "?" {
		return "http://" + address + "/" + handlerInstruction
	}

	return "http://" + address + "/" + handlerInstruction + queryStr
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

func deviceCommandHandlerSocket(deviceShifuSocketHandlerMetaData *DeviceShifuSocketHandlerMetaData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		headerContentType := r.Header.Get("Content-Type")
		if headerContentType != "application/json" {
			http.Error(w, "content-type is not application/json", http.StatusBadRequest)
			log.Println("content-type is not application/json")
			return
		}

		var socketRequest DeviceShifuSocketRequestBody
		err := json.NewDecoder(r.Body).Decode(&socketRequest)
		if err != nil {
			log.Printf("error decode: %v", socketRequest)
			http.Error(w, "error decode JSON "+err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("After decode socket request: '%v', timeout:'%v'", socketRequest.Command, socketRequest.Timeout)
		connection := deviceShifuSocketHandlerMetaData.connection
		command := socketRequest.Command
		timeout := socketRequest.Timeout
		if timeout != 0 {
			(*connection).SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
		}

		log.Printf("Sending %v", []byte(command+"\n"))
		(*connection).Write([]byte(command + "\n"))
		message, err := bufio.NewReader(*connection).ReadString('\n')
		if err != nil {
			log.Printf("Failed to ReadString from Socket, current message: %v", message)
			http.Error(w, "Failed to read message from socket, error: "+err.Error(), http.StatusBadRequest)
		}

		returnMessage := DeviceShifuSocketReturnBody{
			Message: message,
			Status:  http.StatusOK,
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(returnMessage)
	}
}

// this function gathers the instruction name and its arguments from user input via HTTP and create the direct call command
// "flags_no_parameter" is a special key where it contains all flags
// e.g.:
// if we have localhost:8081/start?time=10:00:00&flags_no_parameter=-a,-c,--no-dependency&target=machine2
// and our driverExecution is "/usr/local/bin/python /usr/src/driver/python-car-driver.py"
// then we will get this command string:
// /usr/local/bin/python /usr/src/driver/python-car-driver.py --start time=10:00:00 target=machine2 -a -c --no-dependency
// which is exactly what we need to run if we are operating directly on the device
func createHTTPCommandlineRequestString(r *http.Request, driverExecution string, instruction string) string {
	values := r.URL.Query()
	requestStr := ""
	flagsStr := ""
	for parameterName, parameterValues := range values {
		if parameterName == "flags_no_parameter" {
			if len(parameterValues) == 1 {
				flagsStr = " " + strings.Replace(parameterValues[0], ",", " ", -1)
			} else {
				for _, parameterValue := range parameterValues {
					flagsStr += " " + parameterValue
				}
			}
		} else {
			if len(parameterValues) < 1 {
				continue
			}

			requestStr += " " + parameterName + "="
			for _, parameterValue := range parameterValues {
				requestStr += parameterValue
			}
		}
	}
	return driverExecution + " --" + instruction + requestStr + flagsStr
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
		log.Printf("Status is: %v", status)
		if err != nil {
			log.Printf("Error is: %v", err.Error())
			telemetryOK = false
		}

		if !status && telemetryOK {
			telemetryOK = false
		}
	}

	if telemetryOK {
		ds.updateEdgeDeviceResourcePhase(v1alpha1.EdgeDeviceRunning)
	} else {
		ds.updateEdgeDeviceResourcePhase(v1alpha1.EdgeDeviceFailed)
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
			ds.updateEdgeDeviceResourcePhase(v1alpha1.EdgeDeviceFailed)
		}

		return nil
	}

	return fmt.Errorf("EdgeDevice %v has no telemetry field in configuration\n", ds.Name)
}

func (ds *DeviceShifu) updateEdgeDeviceResourcePhase(edPhase v1alpha1.EdgeDevicePhase) {
	log.Printf("updating device %v status to: %v\n", ds.Name, edPhase)
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
	} else {
		*currEdgeDevice.Status.EdgeDevicePhase = edPhase
	}

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
