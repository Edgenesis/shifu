// TODO: need to test for deviceshifu LwM2M
package deviceshifulwm2m

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifulwm2m/lwm2m"
	"github.com/edgenesis/shifu/pkg/deviceshifu/utils"
	"github.com/edgenesis/shifu/pkg/logger"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
)

// DeviceShifuLwM2M deviceshifu for LwM2M
type DeviceShifuLwM2M struct {
	// lwm2m server which device will connect to, it will be used to send command to device
	server              *lwm2m.Server
	base                *deviceshifubase.DeviceShifuBase
	instructionSettings *deviceshifubase.DeviceShifuInstructionSettings
}

// HandlerMetaData MetaData for HTTPhandler
type HandlerMetaData struct {
	edgeDeviceSpec v1alpha1.EdgeDeviceSpec
	instruction    string
	properties     *LwM2MProtocolProperty
}

// New This function creates a new Device Shifu based on the configuration
func New(deviceShifuMetadata *deviceshifubase.DeviceShifuMetaData) (*DeviceShifuLwM2M, error) {
	if deviceShifuMetadata.Namespace == "" {
		return nil, fmt.Errorf("DeviceShifuLwM2M's namespace can't be empty")
	}

	base, mux, err := deviceshifubase.New(deviceShifuMetadata)
	if err != nil {
		return nil, err
	}

	lwM2MSetting := base.EdgeDevice.Spec.ProtocolSettings.LwM2MSetting
	logger.Debugf("LwM2M endpoint is: %s", lwM2MSetting.EndpointName)
	server, err := lwm2m.NewServer(*lwM2MSetting)
	if err != nil {
		logger.Errorf("Error creating LwM2M server: %v", err)
		return nil, err
	}

	ds := &DeviceShifuLwM2M{base: base, server: server}

	go func() {
		if err := server.Run(); err != nil {
			logger.Fatalf("Error starting LwM2M server: %v", err)
		}
	}()

	ds.instructionSettings = base.DeviceShifuConfig.Instructions.InstructionSettings
	if ds.instructionSettings == nil {
		ds.instructionSettings = &deviceshifubase.DeviceShifuInstructionSettings{}
	}

	if ds.instructionSettings.DefaultTimeoutSeconds == nil {
		var defaultTimeoutSeconds = deviceshifubase.DeviceDefaultGlobalTimeoutInSeconds
		ds.instructionSettings.DefaultTimeoutSeconds = &defaultTimeoutSeconds
	}

	lwM2MInstructions := CreateLwM2MInstructions(&base.DeviceShifuConfig.Instructions)

	if deviceShifuMetadata.KubeConfigPath != deviceshifubase.DeviceKubeconfigDoNotLoadStr {
		// switch for different Shifu Protocols
		switch protocol := *base.EdgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolLwM2M:
			for instruction, properties := range lwM2MInstructions.Instructions {
				HandlerMetaData := &HandlerMetaData{
					base.EdgeDevice.Spec,
					instruction,
					properties,
				}
				handler := DeviceCommandHandlerLwM2M{server, HandlerMetaData}
				mux.HandleFunc("/"+instruction, handler.commandHandleFunc())

				if properties.EnableObserve {
					server.OnRegister(func() error {
						return server.Observe(properties.ObjectId, func(data interface{}) {
							logger.Debugf("Observe data: %v", data)
							// TODO need to push data to telemetry service
						})
					})
				}
			}
		default:
			logger.Errorf("EdgeDevice protocol %v not supported in deviceShifu LwM2M", protocol)
			return nil, errors.New("wrong protocol not supported in deviceShifu LwM2M")
		}
	}
	deviceshifubase.BindDefaultHandler(mux)

	ds.base.UpdateEdgeDeviceResourcePhase(v1alpha1.EdgeDevicePending)
	return ds, nil
}

// DeviceCommandHandlerLwM2M handler for http
type DeviceCommandHandlerLwM2M struct {
	server          *lwm2m.Server
	HandlerMetaData *HandlerMetaData
}

