# DeviceShifu Developing Guide

## Introduction
`DeviceShifu` represents the digital twin of a physical device. It takes HTTP requests and support various protocols to communicate with devices such as MQTT and OPCUA. 
Each protocol is created as a new type of `DeviceShifu`. Check the [DeviceShifu](https://github.com/Edgenesis/shifu/tree/main/pkg/deviceshifu) directory to see the protocols we already supported.


## Components

In order to develop a new type of `deviceShifu_xxx` (xxx stands for the protocol name you want to support) , you need to make change to or create the following components:

### CRD
`CRD` stands for `Customized Resource Definition`, we created a new `CRD` called `edgeDevice` to serve as the definition of the mapping of the actual device.
To create a new type of `deviceShifu_xxx`, you may need to change the files under directory `pkg/k8s/crd/`

Some protocols may require specific settings, like `MQTT`：
```yaml
  MQTTSetting:
    description: MQTTSetting defines MQTT specific settings when connecting
      to an EdgeDevice
    properties:
      MQTTServerAddress:
        type: string
      MQTTServerSecret:
        type: string
      MQTTTopic:
        type: string
    type: object
```
In order to let your `deviceShifu_xxx` pickup those protocol specific settings ,
you need to add these setting schema to `shifu_install.yml`, `config_crd.yaml` and `config_default.yaml` files in `properties` under `PortocolSettings`


### API
We put the `CRD` API definitions in Golang under API directory. For all the settings you added in `CRD`, 
you need to put it in `pkg/k8s/api/v1apha1/edgedevice_types.go`. Also take `MQTT` as example:

```go
// ProtocolSettings defines protocol settings when connecting to an EdgeDevice
type ProtocolSettings struct {
+	MQTTSetting   *MQTTSetting   `json:"MQTTSetting,omitempty"`
	OPCUASetting  *OPCUASetting  `json:"OPCUASetting,omitempty"`
	SocketSetting *SocketSetting `json:"SocketSetting,omitempty"`
}
```
You first need to add `MQTTSetting` to `ProtocolSettings`. You also need to add the newly created settings in `CRD` here as a struct:
```go
// MQTTSetting defines MQTT specific settings when connecting to an EdgeDevice
type MQTTSetting struct {
	MQTTTopic         *string `json:"MQTTTopic,omitempty"`
	MQTTServerAddress *string `json:"MQTTServerAddress,omitempty"`
	MQTTServerSecret  *string `json:"MQTTServerSecret,omitempty"`
}
```
The struct should be exactly aligned with the setting schema you added in CRD.


### DeviceShifu
`DeviceShifu` is the actual digital twin running as the pod in the k8s cluster. It basically takes HTTP requests and transfer it to whatever underlying protocol needs.

#### deviceshifuxxx
You need to create a new `deviceshifuxxx` directory under `pkg/deviceshifu`. This directory normally contains 2 main files, a `deviceshifuxxx.go` and a `deviceshifuxxxconfig.go`
`deviceshifuxxx.go` mainly contains the actual logic of the program, and `deviceshifuconfig.go` mainly contains the config underlying protocol needs. Take `MQTT` as example again:

[deviceshifumqttconfig.go](https://github.com/Edgenesis/shifu/blob/main/pkg/deviceshifu/deviceshifumqtt/deviceshifumqttconfig.go)
```go
package deviceshifumqtt

// ReturnBody Body of mqtt's reply
type ReturnBody struct {
	MQTTMessage   string `json:"mqtt_message"`
	MQTTTimestamp string `json:"mqtt_receive_timestamp"`
}
```

For `deviceshifuxxx.go`, you can follow the following pattern:
Create a struct name `DeviceShifu` take `*deviceshifubase.DeviceShifuBase` as its field and extends `DeviceShifu` interface in `deviceshifubase.go`. `DeviceShifuBase` contains skeleton code for spawn and start a `DeviceShifu`
Take `MQTT` for e.g: 
```go
// DeviceShifu implemented from deviceshifuBase
type DeviceShifu struct {
	base *deviceshifubase.DeviceShifuBase
}
```
You can also add protocol specific structs, take `OPCUA` as e.g:
```go
// DeviceShifu implemented from deviceshifuBase and OPC UA Setting and client
type DeviceShifu struct {
	base              *deviceshifubase.DeviceShifuBase
	opcuaInstructions *OPCUAInstructions
	opcuaClient       *opcua.Client
}
```
Make a method to create an instance of the struct like:
```go
// New This function creates a new Device Shifu based on the configuration
func New(deviceShifuMetadata *deviceshifubase.DeviceShifuMetaData) (*DeviceShifu, error)
```
`deviceShifuMetadata` is used to create `DeviceShifuBase`. 

Inside the function, you can call 
```go
base, mux, err := deviceshifubase.New(deviceShifuMetadata)
if err != nil {
	return nil, err
}
```
which returns an initialized `DeviceShifuBase` and a server mux. The key part for your `deviceshifuxxx` to work is register the proper handler to take care of all the incoming http request.
For the handler function, you cant take `MQTT`'s [handler function](https://github.com/Edgenesis/shifu/blob/main/pkg/deviceshifu/deviceshifumqtt/deviceshifumqtt.go#L140) as example.

To register the handler you can do something like:
```go
handler := DeviceCommandHandlerMQTT{HandlerMetaData}
mux.HandleFunc("/"+MqttDataEndpoint, handler.commandHandleFunc())
```


`DeviceShifu` can also collect device telemetries specified. If you don't know what telemetry is, please take a look at this doc before proceed: https://github.com/Edgenesis/shifu/blob/main/docs/design/deviceshifu/telemetry.md

`DeviceShifuBase` will take a function:
```go
// collectTelemetry struct of collectTelemetry
type collectTelemetry func() (bool, error)
```
and use it periodically to collect telemetries from device. You can specify how you want the telemetries been collected. 
You can reference `MQTT`'s [telemetry collection](https://github.com/Edgenesis/shifu/blob/main/pkg/deviceshifu/deviceshifumqtt/deviceshifumqtt.go#L206) as example.

After you finished the handler and telemetry logic, use `DeviceShifuBase` to start/stop your program by implementing the `Start` and `Stop` function like (also take `MQTT` as example):
``` go
// Start start Mqtt Telemetry
func (ds *DeviceShifu) Start(stopCh <-chan struct{}) error {
	return ds.base.Start(stopCh, ds.collectMQTTTelemetry)
}

// Stop Stop Http Server
func (ds *DeviceShifu) Stop() error {
	return ds.base.Stop()
}
```

#### main 
To actually run the `deviceshifuxxx` program, you need to have a `main.go`. Create a new folder named `cmdxxx` directory under `cmd/deviceshifu` 
and create a `main.go` under the directory.

Take `MQTT` as an example, the content of the main can be look like:
```go
func main() {
	deviceName := os.Getenv("EDGEDEVICE_NAME")
	namespace := os.Getenv("EDGEDEVICE_NAMESPACE")

	deviceShifuMetadata := &deviceshifubase.DeviceShifuMetaData{
		Name:           deviceName,
		ConfigFilePath: deviceshifubase.DeviceConfigmapFolderPath,
		KubeConfigPath: deviceshifubase.KubernetesConfigDefault,
		Namespace:      namespace,
	}
    // TODO: Change deviceshifumqtt to the deviceshifuxxx you just created
	ds, err := deviceshifumqtt.New(deviceShifuMetadata)
	if err != nil {
		panic(err.Error())
	}

	if err := ds.Start(wait.NeverStop); err != nil {
		panic(err.Error())
	}

	select {}
}
```
#### Dockerfile and makefile

To let your go program run in k8s, you need to package it into a docker image. 
And to do that, you need to create a file named `Dockerfile.deviceshifuXXX` under `dockerfiles` directory.
The dockerfile, take `MQTT` as example, looks like:
```dockerfile
# Build the manager binary
FROM --platform=$BUILDPLATFORM golang:1.18.4 as builder

WORKDIR /shifu

ENV GO111MODULE=on
ENV GOPRIVATE=github.com/Edgenesis

COPY go.mod go.mod
COPY go.sum go.sum
COPY pkg/k8s pkg/k8s
# TODO: Cnahge cmdmqtt to cmdxxx
COPY cmd/deviceshifu/cmdmqtt cmd/deviceshifu/cmdmqtt
COPY pkg/deviceshifu pkg/deviceshifu

RUN go mod download

# Build the Go app
ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -o /output/deviceshifu cmd/deviceshifu/cmdmqtt/main.go

FROM gcr.io/distroless/static-debian11
WORKDIR /
COPY --from=builder /output/deviceshifu deviceshifu

# Command to run the executable
USER 65532:65532
ENTRYPOINT ["/deviceshifu"]
```

***Notice:*** some native protocols can require some native c-binding libraries, you may also want to install int in the dockerfile.

You can utilize `Makefile` to push and build docker images. To make use of `Makefile`, you can add the following lines:
We also take `MQTT` as example here.

To build image:
```makefile
buildx-build-image-deviceshifu-http-mqtt:
	docker buildx build --platform=linux/$(shell go env GOARCH) -f ${PROJECT_ROOT}/dockerfiles/Dockerfile.deviceshifuMQTT \
	 	--build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} \
		-t edgehub/deviceshifu-http-mqtt:${IMAGE_VERSION} --load
```
To push image:
```makefile
buildx-push-image-deviceshifu-http-mqtt:
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm -f ${PROJECT_ROOT}/dockerfiles/Dockerfile.deviceshifuMQTT \
		--build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} \
		-t edgehub/deviceshifu-http-mqtt:${IMAGE_VERSION} --push
```

To build and push all `deviceshifu` types:
```makefile
buildx-build-image-deviceshifu: \
	buildx-build-image-deviceshifu-http-http \
+	buildx-build-image-deviceshifu-http-mqtt \
	buildx-build-image-deviceshifu-http-socket \
	buildx-build-image-deviceshifu-http-opcua
	
buildx-push-image-deviceshifu: \
	buildx-push-image-deviceshifu-http-http \
+	buildx-push-image-deviceshifu-http-mqtt \
	buildx-push-image-deviceshifu-http-socket \
	buildx-push-image-deviceshifu-http-opcua
```

#### Testing
For the newly created `deviceshifuxxx` type, we would recommend you to write unit test against it before deploy it to k8s or create PR to merge it back.
You can reference [deviceshifumqtt_test.go](https://github.com/Edgenesis/shifu/blob/main/pkg/deviceshifu/deviceshifumqtt/deviceshifumqtt_test.go)
and [deviceshifumqttconfig_test.go](https://github.com/Edgenesis/shifu/blob/main/pkg/deviceshifu/deviceshifumqtt/deviceshifumqttconfig_test.go) as a reference.

We also run e2e tests on our pipelines. You should also try to create a mockdevice and add your e2e test to our pipeline before creating a PR. 
For mockdevice you can reference to `pkg/deviceshifu/mockdevice` and for how to write e2e tests please reference to our [pipeline](https://github.com/Edgenesis/shifu/blob/main/azure-pipelines/azure-pipelines.yml#L369-L422).

## Miscellaneous
If you are interested in more detailed guide on how to develop a `deviceshifu`, you can reference to the commit [65e124d9](https://github.com/Edgenesis/shifu/commit/65e124d9823afeca9640a7514c893224f67508a0) on how we created `deviceshifuplc4x`, a deviceshifu that utilizes plc4x library.
