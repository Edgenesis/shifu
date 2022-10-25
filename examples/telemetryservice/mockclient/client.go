package main

/*
mockHTTPClient, using this file will send a message to telemetryService
*/
import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
)

type TelemetryRequest struct {
	RawData     []byte       `json:"rawData,omitempty"`
	MQTTSetting *MQTTSetting `json:"mqttSetting,omitempty"`
}

type MQTTSetting struct {
	MQTTTopic         *string `json:"MQTTTopic,omitempty"`
	MQTTServerAddress *string `json:"MQTTServerAddress,omitempty"`
}

func main() {
	targetMqttServer := os.Getenv("TARGET_MQTT_SERVER_ADDRESS")
	targetServer := "http://" + os.Getenv("TARGET_SERVER_ADDRESS")

	err := sendRequest(targetServer, targetMqttServer)
	if err != nil {
		panic(err)
	}
}

func sendRequest(targetServer string, mqttServerAddress string) error {
	defaultTopic := "/test"
	req := &TelemetryRequest{
		RawData: []byte("testData"),
		MQTTSetting: &MQTTSetting{
			MQTTTopic:         &defaultTopic,
			MQTTServerAddress: &mqttServerAddress,
		},
	}

	requestBody, err := json.Marshal(req)
	if err != nil {
		return err
	}

	_, err = http.DefaultClient.Post(targetServer, "application/json", bytes.NewBuffer(requestBody))
	return err
}
