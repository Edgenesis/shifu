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

	"github.com/edgenesis/shifu/deviceshifu/pkg/deviceshifubase"

	"edgenesis.io/shifu/k8s/crd/api/v1alpha1"
	"k8s.io/client-go/rest"
)

type DeviceShifu struct {
	base *deviceshifubase.DeviceShifuBase
}

type DeviceShifuHTTPHandlerMetaData struct {
	edgeDeviceSpec v1alpha1.EdgeDeviceSpec
	instruction    string
	properties     *deviceshifubase.DeviceShifuInstruction
}

type DeviceShifuHTTPCommandlineHandlerMetadata struct {
	edgeDeviceSpec  v1alpha1.EdgeDeviceSpec
	instruction     string
	properties      *deviceshifubase.DeviceShifuInstruction
	driverExecution string
}

type deviceCommandHandler interface {
	commandHandleFunc(w http.ResponseWriter, r *http.Request) http.HandlerFunc
}

var (
	instructionSettings *deviceshifubase.DeviceShifuInstructionSettings
)

// This function creates a new Device Shifu based on the configuration
func New(deviceShifuMetadata *deviceshifubase.DeviceShifuMetaData) (*DeviceShifu, error) {
	if deviceShifuMetadata.Namespace == "" {
		return nil, fmt.Errorf("DeviceShifu's namespace can't be empty\n")
	}

	base, mux, err := deviceshifubase.New(deviceShifuMetadata)
	if err != nil {
		return nil, err
	}

	if deviceShifuMetadata.KubeConfigPath != deviceshifubase.DEVICE_KUBECONFIG_DO_NOT_LOAD_STR {

		instructionSettings = base.DeviceShifuConfig.Instructions.InstructionSettings
		if instructionSettings == nil {
			instructionSettings = &deviceshifubase.DeviceShifuInstructionSettings{}
		}

		if instructionSettings.DefaultTimeoutSeconds == nil {
			var defaultTimeoutSeconds = deviceshifubase.DEVICE_DEFAULT_GLOBAL_TIMEOUT_SECONDS
			instructionSettings.DefaultTimeoutSeconds = &defaultTimeoutSeconds
		} else if *instructionSettings.DefaultTimeoutSeconds < 0 {
			log.Fatalf("defaultTimeoutSeconds must not be negative number")
			return nil, errors.New("defaultTimeout configuration error")
		}

		// switch for different Shifu Protocols
		switch protocol := *base.EdgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolHTTP:
			for instruction, properties := range base.DeviceShifuConfig.Instructions.Instructions {
				deviceShifuHTTPHandlerMetaData := &DeviceShifuHTTPHandlerMetaData{
					base.EdgeDevice.Spec,
					instruction,
					properties,
				}
				handler := DeviceCommandHandlerHTTP{base.RestClient, deviceShifuHTTPHandlerMetaData}
				mux.HandleFunc("/"+instruction, handler.commandHandleFunc())
			}
		case v1alpha1.ProtocolHTTPCommandline:
			driverExecution := base.DeviceShifuConfig.DriverProperties.DriverExecution
			if driverExecution == "" {
				return nil, fmt.Errorf("driverExecution cannot be empty")
			}

			for instruction, properties := range base.DeviceShifuConfig.Instructions.Instructions {
				deviceShifuHTTPCommandlineHandlerMetaData := &DeviceShifuHTTPCommandlineHandlerMetadata{
					base.EdgeDevice.Spec,
					instruction,
					properties,
					base.DeviceShifuConfig.DriverProperties.DriverExecution,
				}

				handler := DeviceCommandHandlerHTTPCommandline{base.RestClient, deviceShifuHTTPCommandlineHandlerMetaData}
				mux.HandleFunc("/"+instruction, handler.commandHandleFunc())
			}
		default:
			log.Printf("EdgeDevice protocol %v not supported in deviceShifu_http_http\n", protocol)
			return nil, errors.New("wrong protocol not supported in deviceShifu_http_http")
		}
	}

	ds := &DeviceShifu{base: base}

	if err := ds.base.ValidateTelemetryConfig(); err != nil {
		log.Println(err)
		return ds, err
	}

	ds.base.UpdateEdgeDeviceResourcePhase(v1alpha1.EdgeDevicePending)
	return ds, nil
}

