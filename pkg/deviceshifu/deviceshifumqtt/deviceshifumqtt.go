package deviceshifumqtt

import (
	"encoding/json"
	"fmt"
	"net/http"
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
