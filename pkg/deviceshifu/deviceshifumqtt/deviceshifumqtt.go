package deviceshifumqtt

import (

	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/edgenesis/shifu/pkg/deviceshifu/utils"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/edgenesis/shifu/pkg/logger"
)

// DeviceShifu implemented from deviceshifuBase
type DeviceShifu struct {
	base             *deviceshifubase.DeviceShifuBase
	mqttInstructions *MQTTInstructions
	state            *MQTTState
}

// MQTTState holds the state for the MQTT device connection and messages
type MQTTState struct {
	client                         mqtt.Client
	mqttMessageInstructionMap      map[string]string
	mqttMessageReceiveTimestampMap map[string]time.Time
	controlMsgs                    map[string]string // The key is controlMsg, the value is completion Msg returned by the device
	currentControlMsg              string
	mutexBlocking                  bool
	mu                             sync.RWMutex
}

// HandlerMetaData MetaData for EdgeDevice Setting
type HandlerMetaData struct {
	edgeDeviceSpec v1alpha1.EdgeDeviceSpec
	instruction    string
	properties     *MQTTProtocolProperty
}

// Str and default value
const (
	DefaultUpdateIntervalInMS int64 = 3000
)

// New new MQTT Deviceshifu
func New(deviceShifuMetadata *deviceshifubase.DeviceShifuMetaData) (*DeviceShifu, error) {
	base, mux, err := deviceshifubase.New(deviceShifuMetadata)
	if err != nil {
		return nil, err
	}

	mqttInstructions := CreateMQTTInstructions(&base.DeviceShifuConfig.Instructions)
	mqttState := &MQTTState{
		mqttMessageInstructionMap:      make(map[string]string),
		mqttMessageReceiveTimestampMap: make(map[string]time.Time),
		controlMsgs:                    make(map[string]string),
	}

	if deviceShifuMetadata.KubeConfigPath != deviceshifubase.DeviceKubeconfigDoNotLoadStr {
		switch protocol := *base.EdgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolMQTT:
			if base.DeviceShifuConfig.ControlMsgs != nil {
				mqttState.controlMsgs = base.DeviceShifuConfig.ControlMsgs
			}
			
			mqttProtocolSetting := base.EdgeDevice.Spec.ProtocolSettings
			if mqttProtocolSetting != nil {
				if mqttProtocolSetting.MQTTSetting != nil && mqttProtocolSetting.MQTTSetting.MQTTServerSecret != nil {
					logger.Infof("MQTT Server Secret is not empty, currently Shifu does not use MQTT Server Secret")
					// TODO Add MQTT Server secret processing logic
				}
			}

			opts := mqtt.NewClientOptions()
			opts.AddBroker(fmt.Sprintf("tcp://%s", *base.EdgeDevice.Spec.Address))
			opts.SetClientID(base.EdgeDevice.Name)
			opts.SetDefaultPublishHandler(mqttState.messagePubHandler)
			opts.OnConnect = connectHandler
			opts.OnConnectionLost = connectLostHandler
			client := mqtt.NewClient(opts)
			mqttState.client = client

			if token := client.Connect(); token.Wait() && token.Error() != nil {
				logger.Errorf("Error connecting to MQTT broker: %v", token.Error())
				// We don't panic here, but return error or let it retry? 
				// The original code panicked. Let's return error.
				// panic(token.Error()) 
				return nil, token.Error()
			}

			// Subscriptions
			for instruction, properties := range mqttInstructions.Instructions {
				topic := properties.MQTTProtocolProperty.MQTTTopic
				sub(client, topic, mqttState)

				HandlerMetaData := &HandlerMetaData{
					base.EdgeDevice.Spec,
					instruction,
					properties.MQTTProtocolProperty,
				}

				handler := DeviceCommandHandlerMQTT{
					HandlerMetaData: HandlerMetaData,
					state:           mqttState,
				}
				mux.HandleFunc("/"+instruction, handler.commandHandleFunc())
			}
		}
	}
	deviceshifubase.BindDefaultHandler(mux)

	ds := &DeviceShifu{
		base:             base,
		mqttInstructions: mqttInstructions,
		state:            mqttState,
	}

	ds.base.UpdateEdgeDeviceResourcePhase(v1alpha1.EdgeDevicePending)
	return ds, nil
}

