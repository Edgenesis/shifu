package deviceshifuMQTT

type DeviceShifuMQTTReturnBody struct {
	MQTTMessage   string `json:"mqtt_message"`
	MQTTTimestamp string `json:"mqtt_receive_timestamp"`
}
