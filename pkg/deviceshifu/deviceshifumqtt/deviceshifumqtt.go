package deviceshifumqtt

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
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

var (
	client                         mqtt.Client
	MQTTTopic                      string
	mqttMessageInstructionMap      = map[string]string{}
	mqttMessageReceiveTimestampMap = map[string]time.Time{}
	mutexBlocking                  bool
	controlMsgs                    map[string]string // The key is controlMsg, the value is completion Msg returned by the device
	currentControlMsg              string
)

// New new MQTT Deviceshifu
func New(deviceShifuMetadata *deviceshifubase.DeviceShifuMetaData) (*DeviceShifu, error) {
	base, mux, err := deviceshifubase.New(deviceShifuMetadata)
	if err != nil {
		return nil, err
	}

	mqttInstructions := CreateMQTTInstructions(&base.DeviceShifuConfig.Instructions)

	if deviceShifuMetadata.KubeConfigPath != deviceshifubase.DeviceKubeconfigDoNotLoadStr {
		switch protocol := *base.EdgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolMQTT:
			ConfigFiniteStateMachine(base.DeviceShifuConfig.ControlMsgs)
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
			opts.SetDefaultPublishHandler(messagePubHandler)
			opts.OnConnect = connectHandler
			opts.OnConnectionLost = connectLostHandler
			client = mqtt.NewClient(opts)
			if token := client.Connect(); token.Wait() && token.Error() != nil {
				panic(token.Error())
			}

			for instruction, properties := range mqttInstructions.Instructions {
				MQTTTopic = properties.MQTTProtocolProperty.MQTTTopic
				sub(client, MQTTTopic)

				HandlerMetaData := &HandlerMetaData{
					base.EdgeDevice.Spec,
					instruction,
					properties.MQTTProtocolProperty,
				}

				handler := DeviceCommandHandlerMQTT{HandlerMetaData}
				mux.HandleFunc("/"+instruction, handler.commandHandleFunc())
			}
		}
	}
	deviceshifubase.BindDefaultHandler(mux)

	ds := &DeviceShifu{
		base:             base,
		mqttInstructions: mqttInstructions,
	}

	ds.base.UpdateEdgeDeviceResourcePhase(v1alpha1.EdgeDevicePending)
	return ds, nil
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	logger.Infof("Received message: %v from topic: %v", msg.Payload(), msg.Topic())
	rawMqttMessageStr := string(msg.Payload())
	mqttMessageInstructionMap[msg.Topic()] = rawMqttMessageStr
	mqttMessageReceiveTimestampMap[msg.Topic()] = time.Now()
	logger.Infof("MESSAGE_STR updated")
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	logger.Infof("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	logger.Infof("Connect lost: %v", err)
}

func sub(client mqtt.Client, topic string) {
	// topic := "topic/test"
	token := client.Subscribe(topic, 1, receiver)
	token.Wait()
	logger.Infof("Subscribed to topic: %s", topic)
}

func receiver(client mqtt.Client, msg mqtt.Message) {
	msg.Ack()
	messagePubHandler(client, msg)
	message := string(msg.Payload())
	MutexProcess(msg.Topic(), message)
	logger.Infof("Received message:{id:%v, message:%v}", strconv.Itoa(int(msg.MessageID())), message)
}

func MutexProcess(topic string, message string) {
	if mutexBlocking && strings.Contains(message, controlMsgs[currentControlMsg]) {
		logger.Infof("Resetting mutex")
		mutexBlocking = false
		currentControlMsg = ""
	}
}

func ConfigFiniteStateMachine(minsts map[string]string) {
	controlMsgs = minsts
}

// DeviceCommandHandlerMQTT handler for Mqtt
type DeviceCommandHandlerMQTT struct {
	// client                         *rest.RESTClient
	HandlerMetaData *HandlerMetaData
}

func (handler DeviceCommandHandlerMQTT) commandHandleFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// handlerEdgeDeviceSpec := handler.HandlerMetaData.edgeDeviceSpec
		reqType := r.Method
		topic := handler.HandlerMetaData.properties.MQTTTopic
		switch reqType {
		case http.MethodGet:
			returnMessage := ReturnBody{
				MQTTMessage:   mqttMessageInstructionMap[topic],
				MQTTTimestamp: mqttMessageReceiveTimestampMap[topic].String(),
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
		case http.MethodPost, http.MethodPut:
			mqttTopic := handler.HandlerMetaData.properties.MQTTTopic
			logger.Infof("the controlMsgs is %v", controlMsgs)
			if mutexBlocking {
				blockedMessage := fmt.Sprintf("Device is blocked by %v controlMsg now! %v", currentControlMsg, time.Now())
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

			// TODO handle error asynchronously
			token := client.Publish(mqttTopic, 1, false, body)
			if token.Error() != nil {
				logger.Errorf("Error when publish Data to MQTTServer,%v", token.Error())
				http.Error(w, "Error to publish a message to server", http.StatusBadRequest)
				return
			}
			if _, isMutexState := controlMsgs[string(requestBody)]; isMutexState {
				mutexBlocking = true
				currentControlMsg = string(requestBody)
				logger.Infof("Message %s is mutex, blocking.", requestBody)
			}
			logger.Infof("Info: Success To publish a message %v to MQTTServer!", requestBody)
			return
		default:
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

	return "", fmt.Errorf("instruction %v not found in list of deviceshifu instructions", instructionName)
}

// TODO: update configs
// TODO: update status based on telemetry

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
					return false, fmt.Errorf("device %v telemetry %v does not have an instruction name", ds.base.Name, telemetry)
				}

				instruction := *telemetryProperties.DeviceShifuTelemetryProperties.DeviceInstructionName
				mqttTopic, err := ds.getMQTTTopicFromInstructionName(instruction)
				if err != nil {
					logger.Errorf("%v", err.Error())
					return false, err
				}

				// use mqtttopic to get the mqttMessageReceiveTimestampMap
				// determine whether the message interval exceed DeviceShifuTelemetryUpdateIntervalInMilliseconds
				// return true if there is a topic message interval is normal
				// return false if the time interval of all topics is abnormal
				nowTime := time.Now()
				if int64(nowTime.Sub(mqttMessageReceiveTimestampMap[mqttTopic]).Milliseconds()) < *telemetrySettings.DeviceShifuTelemetryUpdateIntervalInMilliseconds {
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
