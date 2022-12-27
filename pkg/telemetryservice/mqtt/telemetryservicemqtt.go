package mqtt

import (
	"encoding/json"
	"fmt"
	"github.com/edgenesis/shifu/pkg/telemetryservice/utils"
	"go.uber.org/zap"
	"io"
	"net/http"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"k8s.io/klog"
)

var zlog *zap.SugaredLogger

func init() {
	logger, _ := zap.NewProduction()
	zlog = logger.Sugar()
}

func BindMQTTServicehandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		klog.Errorf("Error when Read Data From Body, error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	klog.Infof("requestBody: %s", string(body))
	request := v1alpha1.TelemetryRequest{}

	err = json.Unmarshal(body, &request)
	if err != nil {
		klog.Errorf("Error to Unmarshal request body to struct")
		http.Error(w, "unexpected end of JSON input", http.StatusBadRequest)
		return
	}

	injectSecret(request.MQTTSetting)

	client, err := connectToMQTT(request.MQTTSetting)
	if err != nil {
		klog.Errorf("Error to connect to mqtt server, error: %#v", err)
		http.Error(w, "Error to connect to server", http.StatusBadRequest)
		return
	}
	defer (*client).Disconnect(0)

	token := (*client).Publish(*request.MQTTSetting.MQTTTopic, 1, false, request.RawData)
	if token.Error() != nil {
		klog.Errorf("Error when publish Data to MQTTServer, error: %#v", err.Error())
		http.Error(w, "Error to publish a message to server", http.StatusBadRequest)
		return
	}
	klog.Infof("Info: Success To publish a message %v to %v", string(request.RawData), request.MQTTSetting.MQTTServerAddress)
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
	klog.Infof("Connect to %v success!", *settings.MQTTServerAddress)
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

func injectSecret(setting *v1alpha1.MQTTSetting) {
	if setting == nil {
		zlog.Warnf("empty telemetry service setting.")
		return
	}
	pwd, err := utils.GetPasswordFromSecret(*setting.MQTTServerSecret)
	if err != nil {
		zlog.Errorf("unable to get secret for telemetry %v, error: %v", *setting.MQTTServerSecret, err)
		return
	}
	*setting.MQTTServerSecret = pwd
	zlog.Infof("MQTTSetting.Secret load from secret")
}
