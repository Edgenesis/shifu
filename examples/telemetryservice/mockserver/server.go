package main

import (
	"fmt"
	"net/http"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/klog"
)

var (
	serverAddress = os.Getenv("MQTT_SERVER_ADDRESS")
	clientAddress = os.Getenv("HTTP_SRVER_ADDRESS")
	tmpMessage    string
)

func main() {
	client, err := connectToMQTT(serverAddress)
	if err != nil {
		klog.Errorf("Error when connect to mqtt server")
	}
	subTopic(client)

	mux := http.NewServeMux()

	mux.HandleFunc("/", GetLatestData)

	klog.Infof("Client listening at %v", clientAddress)
	err = http.ListenAndServe(clientAddress, mux)
	if err != nil {
		klog.Errorf("Error when server running, errors: %v", err)
	}
}

func GetLatestData(w http.ResponseWriter, r *http.Request) {
	if tmpMessage == "" {
		http.Error(w, "empty", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s", string(tmpMessage))
}

func subTopic(client *mqtt.Client) {
	token := (*client).Subscribe("/test", 1, nil)
	token.Wait()
	if token.Error() != nil {
		klog.Fatalf("Error when sub to topic,error: %v", token.Error().Error())
	}
	klog.Infof("Subscribed to topic: /test")
}

var messageSubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	klog.Infof("%v", string(msg.Payload()))
	tmpMessage = string(msg.Payload())
}

func connectToMQTT(address string) (*mqtt.Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", address))
	opts.SetClientID("mockServer")
	opts.SetDefaultPublishHandler(messageSubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		klog.Errorf("Error when connect to server error: %v", token.Error())
		return nil, token.Error()
	}

	return &client, nil
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	klog.Infof("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	klog.Infof("Connect lost: %v", err)
}
