## 和 *deviceShifu* 交互的简单应用
*Shifu* 会对每一个连接的设备创建一个 *deviceShifu*.

*deviceShifu* 是物理设备的数字孪生并负责控制，收集设备指标。

这个教程会创建一个简单的温度检测程序，通过和一个温度计的 *deviceShifu* 交互来演示如何和用应用来和 *deviceShifu* 交互。

### 前提
本示例需要安装 [Go](https://golang.org/dl/), [Docker](https://docs.docker.com/get-docker/), [kind](https://kubernetes.io/docs/tasks/tools/), [kubectl](https://kubernetes.io/docs/tasks/tools/) 和 [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)。

### 1. 运行 *Shifu* 并连接一个简单的温度计
在 `shifu/examples/deviceshifu/demo_device` 路径中已经有一个演示温度计的 deployment 配置。该温度计会上报一个整数代表当前温度，它拥有一个 `read_value` API 来汇报这个数值。

在 `shifu` 根目录下，运行下面两条命令来运行 *shifu* 和演示温度计的 *deviceShifu*：

```bash
./test/scripts/deviceshifu-setup.sh apply         # setup and start shifu services for this demo
kubectl apply -f examples/deviceshifu/demo_device/edgedevice-thermometer    # connect fake thermometer to shifu
```
### 2. 温度检测程序
本应用会通过 HTTP 请求来和 *deviceShifu* 交互，每两秒检测 `read_value` 节点来获取温度计 *deviceShifu* 的读数。

应用示例如下：

**high-temperature-detector.go**
```
package main

import (
	"log"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func main() {
	targetUrl := "http://edgedevice-thermometer/read_value"
	req, _ := http.NewRequest("GET", targetUrl, nil)
	for {
		res, _ := http.DefaultClient.Do(req)
		body, _ := ioutil.ReadAll(res.Body)
		temperature, _ := strconv.Atoi(string(body))
		if temperature > 20 {
			log.Println("High temperature:", temperature)
		} else if temperature > 15 {
			log.Println("Normal temperature:", temperature)
		} else {
			log.Println("Low temperature:", temperature)
		}
		res.Body.Close()
		time.Sleep(2 * time.Second)
	}
}
```

生成 go.mod:
```
go mod init high-temperature-detector
```
### 3. 容器化应用
需要一个应用程序的 `Dockerfile`：

**Dockerfile**
```
# syntax=docker/dockerfile:1

FROM golang:1.17-alpine
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY *.go ./
RUN go build -o /high-temperature-detector
EXPOSE 11111
CMD [ "/high-temperature-detector" ] 
```

之后，创建应用：

```
docker build --tag high-temperature-detector:v0.0.1 .
```

现在温度检测应用的镜像已经构建完成。

### 4. 加载应用镜像并启动应用 Pod

首先将应用镜像加载到 `kind` 中：
```
kind load docker-image high-temperature-detector:v0.0.1
```

之后运行容器 Pod：
```
kubectl run high-temperature-detector --image=high-temperature-detector:v0.0.1 -n deviceshifu
```

### 5. 检查应用输出

温度检测应用会每两秒钟通过温度计的 *deviceShifu* 获取当前数值。

一切准备就绪，通过 log 来查看程序输出：

```
kubectl logs -n default high-temperature-detector -f
```

输出示例：
```
kubectl logs -n default high-temperature-detector -f

2021/10/18 10:35:35 High temperature: 24
2021/10/18 10:35:37 High temperature: 23
2021/10/18 10:35:39 Low temperature: 15
2021/10/18 10:35:41 Low temperature: 11
2021/10/18 10:35:43 Low temperature: 12
2021/10/18 10:35:45 High temperature: 28
2021/10/18 10:35:47 Low temperature: 15
2021/10/18 10:35:49 High temperature: 30
2021/10/18 10:35:51 High temperature: 30
2021/10/18 10:35:53 Low temperature: 15


