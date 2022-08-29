# 构建 ***Shifu*** 组件:
## 需求:
根据我们 [Windows](develop-on-windows-zh.md)/[Mac OS](develop-on-mac-zh.md) 的配置指南来搭建本地环境。

## 概览:
我们提供了一个 `Docker` 开发镜像环境来简化配置步骤以及在不同平台提供一个一致的环境。

## 构建
### 1. 直接构建 ***deviceShifu*** 二进制文件:
导航到 `Shifu` 的根目录, 使用以下命令来构建不同的 ***deviceShifu***:

`HTTP 转 HTTP` ***deviceShifu***:
```sh
CGO_ENABLED=0 GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) go build -a -o output/deviceshifu-http-http cmd/deviceshifu/cmdHTTP/main.go
```
`HTTP 转 Socket` ***deviceShifu***:
```sh
CGO_ENABLED=0 GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) go build -a -o output/deviceshifu-http-socket cmd/deviceshifu/cmdSocket/main.go
```
`HTTP 转 MQTT` ***deviceShifu***:
```sh
CGO_ENABLED=0 GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) go build -a -o output/deviceshifu-http-mqtt cmd/deviceshifu/cmdMQTT/main.go
```
`HTTP 转 OPC UA` ***deviceShifu***:
```sh
CGO_ENABLED=0 GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) go build -a -o output/deviceshifu-http-opcua cmd/deviceshifu/cmdOPCUA/main.go
```

### 2. 构建 ***deviceShifu*** `Docker` 镜像:
运行下面命令来构建以下不同的 ***deviceShifu*** 镜像, 标记为当前版本并载入到 `Docker` 镜像:
```sh
make buildx-load-image-deviceshifu
```

构建的镜像:
```
edgenesis/deviceshifu-http-http:{VERSION}
edgenesis/deviceshifu-http-socket:{VERSION}
edgenesis/deviceshifu-http-mqtt:{VERSION}
edgenesis/deviceshifu-http-opcua:{VERSION}
```

### 3. 直接构建 ***shifuController*** 的二进制文件:
导航到 `shifu/k8s/crd` 目录下, 用下面的命令来构建  ***shifuController*** 二进制文件:
```sh
make build
```

### 4. 构建 ***shifuController*** `Docker` 镜像:
运行下面命令来构建 ***shifuController*** 镜像, 标记为当前版本并载入到 `Docker` images:
```sh
make docker-buildx-load IMG=edgehub/edgedevice-controller-multi:v0.0.1
```

# 接下来?
跟着我们的 [使用指南](use-shifu-zh.md) 来使用 `Shifu` 。