// deviceHealthHandler writes the status as healthy
func deviceHealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, deviceshifubase.DEVICE_IS_HEALTHY_STR)
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

		timeoutStr := r.URL.Query().Get(deviceshifubase.DEVICE_INSTRUCTION_TIMEOUT_URI_QUERY_STR)
		if timeoutStr != "" {
			timeout, parseErr = strconv.Atoi(timeoutStr)
			if parseErr != nil {
				http.Error(w, parseErr.Error(), http.StatusBadRequest)
				log.Printf("timeout URI parsing error" + parseErr.Error())
				return
			}
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

			deviceshifubase.CopyHeader(req.Header, r.Header)
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
			deviceshifubase.CopyHeader(w.Header(), resp.Header)
			w.WriteHeader(resp.StatusCode)
			io.Copy(w, resp.Body)
			return
		}

		// TODO: For now, just write tht instruction to the response
		log.Println("resp is nil")
		w.Write([]byte(handlerInstruction))
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
			deviceshifubase.CopyHeader(w.Header(), resp.Header)
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
	fmt.Printf("deviceShifu %s's http server started\n", ds.base.Name)
	return ds.base.Server.ListenAndServe()
}

// TODO: update configs

func (ds *DeviceShifu) collectHTTPTelemtries() (bool, error) {
	telemetryCollectionResult := false
	if ds.base.EdgeDevice.Spec.Protocol != nil {
		switch protocol := *ds.base.EdgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolHTTP:
			telemetries := ds.base.DeviceShifuConfig.Telemetries.DeviceShifuTelemetries
			for telemetry, telemetryProperties := range telemetries {
				if ds.base.EdgeDevice.Spec.Address == nil {
					return false, fmt.Errorf("Device %v does not have an address", ds.base.Name)
				}

				if telemetryProperties.DeviceShifuTelemetryProperties.DeviceInstructionName == nil {
					return false, fmt.Errorf("Device %v telemetry %v does not have an instruction name", ds.base.Name, telemetry)
				}

				var (
					ctx     context.Context
					cancel  context.CancelFunc
					timeout = *ds.base.DeviceShifuConfig.Telemetries.DeviceShifuTelemetrySettings.DeviceShifuTelemetryTimeoutInMilliseconds
				)

				if timeout == 0 {
					ctx, cancel = context.WithCancel(context.TODO())
				} else {
					ctx, cancel = context.WithTimeout(context.TODO(), time.Duration(timeout)*time.Millisecond)
				}

				defer cancel()
				address := *ds.base.EdgeDevice.Spec.Address
				instruction := *telemetryProperties.DeviceShifuTelemetryProperties.DeviceInstructionName
				req, ReqErr := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+address+"/"+instruction, nil)
				if ReqErr != nil {
					log.Printf("error checking telemetry: %v, error: %v", telemetry, ReqErr.Error())
					return false, ReqErr
				}

				resp, err := ds.base.RestClient.Client.Do(req)
				if err != nil {
					log.Printf("error checking telemetry: %v, error: %v", telemetry, err.Error())
					return false, err
				}

				if resp != nil {
					if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
						if telemetryCollectionService, exist := deviceshifubase.TelemetryCollectionServiceMap[telemetry]; exist {
							deviceshifubase.PushToHTTPTelemetryCollectionService(protocol, resp, telemetryCollectionService)
						}

						telemetryCollectionResult = true
						continue
					}
				}

				return false, nil
			}
		default:
			log.Printf("EdgeDevice protocol %v not supported in deviceShifu\n", protocol)
			return false, nil
		}
	}

	return telemetryCollectionResult, nil
}

func (ds *DeviceShifu) Start(stopCh <-chan struct{}) error {
	return ds.base.Start(stopCh, ds.collectHTTPTelemtries)
}

func (ds *DeviceShifu) Stop() error {
	return ds.base.Stop()
}
