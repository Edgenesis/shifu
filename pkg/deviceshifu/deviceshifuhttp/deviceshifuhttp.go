package deviceshifuhttp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"

	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

// DeviceShifuHTTP deviceshifu for HTTP
type DeviceShifuHTTP struct {
	base *deviceshifubase.DeviceShifuBase
}

// HandlerMetaData MetaData for HTTPhandler
type HandlerMetaData struct {
	edgeDeviceSpec v1alpha1.EdgeDeviceSpec
	instruction    string
	properties     *deviceshifubase.DeviceShifuInstruction
}

//CommandlineHandlerMetadata MetaData for HTTPCommandline handler
type CommandlineHandlerMetadata struct {
	edgeDeviceSpec  v1alpha1.EdgeDeviceSpec
	instruction     string
	properties      *deviceshifubase.DeviceShifuInstruction
	driverExecution string
}

var (
	instructionSettings *deviceshifubase.DeviceShifuInstructionSettings
)

//New This function creates a new Device Shifu based on the configuration
func New(deviceShifuMetadata *deviceshifubase.DeviceShifuMetaData) (*DeviceShifuHTTP, error) {
	if deviceShifuMetadata.Namespace == "" {
		return nil, fmt.Errorf("DeviceShifuHTTP's namespace can't be empty")
	}

	base, mux, err := deviceshifubase.New(deviceShifuMetadata)
	if err != nil {
		return nil, err
	}

	if deviceShifuMetadata.KubeConfigPath != deviceshifubase.DeviceKubeconfigDoNotLoadStr {

		instructionSettings = base.DeviceShifuConfig.Instructions.InstructionSettings
		if instructionSettings == nil {
			instructionSettings = &deviceshifubase.DeviceShifuInstructionSettings{}
		}

		if instructionSettings.DefaultTimeoutSeconds == nil {
			var defaultTimeoutSeconds = deviceshifubase.DeviceDefaultGolbalTimeoutInSeconds
			instructionSettings.DefaultTimeoutSeconds = &defaultTimeoutSeconds
		} else if *instructionSettings.DefaultTimeoutSeconds < 0 {
			klog.Fatalf("defaultTimeoutSeconds must not be negative number")
			return nil, errors.New("defaultTimeout configuration error")
		}

		// switch for different Shifu Protocols
		switch protocol := *base.EdgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolHTTP:
			for instruction, properties := range base.DeviceShifuConfig.Instructions.Instructions {
				HandlerMetaData := &HandlerMetaData{
					base.EdgeDevice.Spec,
					instruction,
					properties,
				}
				handler := DeviceCommandHandlerHTTP{base.RestClient, HandlerMetaData}
				mux.HandleFunc("/"+instruction, handler.commandHandleFunc())
			}
		case v1alpha1.ProtocolHTTPCommandline:
			driverExecution := base.DeviceShifuConfig.DriverProperties.DriverExecution
			if driverExecution == "" {
				return nil, fmt.Errorf("driverExecution cannot be empty")
			}

			for instruction, properties := range base.DeviceShifuConfig.Instructions.Instructions {
				CommandlineHandlerMetadata := &CommandlineHandlerMetadata{
					base.EdgeDevice.Spec,
					instruction,
					properties,
					base.DeviceShifuConfig.DriverProperties.DriverExecution,
				}

				handler := DeviceCommandHandlerHTTPCommandline{base.RestClient, CommandlineHandlerMetadata}
				mux.HandleFunc("/"+instruction, handler.commandHandleFunc())
			}
		default:
			klog.Errorf("EdgeDevice protocol %v not supported in deviceShifu_http_http", protocol)
			return nil, errors.New("wrong protocol not supported in deviceShifu_http_http")
		}
	}
	deviceshifubase.BindDefaultHandler(mux)

	ds := &DeviceShifuHTTP{base: base}

	if err := ds.base.ValidateTelemetryConfig(); err != nil {
		klog.Errorf("%v", err)
		return ds, err
	}

	ds.base.UpdateEdgeDeviceResourcePhase(v1alpha1.EdgeDevicePending)
	return ds, nil
}

// DeviceCommandHandlerHTTP handler for http
type DeviceCommandHandlerHTTP struct {
	client          *rest.RESTClient
	HandlerMetaData *HandlerMetaData
}

