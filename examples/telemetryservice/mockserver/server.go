package main

import (
	"net"
	"net/http"
	"os"
	"time"

	proto "github.com/huin/mqtt"
	"github.com/jeffallen/mqtt"
	"k8s.io/klog"
)

var (
	serverAddress = os.Getenv("MQTT_SERVER_ADDRESS")
	clientAddress = os.Getenv("MQTT_CLIENT_ADDRESS")
)

type myClient struct {
	*mqtt.ClientConn
}

func main() {
	stop := make(chan struct{}, 1)
	status := make(chan struct{}, 1)
	go StartMQTTServer(stop, status)
	<-status
	client := NewClient()

	mux := http.NewServeMux()
	mux.HandleFunc("/", client.GetLatestData)

	klog.Infof("Client listening at %v", clientAddress)
	err := http.ListenAndServe(clientAddress, mux)
	if err != nil {
		klog.Errorf("Error when server running, errors: %v", err)
	}
	stop <- struct{}{}
}

func (client myClient) GetLatestData(w http.ResponseWriter, r *http.Request) {
	select {
	case pubData := <-client.Incoming:
		pubData.Payload.WritePayload(w)
	case <-time.After(time.Second * 5):
		http.Error(w, "no data", http.StatusInternalServerError)
	}
}

func NewClient() *myClient {
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		klog.Fatalf("Enable to connect to MQTT Server, error %v", err)
	}

	client := mqtt.NewClientConn(conn)
	err = client.Connect("username", "password")
	if err != nil {
		klog.Fatalf("error when create client, error: %v", err)
	}

	tq := []proto.TopicQos{
		{
			Topic: "/test",
			Qos:   proto.QosAtLeastOnce,
		},
	}
	client.Subscribe(tq)
	return &myClient{client}
}

func StartMQTTServer(stop <-chan struct{}, status chan struct{}) {
	lis, err := net.Listen("tcp", serverAddress)
	if err != nil {
		klog.Fatalf("Error when Listen ad %v, error: %v", serverAddress, err)
	}
	klog.Infof("mockDevice listen at %v", serverAddress)
	svr := mqtt.NewServer(lis)
	svr.Start()

	status <- struct{}{}
	select {
	case <-svr.Done:
	case <-stop:
	}
	lis.Close()
}
