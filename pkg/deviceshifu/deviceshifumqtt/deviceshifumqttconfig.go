package deviceshifumqtt

// ReturnBody Body of mqtt's reply
type ReturnBody struct {
	MQTTMessage   string `json:"mqtt_message"`
	MQTTTimestamp string `json:"mqtt_receive_timestamp"`
}

// RequestBody Body of mqtt's request by POST method
type RequestBody struct {
	MQTTTopic     string `json:"mqtt_topic"`
	MQTTMessage   []byte `json:"mqtt_message"`
}
