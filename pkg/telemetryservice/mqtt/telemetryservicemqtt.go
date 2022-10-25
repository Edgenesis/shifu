package mqtt

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"k8s.io/klog"
)

func BindMQTTServicehandler(request v1alpha1.TelemetryRequest) error {
	client, err := connectToMQTT(request.MQTTSetting)
	if err != nil {
		klog.Errorf("Error to connect to mqtt server, error: %#v", err)
		return err
	}
	defer (*client).Disconnect(0)

	token := (*client).Publish(*request.MQTTSetting.MQTTTopic, 1, false, request.RawData)
	if token.Error() != nil {
		klog.Errorf("Error when publish Data to MQTTServer, error: %#v", err.Error())
		return err
	}
	klog.Infof("Info: Success To publish a message %v to %v", string(request.RawData), request.MQTTSetting.MQTTServerAddress)
	return nil
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
	klog.Infof("Connect to %v success!", settings.MQTTServerAddress)
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
