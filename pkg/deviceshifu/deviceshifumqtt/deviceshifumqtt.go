package deviceshifumqtt

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"k8s.io/klog/v2"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// DeviceShifu implemented from deviceshifuBase
type DeviceShifu struct {
	base *deviceshifubase.DeviceShifuBase
}

// HandlerMetaData MetaData for EdgeDevice Setting
type HandlerMetaData struct {
	edgeDeviceSpec v1alpha1.EdgeDeviceSpec
	// instruction    string
	// properties     *DeviceShifuInstruction
}

// Str and default value
const (
	MqttDataEndpoint          string = "mqtt_data"
	DefaultUpdateIntervalInMS int64  = 3000
)

var (
	mqttMessageStr              string
	mqttMessageReceiveTimestamp time.Time
)

// New new MQTT Deviceshifu
func New(deviceShifuMetadata *deviceshifubase.DeviceShifuMetaData) (*DeviceShifu, error) {
	base, mux, err := deviceshifubase.New(deviceShifuMetadata)
	if err != nil {
		return nil, err
	}

	if deviceShifuMetadata.KubeConfigPath != deviceshifubase.DeviceKubeconfigDoNotLoadStr {
		switch protocol := *base.EdgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolMQTT:
			mqttSetting := *base.EdgeDevice.Spec.ProtocolSettings.MQTTSetting
			var mqttServerAddress string
			if mqttSetting.MQTTTopic == nil || *mqttSetting.MQTTTopic == "" {
				return nil, fmt.Errorf("MQTT Topic cannot be empty")
			}

			if mqttSetting.MQTTServerAddress == nil || *mqttSetting.MQTTServerAddress == "" {
				// return nil, fmt.Errorf("MQTT server cannot be empty")
				klog.Errorf("MQTT Server Address is empty, use address instead")
				mqttServerAddress = *base.EdgeDevice.Spec.Address
			} else {
				mqttServerAddress = *mqttSetting.MQTTServerAddress
			}

			opts := mqtt.NewClientOptions()
			opts.AddBroker(fmt.Sprintf("tcp://%s", mqttServerAddress))
			opts.SetClientID(base.EdgeDevice.Name)
			opts.SetDefaultPublishHandler(messagePubHandler)
			opts.OnConnect = connectHandler
			opts.OnConnectionLost = connectLostHandler
			client := mqtt.NewClient(opts)
			if token := client.Connect(); token.Wait() && token.Error() != nil {
				panic(token.Error())
			}

			sub(client, *mqttSetting.MQTTTopic)

			HandlerMetaData := &HandlerMetaData{
				base.EdgeDevice.Spec,
			}

			handler := DeviceCommandHandlerMQTT{HandlerMetaData}
			mux.HandleFunc("/"+MqttDataEndpoint, handler.commandHandleFunc())
		}
	}
	deviceshifubase.BindDefaultHandler(mux)

	ds := &DeviceShifu{base: base}

	ds.base.UpdateEdgeDeviceResourcePhase(v1alpha1.EdgeDevicePending)
	return ds, nil
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	klog.Infof("Received message: %v from topic: %v", msg.Payload(), msg.Topic())
	mqttMessageStr = string(msg.Payload())
	mqttMessageReceiveTimestamp = time.Now()
	klog.Infof("MESSAGE_STR updated")
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	klog.Infof("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	klog.Infof("Connect lost: %v", err)
}

func sub(client mqtt.Client, topic string) {
	// topic := "topic/test"
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	klog.Infof("Subscribed to topic: %s", topic)
}

// DeviceCommandHandlerMQTT handler for Mqtt
type DeviceCommandHandlerMQTT struct {
	// client                         *rest.RESTClient
	HandlerMetaData *HandlerMetaData
}

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

func (handler DeviceCommandHandlerMQTT) commandHandleFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// handlerEdgeDeviceSpec := handler.HandlerMetaData.edgeDeviceSpec
		reqType := r.Method

		if reqType == http.MethodGet {
			returnMessage := ReturnBody{
				MQTTMessage:   mqttMessageStr,
				MQTTTimestamp: mqttMessageReceiveTimestamp.String(),
			}

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(returnMessage)
			if err != nil {
				http.Error(w, "Cannot Encode message to json", http.StatusInternalServerError)
				klog.Errorf("Cannot Encode message to json")
				return
			}
		} else {
			http.Error(w, "must be GET method", http.StatusBadRequest)
			klog.Errorf("Request type %v is not supported yet!", reqType)
			return
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

			if interval := telemetrySettings.DeviceShifuTelemetryUpdateIntervalInMilliseconds; interval == nil {
				var telemetryUpdateIntervalInMilliseconds = DefaultUpdateIntervalInMS
				telemetrySettings.DeviceShifuTelemetryUpdateIntervalInMilliseconds = &telemetryUpdateIntervalInMilliseconds
			}

			nowTime := time.Now()
			if int64(nowTime.Sub(mqttMessageReceiveTimestamp).Milliseconds()) < *telemetrySettings.DeviceShifuTelemetryUpdateIntervalInMilliseconds {
				return true, nil
			}
		default:
			klog.Warningf("EdgeDevice protocol %v not supported in deviceshifu", protocol)
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
