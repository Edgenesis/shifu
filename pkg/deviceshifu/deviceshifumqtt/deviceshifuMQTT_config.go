package deviceshifumqtt

// ReturnBody Body of mqtt's reply
type ReturnBody struct {
	MQTTMessage   string `json:"mqtt_message"`
	MQTTTimestamp string `json:"mqtt_receive_timestamp"`
}
