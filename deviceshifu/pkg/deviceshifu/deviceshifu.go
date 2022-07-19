package deviceshifu

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"edgenesis.io/shifu/k8s/crd/api/v1alpha1"
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
	edgeDeviceSpec v1alpha1.EdgeDeviceSpec
	instruction    string
	properties     *DeviceShifuInstruction
}

type DeviceShifuHTTPCommandlineHandlerMetadata struct {
	edgeDeviceSpec  v1alpha1.EdgeDeviceSpec
	instruction     string
	properties      *DeviceShifuInstruction
	driverExecution string
}

type deviceCommandHandler interface {
	commandHandleFunc(w http.ResponseWriter, r *http.Request) http.HandlerFunc
}

const (
	DEVICE_IS_HEALTHY_STR                    string = "Device is healthy"
	DEVICE_CONFIGMAP_FOLDER_PATH             string = "/etc/edgedevice/config"
	DEVICE_KUBECONFIG_DO_NOT_LOAD_STR        string = "NULL"
	DEVICE_NAMESPACE_DEFAULT                 string = "default"
	DEVICE_DEFAULT_PORT_STR                  string = ":8080"
	KUBERNETES_CONFIG_DEFAULT                string = ""
	DEVICE_INSTRUCTION_TIMEOUT_URI_QUERY_STR string = "timeout"
	DEVICE_DEFAULT_GLOBAL_TIMEOUT_SECONDS    int    = 3
	DEVICE_TELEMETRY_TIMEOUT_MS              int64  = 3000
	DEVICE_TELEMETRY_UPDATE_INTERVAL_MS      int64  = 3000
	DEVICE_TELEMETRY_INITIAL_DELAY_MS        int64  = 3000
)

var (
	instructionSettings *DeviceShifuInstructionSettings
)

// This function creates a new Device Shifu based on the configuration
func New(deviceShifuMetadata *DeviceShifuMetaData) (*DeviceShifu, error) {
	if deviceShifuMetadata.Name == "" {
		return nil, fmt.Errorf("DeviceShifu's name can't be empty\n")
	}

	if deviceShifuMetadata.Namespace == "" {
		return nil, fmt.Errorf("DeviceShifu's namespace can't be empty\n")
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

		instructionSettings = deviceShifuConfig.Instructions.InstructionSettings
		if instructionSettings == nil {
			instructionSettings = &DeviceShifuInstructionSettings{}
		}

		if instructionSettings.DefaultTimeoutSeconds == nil {
			var defaultTimeoutSeconds = DEVICE_DEFAULT_GLOBAL_TIMEOUT_SECONDS
			instructionSettings.DefaultTimeoutSeconds = &defaultTimeoutSeconds
		} else if *instructionSettings.DefaultTimeoutSeconds < 0 {
			log.Fatalf("defaultTimeoutSeconds must not be negative number")
			return nil, errors.New("defaultTimeout configuration error")
		}

		// switch for different Shifu Protocols
		switch protocol := *edgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolHTTP:
			for instruction, properties := range deviceShifuConfig.Instructions.Instructions {
				deviceShifuHTTPHandlerMetaData := &DeviceShifuHTTPHandlerMetaData{
					edgeDevice.Spec,
					instruction,
					properties,
				}
				handler := DeviceCommandHandlerHTTP{client, deviceShifuHTTPHandlerMetaData}
				mux.HandleFunc("/"+instruction, handler.commandHandleFunc())
			}
		case v1alpha1.ProtocolHTTPCommandline:
			driverExecution := deviceShifuConfig.driverProperties.DriverExecution
			if driverExecution == "" {
				return nil, fmt.Errorf("driverExecution cannot be empty")
			}

			for instruction, properties := range deviceShifuConfig.Instructions.Instructions {
				deviceShifuHTTPCommandlineHandlerMetaData := &DeviceShifuHTTPCommandlineHandlerMetadata{
					edgeDevice.Spec,
					instruction,
					properties,
					deviceShifuConfig.driverProperties.DriverExecution,
				}

				handler := DeviceCommandHandlerHTTPCommandline{client, deviceShifuHTTPCommandlineHandlerMetaData}
				mux.HandleFunc("/"+instruction, handler.commandHandleFunc())
			}
		default:
			log.Printf("EdgeDevice protocol %v not supported in deviceShifu_http_http\n", protocol)
			return nil, errors.New("wrong protocol not supported in deviceShifu_http_http")
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

	if err := ds.ValidateTelemetryConfig(); err != nil {
		log.Println(err)
		return ds, err
	}

	ds.updateEdgeDeviceResourcePhase(v1alpha1.EdgeDevicePending)
	return ds, nil
}

// deviceHealthHandler writes the status as healthy
func deviceHealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, DEVICE_IS_HEALTHY_STR)
}