// This function is to create a URL containing directives from the requested URL
// e.g.:
// if we have http://localhost:8081/start?time=10:00:00&target=machine1&target=machine2
// and our address is http://localhost:8088 and instruction is start
// then we will get this URL string:
// http://localhost:8088/start?time=10:00:00&target=machine1&target=machine2
func createURIFromRequest(address string, handlerInstruction string, r *http.Request) string {

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

// This function executes the instruction by requesting the url returned by createURIFromRequest
func (handler DeviceCommandHandlerHTTP) commandHandleFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handlerProperties := handler.HandlerMetaData.properties
		handlerInstruction := handler.HandlerMetaData.instruction
		handlerEdgeDeviceSpec := handler.HandlerMetaData.edgeDeviceSpec
		handlerHTTPClient := handler.client.Client

		if handlerProperties != nil {
			// TODO: handle validation compile
			for _, instructionProperty := range handlerProperties.DeviceShifuInstructionProperties {
				klog.Infof("Properties of command: %v %v", handlerInstruction, instructionProperty)
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

		klog.Infof("handling instruction '%v' to '%v' with request type %v", handlerInstruction, *handlerEdgeDeviceSpec.Address, reqType)

		timeoutStr := r.URL.Query().Get(deviceshifubase.DevuceInstructionTimeoutURIQueryStr)
		if timeoutStr != "" {
			timeout, parseErr = strconv.Atoi(timeoutStr)
			if parseErr != nil {
				http.Error(w, parseErr.Error(), http.StatusBadRequest)
				klog.Errorf("timeout URI parsing error" + parseErr.Error())
				return
			}
		}

		switch reqType {
		case http.MethodPost:
			requestBody, parseErr = io.ReadAll(r.Body)
			if parseErr != nil {
				http.Error(w, "Error on parsing body", http.StatusBadRequest)
				klog.Errorf("Error on parsing body" + parseErr.Error())
				return
			}

			fallthrough
		case http.MethodGet:
			// for shifu.cloud timeout=0 is emptyomit
			// for hikivison's rtsp stream need never timeout
			if timeout <= 0 {
				ctx, cancel = context.WithCancel(context.Background())
			} else {
				ctx, cancel = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
			}

			defer cancel()
			httpURL := createURIFromRequest(*handlerEdgeDeviceSpec.Address, handlerInstruction, r)
			req, reqErr := http.NewRequestWithContext(ctx, reqType, httpURL, bytes.NewBuffer(requestBody))
			if reqErr != nil {
				http.Error(w, reqErr.Error(), http.StatusBadRequest)
				klog.Errorf("HTTP GET error" + reqErr.Error())
				return
			}

			deviceshifubase.CopyHeader(req.Header, r.Header)
			resp, httpErr = handlerHTTPClient.Do(req)
			if httpErr != nil {
				http.Error(w, httpErr.Error(), http.StatusServiceUnavailable)
				klog.Errorf("HTTP POST error" + httpErr.Error())
				return
			}
		default:
			http.Error(w, httpErr.Error(), http.StatusBadRequest)
			klog.Errorf("Request type %v is not supported yet!", reqType)
			return
		}

		if resp != nil {
			deviceshifubase.CopyHeader(w.Header(), resp.Header)
			w.WriteHeader(resp.StatusCode)
			_, err := io.Copy(w, resp.Body)
			if err != nil {
				klog.Errorf("error when copy requestBody from responseBody, err: %v", err)
			}
			return
		}

		// TODO: For now, just write tht instruction to the response
		klog.Warningf("resp is nil")
		_, err := w.Write([]byte(handlerInstruction))
		if err != nil {
			klog.Errorf("cannot write instruction into response's body, err: %v", err)
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

// DeviceCommandHandlerHTTPCommandline handler for http commandline
type DeviceCommandHandlerHTTPCommandline struct {
	client                     *rest.RESTClient
	CommandlineHandlerMetadata *CommandlineHandlerMetadata
}

func (handler DeviceCommandHandlerHTTPCommandline) commandHandleFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		driverExecution := handler.CommandlineHandlerMetadata.driverExecution
		handlerProperties := handler.CommandlineHandlerMetadata.properties
		handlerInstruction := handler.CommandlineHandlerMetadata.instruction
		handlerEdgeDeviceSpec := handler.CommandlineHandlerMetadata.edgeDeviceSpec
		handlerHTTPClient := handler.client.Client

		if handlerProperties != nil {
			// TODO: handle validation compile
			for _, instructionProperty := range handlerProperties.DeviceShifuInstructionProperties {
				klog.Infof("Properties of command: %v %v", handlerInstruction, instructionProperty)
			}
		}

		klog.Infof("handling instruction '%v' to '%v'", handlerInstruction, *handlerEdgeDeviceSpec.Address)

		commandString := createHTTPCommandlineRequestString(r, driverExecution, handlerInstruction)
		postAddressString := "http://" + *handlerEdgeDeviceSpec.Address + "/post"
		klog.Infof("posting '%v' to '%v'", commandString, postAddressString)
		resp, err := handlerHTTPClient.Post(postAddressString, "text/plain", bytes.NewBuffer([]byte(commandString)))

		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			klog.Errorf("HTTP error" + err.Error())
		}

		if resp != nil {
			deviceshifubase.CopyHeader(w.Header(), resp.Header)
			w.WriteHeader(resp.StatusCode)
			_, err := io.Copy(w, resp.Body)
			if err != nil {
				klog.Errorf("cannot copy requestBody from requestBody, error: %v", err)
			}
			return
		}

		// TODO: For now, just write tht instruction to the response
		klog.Warningf("resp is nil")
		_, err = w.Write([]byte(handlerInstruction))
		if err != nil {
			klog.Errorf("cannot write instruction into responseBody")
		}
	}
}

// TODO: update configs

func (ds *DeviceShifuHTTP) collectHTTPTelemtries() (bool, error) {
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
					klog.Errorf("error checking telemetry: %v, error: %v", telemetry, ReqErr.Error())
					return false, ReqErr
				}

				resp, err := ds.base.RestClient.Client.Do(req)
				if err != nil {
					klog.Errorf("error checking telemetry: %v, error: %v", telemetry, err.Error())
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
			klog.Warningf("EdgeDevice protocol %v not supported in deviceshif", protocol)
			return false, nil
		}
	}

	return telemetryCollectionResult, nil
}

// Start start http telemetry
func (ds *DeviceShifuHTTP) Start(stopCh <-chan struct{}) error {
	return ds.base.Start(stopCh, ds.collectHTTPTelemtries)
}

// Stop stop http server
func (ds *DeviceShifuHTTP) Stop() error {
	return ds.base.Stop()
}
