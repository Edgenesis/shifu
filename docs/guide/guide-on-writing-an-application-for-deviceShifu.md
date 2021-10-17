## Interact with *deviceShifu*
*Shifu* automatically creates *deviceShifu* for each device connected. \
*deviceShifu* serves as a digital twin of the physical device and is responsible of controlling the device and collecting the device updates.

This instruction will show how to build and use an application to interact with the *deviceShifu*, by building a simple high temperature detector application to interact with the *deviceShifu* of a thermometer.

### 1. Start *shifu* and connect a simple thermometer device
A fake thermometer which produces a integer value representing the temperature read should be already in the `shifu/deviceshifu/examples/demo_device` directory.\
The device driver has an API `read_value` which returns such integer value.
Under `shifu` root directory, we can run the following three commands to have *shifu* and the fake thermometer *deviceShifu* ready:
```
./test/scripts/deviceshifu-setup.sh apply                          # load images and start cluster
./test/scripts/deviceshifu-sample.sh apply                         # start shifu services
./test/scripts/deviceshifu-demo.sh apply edgedevice-thermometer    # connect fake thermometer to shifu
```
### 2. High temperature detector application
The application interacts with *deviceShifu* via HTTP requests.\
Every 2 seconds, it will check the `read_value` endpoint and get the value from the thermometer *deviceShifu*. 

Here is what this application looks like:

**high-temperature-detector.go**
```
package main

import (
	"fmt"
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
			fmt.Println("High temperature:", temperature)
		} else if temperature > 15 {
			fmt.Println("Normal temperature:", temperature)
		} else {
			fmt.Println("Low temperature:", temperature)
		}
		res.Body.Close()
		time.Sleep(2 * time.Second)
	}
}
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

### 4. Prepare the manifest of the application pod

Kubernetes requires us to prepare the manifest for the application pod.

**high-temperature-detector.yaml**
```
apiVersion: v1
kind: Pod
metadata:
  name: high-temperature-detector
spec:
  containers:
  - image: high-temperature-detector:v0.0.1
    name: high-temperature-detector
```

### 5. Load the application image and start the application pod

First, we need to let kind load the the application image:
```
kind load docker-image high-temperature-detector:v0.0.1
```
Then we can apply the manifest and start the application pod:
```
kubectl apply -f high-temperature-detector.yaml
```

### 6. Use the application

The high temperature detector application gets the value from the thermometer *deviceShifu* every 2 seconds.\
With everything is ready, you can start the application now:
```
kubectl exec -it --namespace default high-temperature-detector -- sh  -c "/high-temperature-detector"
```
An example output will be like this:
```
high-temperature-detector % kubectl exec -it --namespace default high-temperature-detector -- sh  -c "/high-temperature-detector"

High temperature: 25
Low temperature: 15
High temperature: 29
Low temperature: 11
High temperature: 30
Normal temperature: 16
High temperature: 27
High temperature: 28
Low temperature: 15
Normal temperature: 17
Normal temperature: 20
High temperature: 25
High temperature: 30
High temperature: 25
Normal temperature: 16
Normal temperature: 17
Normal temperature: 17
Normal temperature: 18
Low temperature: 10
High temperature: 21
High temperature: 22
High temperature: 27
Low temperature: 11
Normal temperature: 19
High temperature: 24
High temperature: 26
Low temperature: 13
High temperature: 29
High temperature: 21
Normal temperature: 16
High temperature: 27
High temperature: 23
```
