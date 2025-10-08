# DeviceShifu Development Guide

## Prerequisite

If you haven't done so, please take a look at the following documents first:

1. [contribution guide](../contribution/contributing.md)
2. [deviceshifu design](../design/design-deviceShifu.md)

## Introduction

`DeviceShifu` is the digital twin of physical device. It receives HTTP requests and support various protocols to communicate with devices such as MQTT and OPCUA.
A new type of `DeviceShifu` is assigned to each protocol.
Check [DeviceShifu](https://github.com/Edgenesis/shifu/tree/main/pkg/deviceshifu) directory to see the protocols we already support.

## Overview

Below is the high-level of how to build a `DeviceShifu` called `deviceshifuxxx`:

1. [Modify CRD](#crd).
2. [Modify API](#api).
3. [Add DeviceShifu source files](#deviceshifu)

## Components

In order to create a new type of `deviceShifuxxx` (xxx stands for the protocol name you wish to support), you need to create or modify the following components:

### CRD

`CRD` stands for `Customized Resource Definition`, we created a new `CRD` called `edgeDevice` to serve as the definition of the mapping of the physical device.
To create a new type of `deviceShifuxxx`, you may need to change the files under directory `pkg/k8s/crd/`

Specific settings may be required for some protocols, like `MQTT`：

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

In order to allow your `deviceShifuXXX` to receive these specific settings,
you need to add the setting schema to `shifu_install.yml`, `config_crd.yaml` and `config_default.yaml` files in `properties` under `PortocolSettings`

### API

We put the `CRD` API definitions in Golang under API directory. For all the settings you added in `CRD`, 
you need to put it in `pkg/k8s/api/v1apha1/edgedevice_types.go`. Also take `MQTT` as example:

```go
// ProtocolSettings defines protocol settings when connecting to an EdgeDevice
type ProtocolSettings struct {
+   MQTTSetting   *MQTTSetting   `json:"MQTTSetting,omitempty"`
    OPCUASetting  *OPCUASetting  `json:"OPCUASetting,omitempty"`
    SocketSetting *SocketSetting `json:"SocketSetting,omitempty"`
}
```

First you need to add `MQTTSetting` to `ProtocolSettings`. Then you need to add the newly written settings in `CRD` as a struct:

```go
// MQTTSetting defines MQTT specific settings when connecting to an EdgeDevice
type MQTTSetting struct {
    MQTTTopic           *string `json:"MQTTTopic,omitempty"`
    MQTTServerAddress   *string `json:"MQTTServerAddress,omitempty"`
    MQTTServerSecret    *string `json:"MQTTServerSecret,omitempty"``
    MQTTServerPassword   string `json:"MQTTServerPassword,omitempty"`
    MQTTServerUserName   string  `json:"MQTTServerUserName,omitempty"`
}
```

The struct should be perfectly aligned with the setting schema you added in CRD.

### DeviceShifu

`DeviceShifu` is the digital twin running as a pod in the k8s cluster. Basically it converts HTTP requests to whatever underlying protocol needs.

#### shifuctl

`shifuctl` is a command-line tool can help bootstrapping a new `DeviceShifu`. Here's how to use it:

1. Set environment variable `SHIFU_ROOT_DIR` to root directory of shifu.
On Linux: `export SHIFU_ROOT_DIR=[root directory of shifu]`.
2. Bootstrap barebone source files with `shifuctl add deviceshifu--name deviceshifuxxx`.
3. Under `cmd/deviceshifu/`, you will see `deviceshifuxxx/main.go`.You don't need to modify this file.
4. Under `pkg/deviceshifu/`, you will see 4 files in `deviceshifuxxx`directory. You need to modify these files accordingly.

#### deviceshifuxxx

In `pkg/deviceshifu/deviceshifuxxx`, there are 4 files, including 2 main ones: `deviceshifuxxx.go` and `deviceshifuxxxconfig.go`
`deviceshifuxxx.go` mainly contains actual logic of the program, and `deviceshifuconfig.go` mainly contains the configuration underlying protocol needs. Take `MQTT` as example:

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
Create a struct named `DeviceShifu`, and take `*deviceshifubase.DeviceShifuBase` as its field. Implement `DeviceShifu` interface in `deviceshifubase.go`. `DeviceShifuBase` contains skeleton code for creating and starting a `DeviceShifu`.
Take `MQTT` for example:

```go
// DeviceShifu implemented from deviceshifuBase
type DeviceShifu struct {
    base *deviceshifubase.DeviceShifuBase
}
```

You can also add protocol specific structs, take `OPCUA` as example:

```go
// DeviceShifu implemented from deviceshifuBase and OPC UA Setting and client
type DeviceShifu struct {
    base              *deviceshifubase.DeviceShifuBase
    opcuaInstructions *OPCUAInstructions
    opcuaClient       *opcua.Client
}
```

Write a method to create an instance of the struct:

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

which returns an initialized `DeviceShifuBase` and a server mux. The key part for your `deviceshifuxxx` to work is to register the proper handler to take care of all the incoming http request.
For the handler function, you can take `MQTT`'s [handler function](https://github.com/Edgenesis/shifu/blob/main/pkg/deviceshifu/deviceshifumqtt/deviceshifumqtt.go#L140) as example.

To register the handler you can:

```go
handler := DeviceCommandHandlerMQTT{HandlerMetaData}
mux.HandleFunc("/"+MqttDataEndpoint, handler.commandHandleFunc())
```

`DeviceShifu` can also collect device specific telemetry. If you don't know what telemetry is, please check this doc before you proceed: https://github.com/Edgenesis/shifu/blob/main/docs/design/deviceshifu/telemetry.md

`DeviceShifuBase` will take a function:

```go
// collectTelemetry struct of collectTelemetry
type collectTelemetry func() (bool, error)
```

and use it periodically to collect telemetry data from device. You can define how you want the telemetry data to be collected. 
You can reference `MQTT`'s [telemetry collection](https://github.com/Edgenesis/shifu/blob/main/pkg/deviceshifu/deviceshifumqtt/deviceshifumqtt.go#L206) as example.

After you finished the handler and telemetry logic, use `DeviceShifuBase` to start/stop your program by implementing the `Start` and `Stop` function like (also take `MQTT` as example):

``` go
// Start start Mqtt Telemetry
func (ds *DeviceShifu) Start(stopCh <-chan struct{}) error {
    return ds.base.Start(stopCh, ds.collectMQTTTelemetry)
}

// Stop Http Server
func (ds *DeviceShifu) Stop() error {
    return ds.base.Stop()
}
```

#### main

In order to run the `deviceshifuxxx` program, you need a `main.go`. Create a new folder named `cmdxxx` directory under `cmd/deviceshifu` 
and create a `main.go` under the directory.

Take `MQTT` as an example, the content of the main can be:

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

To run go program in k8s, you need to package it into a docker image. 
To do that, you need to create a file named `Dockerfile.deviceshifuxxx` under `dockerfiles` directory.
The dockerfile, take `MQTT` as example, can be:

```dockerfile
# Build the manager binary
FROM --platform=$BUILDPLATFORM golang:1.22.0 as builder

WORKDIR /shifu

ENV GO111MODULE=on
ENV GOPRIVATE=github.com/Edgenesis

COPY go.mod go.mod
COPY go.sum go.sum
COPY pkg/k8s pkg/k8s
# TODO: Change cmdmqtt to cmdxxx
COPY cmd/deviceshifu/cmdmqtt cmd/deviceshifu/cmdmqtt
COPY pkg/deviceshifu pkg/deviceshifu

RUN go mod download

# Build the Go App
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

***Notice:*** some native protocols need some native c-binding libraries, you may need to install it in the dockerfile.

You can utilize `Makefile` to push and build docker images. To make use of `Makefile`, you can add the following lines:
Take `MQTT` as example here.

Build image:

```makefile
buildx-build-image-deviceshifu-http-mqtt:
    docker buildx build --platform=linux/$(shell go env GOARCH) -f ${PROJECT_ROOT}/dockerfiles/Dockerfile.deviceshifuMQTT \
        --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} \
        -t edgehub/deviceshifu-http-mqtt:${IMAGE_VERSION} --load
```

Push image:

```makefile
buildx-push-image-deviceshifu-http-mqtt:
    docker buildx build --platform=linux/amd64,linux/arm64,linux/arm -f ${PROJECT_ROOT}/dockerfiles/Dockerfile.deviceshifuMQTT \
        --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} \
        -t edgehub/deviceshifu-http-mqtt:${IMAGE_VERSION} --push
```

Build and push all `deviceshifu` types:

```makefile
buildx-build-image-deviceshifu: \
    buildx-build-image-deviceshifu-http-http \
+   buildx-build-image-deviceshifu-http-mqtt \
    buildx-build-image-deviceshifu-http-socket \
    buildx-build-image-deviceshifu-http-opcua

buildx-push-image-deviceshifu: \
    buildx-push-image-deviceshifu-http-http \
+   buildx-push-image-deviceshifu-http-mqtt \
    buildx-push-image-deviceshifu-http-socket \
    buildx-push-image-deviceshifu-http-opcua
```

#### Test

For the newly created `deviceshifuxxx` type, we recommend you to run unit test on it before deploying it to k8s or creating PR to merge it.
You can check [deviceshifumqtt_test.go](https://github.com/Edgenesis/shifu/blob/main/pkg/deviceshifu/deviceshifumqtt/deviceshifumqtt_test.go)
and [deviceshifumqttconfig_test.go](https://github.com/Edgenesis/shifu/blob/main/pkg/deviceshifu/deviceshifumqtt/deviceshifumqttconfig_test.go).

We also run e2e tests on our pipelines. You can try to create a mockdevice and add your e2e test to our pipeline before creating a PR. 
For mockdevice you can check `pkg/deviceshifu/mockdevice` and for how to write e2e tests please reference to our [pipeline](https://github.com/Edgenesis/shifu/blob/main/azure-pipelines/azure-pipelines.yml#L369-L422).

## Miscellaneous

If you are interested in more detailed guides on how to develop a `deviceshifu`, you can reference the historical commit [65e124d9](https://github.com/Edgenesis/shifu/commit/65e124d9823afeca9640a7514c893224f67508a0) that introduced the now-retired `deviceshifuplc4x` implementation.