// commandHandleFunc handle http request
func (handler DeviceCommandHandlerLwM2M) commandHandleFunc() http.HandlerFunc {
	handlerProperties := handler.HandlerMetaData.properties
	handlerInstruction := handler.HandlerMetaData.instruction
	handlerServer := handler.server

	objectId := handlerProperties.ObjectId
	return func(w http.ResponseWriter, r *http.Request) {
		var respString string

		switch r.Method {
		case http.MethodPut:
			requestBody, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Error on parsing body", http.StatusBadRequest)
				logger.Errorf("Error on parsing body" + err.Error())
				return
			}

			err = handlerServer.Write(objectId, string(requestBody))
			if err != nil {
				http.Error(w, "Error on writing object", http.StatusBadRequest)
				logger.Errorf("Error on writing object" + err.Error())
				return
			}

			respString = "Success"
		case http.MethodGet:
			data, err := handlerServer.Read(objectId)
			if err != nil {
				http.Error(w, "Error on reading object", http.StatusBadRequest)
				logger.Errorf("Error on reading object" + err.Error())
				return
			}
			respString = data
		case http.MethodPost:
			requestBody, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Error on parsing body", http.StatusBadRequest)
				logger.Errorf("Error on parsing body" + err.Error())
				return
			}

			err = handlerServer.Execute(objectId, string(requestBody))
			if err != nil {
				http.Error(w, "Error on executing object", http.StatusBadRequest)
				logger.Errorf("Error on executing object" + err.Error())
				return
			}

			respString = "Success"
		default:
			http.Error(w, "request method not support", http.StatusMethodNotAllowed)
			logger.Errorf("Request method %s is not support!", r.Method)
			return
		}

		instructionFuncName, shouldUsePythonCustomProcessing := deviceshifubase.CustomInstructionsPython[handlerInstruction]
		logger.Debugf("Instruction %v is custom: %v", handlerInstruction, shouldUsePythonCustomProcessing)
		if shouldUsePythonCustomProcessing {
			logger.Debugf("Instruction %v has a python customized handler configured.", handlerInstruction)
			respString = utils.ProcessInstruction(deviceshifubase.PythonHandlersModuleName, instructionFuncName, respString, deviceshifubase.PythonScriptDir)
		}
		_, _ = fmt.Fprintf(w, "%v", respString)
	}
}

func (ds *DeviceShifuLwM2M) collectHTTPTelemtries() (bool, error) {
	if ds.server.Conn == nil {
		return false, nil
	}

	if err := ds.server.Conn.Ping(ds.server.Conn.Context()); err != nil {
		logger.Errorf("cannot ping the device, please check device is online, error: %v", err.Error())
		return false, err
	}

	if ds.base.EdgeDevice.Spec.Protocol != nil {
		switch protocol := *ds.base.EdgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolLwM2M:
			telemetries := ds.base.DeviceShifuConfig.Telemetries.DeviceShifuTelemetries
			instructions := ds.base.DeviceShifuConfig.Instructions
			for telemetry, telemetryProperties := range telemetries {
				if ds.base.EdgeDevice.Spec.Address == nil {
					return false, fmt.Errorf("device %v does not have an address", ds.base.Name)
				}

				instructionName := telemetryProperties.DeviceShifuTelemetryProperties.DeviceInstructionName
				if instructionName == nil {
					return false, fmt.Errorf("device %v telemetry %v does not have an instruction name", ds.base.Name, telemetry)
				}

				if instructions.Instructions[*instructionName] == nil {
					return false, fmt.Errorf("device %v telemetry %v instruction %v does not exist", ds.base.Name, telemetry, *instructionName)
				}

				objectId := instructions.Instructions[*instructionName].DeviceShifuProtocolProperties["ObjectId"]
				if objectId == "" {
					return false, fmt.Errorf("device %v telemetry %v does not have an object id", ds.base.Name, telemetry)
				}
				data, err := ds.server.Read(objectId)
				if err != nil {
					return false, err
				}

				resp := &http.Response{
					Body: io.NopCloser(strings.NewReader(data)),
				}

				telemetryCollectionService, exist := deviceshifubase.TelemetryCollectionServiceMap[telemetry]
				if exist && *telemetryCollectionService.TelemetryServiceEndpoint != "" {
					err = deviceshifubase.PushTelemetryCollectionService(&telemetryCollectionService, &ds.base.EdgeDevice.Spec, resp)
					if err != nil {
						return false, err
					}
				}
				return true, nil
			}
		default:
			logger.Warnf("EdgeDevice protocol %v not supported in deviceshifu", protocol)
			return false, nil
		}
	}

	return true, nil
}

// Start start http telemetry
func (ds *DeviceShifuLwM2M) Start(stopCh <-chan struct{}) error {
	return ds.base.Start(stopCh, ds.collectHTTPTelemtries)
}

// Stop stop http server
func (ds *DeviceShifuLwM2M) Stop() error {
	return ds.base.Stop()
}
