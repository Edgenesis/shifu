package mqtt

import "github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"

type TelemetryRequest struct {
	RawData     []byte                `json:"rawData,omitempty"`
	MQTTSetting *v1alpha1.MQTTSetting `json:"mqttSetting,omitempty"`
}
