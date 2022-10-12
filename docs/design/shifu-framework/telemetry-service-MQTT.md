# Telemetry Service MQTT Endpoint Design

## Introduction
Telemetry Service is a standalone service that takes telemetry data collected by `deviceShifu` and push it to designated endpoints for future process.
This doc aims to provide a design on how to send telemetry to MQTT endpoints.

## Design-Goal
Let telemetry service support pushing telemetries to MQTT endpoints.

## Design Non-Goal
1. Let telemetry service support any random endpoints
2. let telemetry service serve as MQTT broker.

## Design Details

Telemetry will be served as an HTTP server. DeviceShifu will push the telemetries collected from physical devices to telemetry service,
and telemetry service would then push the telemetries to the endpoints specified by user.

Request Struct:
```go
type TelemetryRequest struct {
	rawData     string               `json:"raw_data,omitempty"`
	mqttSetting v1alpha1.MQTTSetting `json:"mqtt_setting,omitempty"`
	httpAddress string               `json:"http_address,omitempty"`
}
```

Everytime telemetry service receives a new request, it will fetch out endpoint settings and send the raw data to the corresponding endpoint.

```mermaid
graph LR;
DeviceShifu -->|TelemetryRequest| TelemetryService;
TelemetryService -->|RawData| MQTTEndpoint;

```

TelemtryService would have 2 methods, one extract rawData and endpoint settings, the other push the rawData to the endpoint according to the settings extracted from the first one.

```go
func HandleTelemtryRequest(request *TelemetryRequest) err {
	// extract rawdata and endpoint settings
	...
	pushToMQTTEndPoint(rawData, &mqttSettings)
}

func pushToMQTTEndPoint(rawData byte[], mqttSettings *v1alpha1.MQTTSetting) err {
	// push rawData to mqtt broker
}
```