package mqtt

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"k8s.io/klog"
)

func BindMQTTServicehandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		klog.Errorf("Error when Read Data From Body, error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	telemetryRequest := &TelemetryRequest{}

	err = json.Unmarshal(body, telemetryRequest)
	if err != nil || telemetryRequest.MQTTSetting == nil {
		klog.Errorf("Error when unmarshal body to telemetryBody")
		http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
		return
	}

	klog.Infof("Info: pub Info %v To %v", string(telemetryRequest.RawData), *telemetryRequest.MQTTSetting.MQTTServerAddress)

	client, err := connectToMQTT(telemetryRequest.MQTTSetting)
	if err != nil {
		klog.Errorf("Error to connect to mqtt server, error: %#v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer (*client).Disconnect(0)

	token := (*client).Publish(*telemetryRequest.MQTTSetting.MQTTTopic, 1, false, telemetryRequest.RawData)
	if token.Error() != nil {
		klog.Errorf("Error when publish Data to MqttServer, error: %#v", err.Error())
		return
	}
	klog.Infof("Info: Success To publish a message %v to %v", string(telemetryRequest.RawData), telemetryRequest.MQTTSetting.MQTTServerAddress)
}

func connectToMQTT(settings *v1alpha1.MQTTSetting) (*mqtt.Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", *settings.MQTTServerAddress))
	opts.SetClientID("shifu-service")
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		klog.Errorf("Error when connect to server error: %v", token.Error())
		return nil, token.Error()
	}

	return &client, nil
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	klog.Infof("MESSAGE_STR updated")
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	klog.Infof("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	klog.Infof("Connect lost: %v", err)
}