type DeviceCommandHandlerHTTP struct {
	client                         *rest.RESTClient
	deviceShifuHTTPHandlerMetaData *DeviceShifuHTTPHandlerMetaData
}

func instructionNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Error: Device instruction does not exist!")
	http.Error(w, "Error: Device instruction does not exist!", http.StatusNotFound)
}

// This function is to create a URL containing directives from the requested URL
// e.g.:
// if we have http://localhost:8081/start?time=10:00:00&target=machine1&target=machine2
// and our address is http://localhost:8088 and instruction is start
// then we will get this URL string:
// http://localhost:8088/start?time=10:00:00&target=machine1&target=machine2
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

// This function executes the instruction by requesting the url returned by createUriFromRequest
func (handler DeviceCommandHandlerHTTP) commandHandleFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handlerProperties := handler.deviceShifuHTTPHandlerMetaData.properties
		handlerInstruction := handler.deviceShifuHTTPHandlerMetaData.instruction
		handlerEdgeDeviceSpec := handler.deviceShifuHTTPHandlerMetaData.edgeDeviceSpec
		handlerHTTPClient := handler.client.Client

		if handlerProperties != nil {
			// TODO: handle validation compile
			for _, instructionProperty := range handlerProperties.DeviceShifuInstructionProperties {
				log.Printf("Properties of command: %v %v\n", handlerInstruction, instructionProperty)
			}
		}

		var (
			resp              *http.Response
			httpErr, parseErr error
			requestBody       []byte
			ctx               context.Context
			cancel            context.CancelFunc
			timeout           = *instructionSettings.DefaultTimeoutSeconds
			reqType           = r.Method
		)

		log.Printf("handling instruction '%v' to '%v' with request type %v", handlerInstruction, *handlerEdgeDeviceSpec.Address, reqType)

		timeoutStr := r.URL.Query().Get(DEVICE_INSTRUCTION_TIMEOUT_URI_QUERY_STR)
		if timeoutStr != "" {
			timeout, parseErr = strconv.Atoi(timeoutStr)
			if parseErr != nil {
				http.Error(w, parseErr.Error(), http.StatusBadRequest)
				log.Printf("timeout URI parsing error" + parseErr.Error())
				return
			}

			r.URL.Query().Del(DEVICE_INSTRUCTION_TIMEOUT_URI_QUERY_STR)
		}

		switch reqType {
		case http.MethodPost:
			requestBody, parseErr = io.ReadAll(r.Body)
			if parseErr != nil {
				http.Error(w, "Error on parsing body", http.StatusBadRequest)
				log.Printf("Error on parsing body" + parseErr.Error())
				return
			}

			fallthrough
		case http.MethodGet:
			if timeout == 0 {
				ctx, cancel = context.WithCancel(context.Background())
			} else {
				ctx, cancel = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
			}

			defer cancel()
			httpUri := createUriFromRequest(*handlerEdgeDeviceSpec.Address, handlerInstruction, r)
			req, reqErr := http.NewRequestWithContext(ctx, reqType, httpUri, bytes.NewBuffer(requestBody))
			if reqErr != nil {
				http.Error(w, reqErr.Error(), http.StatusBadRequest)
				log.Printf("HTTP GET error" + reqErr.Error())
				return
			}

			copyHeader(req.Header, r.Header)
			resp, httpErr = handlerHTTPClient.Do(req)
			if httpErr != nil {
				http.Error(w, httpErr.Error(), http.StatusServiceUnavailable)
				log.Printf("HTTP POST error" + httpErr.Error())
				return
			}
		default:
			http.Error(w, httpErr.Error(), http.StatusBadRequest)
			log.Println("Request type " + reqType + " is not supported yet!")
			return
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

type DeviceCommandHandlerHTTPCommandline struct {
	client                                    *rest.RESTClient
	deviceShifuHTTPCommandlineHandlerMetadata *DeviceShifuHTTPCommandlineHandlerMetadata
}

func (handler DeviceCommandHandlerHTTPCommandline) commandHandleFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		driverExecution := handler.deviceShifuHTTPCommandlineHandlerMetadata.driverExecution
		handlerProperties := handler.deviceShifuHTTPCommandlineHandlerMetadata.properties
		handlerInstruction := handler.deviceShifuHTTPCommandlineHandlerMetadata.instruction
		handlerEdgeDeviceSpec := handler.deviceShifuHTTPCommandlineHandlerMetadata.edgeDeviceSpec
		handlerHTTPClient := handler.client.Client

		if handlerProperties != nil {
			// TODO: handle validation compile
			for _, instructionProperty := range handlerProperties.DeviceShifuInstructionProperties {
				log.Printf("Properties of command: %v %v\n", handlerInstruction, instructionProperty)
			}
		}

		log.Printf("handling instruction '%v' to '%v'", handlerInstruction, *handlerEdgeDeviceSpec.Address)

		commandString := createHTTPCommandlineRequestString(r, driverExecution, handlerInstruction)
		postAddressString := "http://" + *handlerEdgeDeviceSpec.Address + "/post"
		log.Printf("posting '%v' to '%v'", commandString, postAddressString)
		resp, err := handlerHTTPClient.Post(postAddressString, "text/plain", bytes.NewBuffer([]byte(commandString)))

		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			log.Printf("HTTP error" + err.Error())
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

	var (
		ctx     context.Context
		cancel  context.CancelFunc
		timeout = *ds.deviceShifuConfig.Telemetries.DeviceShifuTelemetrySettings.DeviceShifuTelemetryTimeoutInMilliseconds
	)

	if timeout == 0 {
		ctx, cancel = context.WithCancel(context.TODO())
	} else {
		ctx, cancel = context.WithTimeout(context.TODO(), time.Duration(timeout)*time.Millisecond)
	}

	defer cancel()
	address := *ds.edgeDevice.Spec.Address
	instruction := *telemetryProperties.DeviceInstructionName
	req, ReqErr := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+address+"/"+instruction, nil)
	if ReqErr != nil {
		log.Printf("error checking telemetry: %v, error: %v", telemetry, ReqErr.Error())
		return false, ReqErr
	}

	resp, err := ds.restClient.Client.Do(req)
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
	log.Printf("deviceShifu %s's telemetry collection started\n", ds.Name)

	telemetryOK := true
	telemetries := ds.deviceShifuConfig.Telemetries
	for telemetry, telemetryProperties := range telemetries.DeviceShifuTelemetries {
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
	var telemetrySettings = ds.deviceShifuConfig.Telemetries.DeviceShifuTelemetrySettings
	log.Println("Waiting before updating status")
	time.Sleep(time.Duration(*telemetrySettings.DeviceShifuTelemetryInitialDelayInMilliseconds) * time.Millisecond)

	for {
		ds.collectHTTPTelemetries()
		time.Sleep(time.Duration(*telemetrySettings.DeviceShifuTelemetryUpdateIntervalInMilliseconds) * time.Millisecond)
	}
}

func (ds *DeviceShifu) Start(stopCh <-chan struct{}) error {
	log.Printf("deviceShifu %s started\n", ds.Name)

	go ds.startHttpServer(stopCh)
	go ds.StartTelemetryCollection()

	return nil
}

func (ds *DeviceShifu) Stop() error {
	if err := ds.server.Shutdown(context.TODO()); err != nil {
		return err
	}

	log.Printf("deviceShifu %s's http server stopped\n", ds.Name)
	return nil
}
