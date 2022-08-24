package deviceshifuMQTT

import (
	"encoding/json"
	"fmt"
	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"log"
	"net/http"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type DeviceShifu struct {
	base *deviceshifubase.DeviceShifuBase
}

type DeviceShifuMQTTHandlerMetaData struct {
	edgeDeviceSpec v1alpha1.EdgeDeviceSpec
	// instruction    string
	// properties     *DeviceShifuInstruction
}

const (
	MQTT_DATA_ENDPOINT         string = "mqtt_data"
	DEFAULT_UPDATE_INTERVAL_MS int64  = 3000
)

var (
	MQTT_MESSAGE_STR               string
	MQTT_MESSAGE_RECEIVE_TIMESTAMP time.Time
)

func New(deviceShifuMetadata *deviceshifubase.DeviceShifuMetaData) (*DeviceShifu, error) {
	base, mux, err := deviceshifubase.New(deviceShifuMetadata)
	if err != nil {
		return nil, err
	}

	if deviceShifuMetadata.KubeConfigPath != deviceshifubase.DEVICE_KUBECONFIG_DO_NOT_LOAD_STR {
		switch protocol := *base.EdgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolMQTT:
			mqttSetting := *base.EdgeDevice.Spec.ProtocolSettings.MQTTSetting
			var mqttServerAddress string
			if mqttSetting.MQTTTopic == nil || *mqttSetting.MQTTTopic == "" {
				return nil, fmt.Errorf("MQTT Topic cannot be empty")
			}

			if mqttSetting.MQTTServerAddress == nil || *mqttSetting.MQTTServerAddress == "" {
				// return nil, fmt.Errorf("MQTT server cannot be empty")
				log.Println("MQTT Server Address is empty, use address instead")
				mqttServerAddress = *base.EdgeDevice.Spec.Address
			} else {
				mqttServerAddress = *mqttSetting.MQTTServerAddress
			}

			opts := mqtt.NewClientOptions()
			opts.AddBroker(fmt.Sprintf("tcp://%s", mqttServerAddress))
			opts.SetClientID(*&base.EdgeDevice.Name)
			opts.SetDefaultPublishHandler(messagePubHandler)
			opts.OnConnect = connectHandler
			opts.OnConnectionLost = connectLostHandler
			client := mqtt.NewClient(opts)
			if token := client.Connect(); token.Wait() && token.Error() != nil {
				panic(token.Error())
			}

			sub(client, *mqttSetting.MQTTTopic)

			deviceShifuMQTTHandlerMetaData := &DeviceShifuMQTTHandlerMetaData{
				base.EdgeDevice.Spec,
			}

			handler := DeviceCommandHandlerMQTT{deviceShifuMQTTHandlerMetaData}
			mux.HandleFunc("/"+MQTT_DATA_ENDPOINT, handler.commandHandleFunc())
		}
	}
	ds := &DeviceShifu{base: base}

	ds.base.UpdateEdgeDeviceResourcePhase(v1alpha1.EdgeDevicePending)
	return ds, nil
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message: %v from topic: %v\n", msg.Payload(), msg.Topic())
	MQTT_MESSAGE_STR = string(msg.Payload())
	MQTT_MESSAGE_RECEIVE_TIMESTAMP = time.Now()
	log.Print("MESSAGE_STR updated")
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connect lost: %v", err)
}

func sub(client mqtt.Client, topic string) {
	// topic := "topic/test"
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	log.Printf("Subscribed to topic: %s", topic)
}

func deviceHealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, deviceshifubase.DEVICE_IS_HEALTHY_STR)
}

type DeviceCommandHandlerMQTT struct {
	// client                         *rest.RESTClient
	deviceShifuMQTTHandlerMetaData *DeviceShifuMQTTHandlerMetaData
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

func (handler DeviceCommandHandlerMQTT) commandHandleFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// handlerEdgeDeviceSpec := handler.deviceShifuMQTTHandlerMetaData.edgeDeviceSpec
		reqType := r.Method

		if reqType == http.MethodGet {
			returnMessage := DeviceShifuMQTTReturnBody{
				MQTTMessage:   MQTT_MESSAGE_STR,
				MQTTTimestamp: MQTT_MESSAGE_RECEIVE_TIMESTAMP.String(),
			}

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(returnMessage)
		} else {
			http.Error(w, "must be GET method", http.StatusBadRequest)
			log.Println("Request type " + reqType + " is not supported yet!")
			return
		}

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

func (ds *DeviceShifu) startHttpServer(stopCh <-chan struct{}) error {
	fmt.Printf("deviceshifu %s's http server started\n", ds.base.Name)
	return ds.base.Server.ListenAndServe()
}

// TODO: update configs
// TODO: update status based on telemetry

func (ds *DeviceShifu) collectMQTTTelemetry() (bool, error) {

	if ds.base.EdgeDevice.Spec.Protocol != nil {
		switch protocol := *ds.base.EdgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolMQTT:
			telemetrySettings := ds.base.DeviceShifuConfig.Telemetries.DeviceShifuTelemetrySettings
			if ds.base.EdgeDevice.Spec.Address == nil {
				return false, fmt.Errorf("Device %v does not have an address", ds.base.Name)
			}

			if telemetrySettings.DeviceShifuTelemetryUpdateIntervalInMilliseconds == nil {
				*telemetrySettings.DeviceShifuTelemetryUpdateIntervalInMilliseconds = DEFAULT_UPDATE_INTERVAL_MS
			}

			nowTime := time.Now()
			if int64(nowTime.Sub(MQTT_MESSAGE_RECEIVE_TIMESTAMP).Milliseconds()) < *telemetrySettings.DeviceShifuTelemetryUpdateIntervalInMilliseconds {
				return true, nil
			}
		default:
			log.Printf("EdgeDevice protocol %v not supported in deviceshifu\n", protocol)
			return false, nil
		}
	}

	return false, nil
}

func (ds *DeviceShifu) Start(stopCh <-chan struct{}) error {
	return ds.base.Start(stopCh, ds.collectMQTTTelemetry)
}

func (ds *DeviceShifu) Stop() error {
	return ds.base.Stop()
}
