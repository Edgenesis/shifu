## Simple application to interact with *deviceShifu*
*Shifu* creates *deviceShifu* for each device connected. \
*deviceShifu* serves as a digital twin of the physical device and is responsible of controlling the device and collecting the device telemetries.

This instruction will show how to build and use an application to interact with the *deviceShifu*, by building a simple high temperature detector application to interact with the *deviceShifu* of a thermometer.

### Prerequisite
The following example requires [Go](https://golang.org/dl/), [Docker](https://docs.docker.com/get-docker/), [kind](https://kubernetes.io/docs/tasks/tools/), [kubectl](https://kubernetes.io/docs/tasks/tools/) and [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) installed.

### 1. Start *Shifu* and connect a simple thermometer device
The deployment config for a fake thermometer which produces a integer value representing the temperature read should be already in the `shifu/examples/deviceshifu/demo_device` directory.\
The device driver has an API `read_value` which returns such integer value.
Under `shifu` root directory, we can run the following two commands to have *shifu* and the fake thermometer *deviceShifu* ready:
```bash
./test/scripts/deviceshifu-setup.sh apply         # setup and start shifu services for this demo
kubectl apply -f examples/deviceshifu/demo_device/edgedevice-thermometer    # connect fake thermometer to shifu
```
### 2. High temperature detector application
The application interacts with *deviceShifu* via HTTP requests.\
Every 2 seconds, it will check the `read_value` endpoint and get the value from the thermometer *deviceShifu*. 

Here is what this application looks like:

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
	targetUrl := "http://deviceshifu-thermometer/read_value"
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

and generate the go.mod:
```
go mod init high-temperature-detector
```
### 3. Containerize the application
We will need a `Dockerfile` for this application:

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

After that, we can build the application:

```
docker build --tag high-temperature-detector:v0.0.1 .
```

Now we will have the image of the high temperature detector application.

### 4. Load the application image and start the application pod

First, we need to let kind load the the application image:
```
kind load docker-image high-temperature-detector:v0.0.1
```
Then we can apply the manifest and start the application pod:
```
kubectl run high-temperature-detector --image=high-temperature-detector:v0.0.1 -n deviceshifu
```

### 5. Check the application output

The high temperature detector application gets the value from the thermometer *deviceShifu* every 2 seconds.\
With everything is ready, you can check the logged output now:
```
kubectl logs -n default high-temperature-detector -f
```
An example output will be like this:
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