func (s *MQTTState) messagePubHandler(client mqtt.Client, msg mqtt.Message) {
	logger.Infof("Received message: %v from topic: %v", msg.Payload(), msg.Topic())
	rawMqttMessageStr := string(msg.Payload())
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.mqttMessageInstructionMap[msg.Topic()] = rawMqttMessageStr
	s.mqttMessageReceiveTimestampMap[msg.Topic()] = time.Now()
	logger.Infof("MESSAGE_STR updated")
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	logger.Infof("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	logger.Infof("Connect lost: %v", err)
}

func sub(client mqtt.Client, topic string, s *MQTTState) {
	token := client.Subscribe(topic, 1, s.receiver)
	token.Wait()
	logger.Infof("Subscribed to topic: %s", topic)
}

func (s *MQTTState) receiver(client mqtt.Client, msg mqtt.Message) {
	msg.Ack()
	s.messagePubHandler(client, msg)
	message := string(msg.Payload())
	s.mutexProcess(msg.Topic(), message)
	logger.Infof("Received message:{id:%v, message:%v}", strconv.Itoa(int(msg.MessageID())), message)
}

func (s *MQTTState) mutexProcess(topic string, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.mutexBlocking && strings.Contains(message, s.controlMsgs[s.currentControlMsg]) {
		logger.Infof("Resetting mutex")
		s.mutexBlocking = false
		s.currentControlMsg = ""
	}
}

// DeviceCommandHandlerMQTT handler for Mqtt
type DeviceCommandHandlerMQTT struct {
	HandlerMetaData *HandlerMetaData
	state           *MQTTState
}

func (handler DeviceCommandHandlerMQTT) commandHandleFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqType := r.Method
		topic := handler.HandlerMetaData.properties.MQTTTopic
		
		if reqType == http.MethodGet {
			handler.state.mu.RLock()
			msg, exists := handler.state.mqttMessageInstructionMap[topic]
			ts := handler.state.mqttMessageReceiveTimestampMap[topic]
			handler.state.mu.RUnlock()
			
			if !exists {
				// Handle case where no message received yet
				// return empty or error? Original code would probably return empty string
			}

			returnMessage := ReturnBody{
				MQTTMessage:   msg,
				MQTTTimestamp: ts.String(),
			}

			responseMessage, err := json.Marshal(returnMessage)
			if err != nil {
				http.Error(w, "Cannot Encode message to json", http.StatusInternalServerError)
				logger.Errorf("Cannot Encode message to json")
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			instructionFuncName, shouldUsePythonCustomProcessing := deviceshifubase.CustomInstructionsPython[handler.HandlerMetaData.instruction]
			if shouldUsePythonCustomProcessing {
				logger.Infof("Topic %v has a python customized handler configured.", topic)
				responseMessage = []byte(utils.ProcessInstruction(deviceshifubase.PythonHandlersModuleName, instructionFuncName, string(responseMessage), deviceshifubase.PythonScriptDir))
				if !json.Valid(responseMessage) {
					w.Header().Set("Content-Type", "text/plain")
				}
			}

			_, err = w.Write(responseMessage)
			if err != nil {
				http.Error(w, "Cannot Encode message to json", http.StatusInternalServerError)
				logger.Errorf("Cannot Encode message to json")
				return
			}
		} else if reqType == http.MethodPost || reqType == http.MethodPut {
			mqttTopic := handler.HandlerMetaData.properties.MQTTTopic
			
			handler.state.mu.RLock()
			isBlocked := handler.state.mutexBlocking
			blockMsg := handler.state.currentControlMsg
			handler.state.mu.RUnlock()

			if isBlocked {
				blockedMessage := fmt.Sprintf("Device is blocked by %v controlMsg now! %v", blockMsg, time.Now())
				logger.Errorf(blockedMessage)
				http.Error(w, blockedMessage, http.StatusConflict)
				return
			}
			
			body, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Errorf("Error when Read Data From Body, error: %v", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			requestBody := RequestBody(body)
			logger.Infof("requestBody: %v", requestBody)

			token := handler.state.client.Publish(mqttTopic, 1, false, body)
			if token.Error() != nil {
				logger.Errorf("Error when publish Data to MQTTServer,%v", token.Error())
				http.Error(w, "Error to publish a message to server", http.StatusBadRequest)
				return
			}
			
			handler.state.mu.Lock()
			if _, isMutexState := handler.state.controlMsgs[string(requestBody)]; isMutexState {
				handler.state.mutexBlocking = true
				handler.state.currentControlMsg = string(requestBody)
				logger.Infof("Message %s is mutex, blocking.", requestBody)
			}
			handler.state.mu.Unlock()
			
			logger.Infof("Info: Success To publish a message %v to MQTTServer!", requestBody)
			return
		} else {
			http.Error(w, "must be GET or PUT method", http.StatusBadRequest)
			logger.Errorf("Request type %v is not supported yet!", reqType)
			return
		}

	}
}

func (ds *DeviceShifu) getMQTTTopicFromInstructionName(instructionName string) (string, error) {
	if instructionProperties, exists := ds.mqttInstructions.Instructions[instructionName]; exists {
		return instructionProperties.MQTTProtocolProperty.MQTTTopic, nil
	}

	return "", fmt.Errorf("Instruction %v not found in list of deviceshifu instructions", instructionName)
}

func (ds *DeviceShifu) collectMQTTTelemetry() (bool, error) {
	if ds.base.EdgeDevice.Spec.Protocol != nil {
		switch protocol := *ds.base.EdgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolMQTT:
			telemetrySettings := ds.base.DeviceShifuConfig.Telemetries.DeviceShifuTelemetrySettings
			if ds.base.EdgeDevice.Spec.Address == nil {
				return false, fmt.Errorf("device %v does not have an address", ds.base.Name)
			}

			if interval := telemetrySettings.DeviceShifuTelemetryUpdateIntervalInMilliseconds; interval == nil {
				var telemetryUpdateIntervalInMilliseconds = DefaultUpdateIntervalInMS
				telemetrySettings.DeviceShifuTelemetryUpdateIntervalInMilliseconds = &telemetryUpdateIntervalInMilliseconds
			}

			telemetries := ds.base.DeviceShifuConfig.Telemetries.DeviceShifuTelemetries
			for telemetry, telemetryProperties := range telemetries {
				if telemetryProperties.DeviceShifuTelemetryProperties.DeviceInstructionName == nil {
					return false, fmt.Errorf("Device %v telemetry %v does not have an instruction name", ds.base.Name, telemetry)
				}

				instruction := *telemetryProperties.DeviceShifuTelemetryProperties.DeviceInstructionName
				mqttTopic, err := ds.getMQTTTopicFromInstructionName(instruction)
				if err != nil {
					logger.Errorf("%v", err.Error())
					return false, err
				}

				ds.state.mu.RLock()
				lastTime, exists := ds.state.mqttMessageReceiveTimestampMap[mqttTopic]
				ds.state.mu.RUnlock()
				
				if !exists {
					// No message received yet, unsure if we should fail or wait. 
					// Assuming failure/disconnect if we rely on telemetry.
					return false, nil
				}

				nowTime := time.Now()
				if int64(nowTime.Sub(lastTime).Milliseconds()) < *telemetrySettings.DeviceShifuTelemetryUpdateIntervalInMilliseconds {
					return true, nil
				}
			}
		default:
			logger.Warnf("EdgeDevice protocol %v not supported in deviceshifu", protocol)
			return false, nil
		}
	}

	return false, nil
}

// Start start Mqtt Telemetry
func (ds *DeviceShifu) Start(stopCh <-chan struct{}) error {
	return ds.base.Start(stopCh, ds.collectMQTTTelemetry)
}

// Stop Stop Http Server
func (ds *DeviceShifu) Stop() error {
	return ds.base.Stop()
}
