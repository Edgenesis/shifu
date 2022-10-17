# Building ***Shifu*** components:
## Requirement:
Finish setting up using the setup guide: [Windows](develop-on-windows.md)/[Mac OS](develop-on-mac.md)

## Overview:
We have provided a `Docker` Dev Container environment that simplifies setup and provides consistent environment across all platforms.

## Building
### 1. Build ***deviceShifu*** binaries directly:
Navigate to `Shifu`'s root directory, use the following commands to build different ***deviceShifu***:

For `HTTP to HTTP` ***deviceShifu***:
```sh
CGO_ENABLED=0 GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) go build -a -o output/deviceshifu-http-http cmd/deviceshifu/cmdHTTP/main.go
```
For `HTTP to Socket` ***deviceShifu***:
```sh
CGO_ENABLED=0 GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) go build -a -o output/deviceshifu-http-socket cmd/deviceshifu/cmdSocket/main.go
```
For `HTTP to MQTT` ***deviceShifu***:
```sh
CGO_ENABLED=0 GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) go build -a -o output/deviceshifu-http-mqtt cmd/deviceshifu/cmdMQTT/main.go
```
For `HTTP to OPC UA` ***deviceShifu***:
```sh
CGO_ENABLED=0 GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) go build -a -o output/deviceshifu-http-opcua cmd/deviceshifu/cmdOPCUA/main.go
```

### 2. Build ***deviceShifu*** `Docker` images:
Run the following command, this will build the following ***deviceShifu*** images with current tag and load it into `Docker` images:
```sh
make buildx-load-image-deviceshifu
```

Images built:
```
edgenesis/deviceshifu-http-http:{VERSION}
edgenesis/deviceshifu-http-socket:{VERSION}
edgenesis/deviceshifu-http-mqtt:{VERSION}
edgenesis/deviceshifu-http-opcua:{VERSION}
```

### 3. Build ***shifuController*** binary directly:
Navigate to `shifu/k8s/crd`, use the following command to build ***shifuController*** binary:
```sh
make build
```

### 4. Build ***shifuController*** `Docker` image:
Run the following command, this will build the following ***shifuController*** image with current tag and load it into `Docker` images:
```sh
make docker-buildx-load IMG=edgehub/edgedevice-controller-multi:v0.0.1
```

# What's next?
Follow our [user guide](use-shifu.md) and start using `Shifu`.