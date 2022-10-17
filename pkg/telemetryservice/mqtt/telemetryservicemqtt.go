package mqtt

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/edgenesis/shifu/pkg/deviceshifu/utils"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/edgenesis/shifu/pkg/telemetryservice/config"
	"k8s.io/klog"
)

const DefaultServerPort = ":8080"

var (
	mqttMessageStr              string
	mqttMessageReceiveTimestamp time.Time
)

func New() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	stop := make(chan struct{}, 1)
	Start(stop, mux)
}

func Start(stop <-chan struct{}, mux *http.ServeMux) {
	addr := DefaultServerPort

	server := http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go server.ListenAndServe()
	klog.Infof("Listening at %#v", addr)
	<-stop
	server.Close()
}

func handler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		klog.Errorf("Error when Read Data From Body, error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	telemetryRequest := &config.TelemetryRequest{}

	err = json.Unmarshal(body, telemetryRequest)
	if err != nil || telemetryRequest.MQTTSetting == nil {
		klog.Errorf("Error when unmarshal body to telemetryBody")
		http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
		return
	}

	client, err := connectToMQTT(telemetryRequest.MQTTSetting)
	if err != nil {
		klog.Errorf("Error to connect to mqtt server, error: %#v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer (*client).Disconnect(0)

	(*client).Publish(*telemetryRequest.MQTTSetting.MQTTTopic, 1, false, telemetryRequest.RawData)
}

func connectToMQTT(settings *v1alpha1.MQTTSetting) (*mqtt.Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", *settings.MQTTServerAddress))
	opts.SetClientID("shifu")
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
	klog.Infof("Received message: %v from topic: %v", msg.Payload(), msg.Topic())
	rawMqttMessageStr := string(msg.Payload())
	_, shouldUsePythonCustomProcessing := deviceshifubase.CustomInstructionsPython[msg.Topic()]
	klog.Infof("Topic %v is custom: %v", msg.Topic(), shouldUsePythonCustomProcessing)
	if shouldUsePythonCustomProcessing {
		klog.Infof("Topic %v has a python customized handler configured.\n", msg.Topic())
		mqttMessageStr = utils.ProcessInstruction(deviceshifubase.PythonHandlersModuleName, msg.Topic(), rawMqttMessageStr, deviceshifubase.PythonScriptDir)
	} else {
		mqttMessageStr = rawMqttMessageStr
	}
	mqttMessageReceiveTimestamp = time.Now()
	klog.Infof("MESSAGE_STR updated")
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	klog.Infof("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	klog.Infof("Connect lost: %v", err)
}
