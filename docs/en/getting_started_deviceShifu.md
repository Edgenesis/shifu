# Getting started: DeviceShifu

This article will go through the steps to build a virtual IoT device "helloworld", and connect it to Shifu, thus creating a new DeviceShifu. 


## Helloworld device
The helloworld device only does one job: respond hello world message upon request. 

### Steps
1. ### Prepare the EdgeDevice:  Docker image
   The expected EdgeDevice is an application that response "Hello_world from device via shifu!" every time it receives an HTTP GET request. Additionally, the telemetry collection feature is used to collect this response every second.

   In the working directory, create a `helloworld.go` with following content:
   ```
    package main
    import (
      "fmt"
      "net/http"
    )

    func process_hello(w http.ResponseWriter, req *http.Request) {
      fmt.Fprintln(w, "Hello_world from device via shifu!")
    }

    func headers(w http.ResponseWriter, req *http.Request) {
      for name, headers := range req.Header {
        for _, header := range headers {
          fmt.Fprintf(w, "%v: %v\n", name, header)
        }
      }
    }

    func main() {
      http.HandleFunc("/hello", process_hello)
      http.HandleFunc("/headers", headers)

      http.ListenAndServe(":11111", nil)
    }
    ```
    
   Generate the go.mod file:
   ```
   go mod init helloworld
   ```

   Add its Dockerfile:
   ```
   # syntax=docker/dockerfile:1

   FROM golang:1.17-alpine
   WORKDIR /app
   COPY go.mod ./
   RUN go mod download
   COPY *.go ./
   RUN go build -o /helloworld
   EXPOSE 11111
   CMD [ "/helloworld" ]    
   ```
   Build the image

   ```
   docker build --tag helloworld-device:v0.0.1 .
   ```

2. ### Prepare the configuration for the EdgeDevice: 
   The basic information of the EdgeDevice:  
   Assuming all configurations are saved in `<working_dir>/helloworld-device/configuration`

   Deployment for the EdgeDevice:\
   **helloworld-deployment.yaml**
    ```
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      labels:
        app: helloworld
      name: helloworld
      namespace: devices
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: helloworld
      template:
        metadata:
          labels:
            app: helloworld
        spec:
          containers:
            - image: helloworld-device:v0.0.1
              name: helloworld
              ports:
                - containerPort: 11111
      ```
   
   Hardware and connection info for the EdgeDevice:\
    **helloworld-edgedevice.yaml**
    ```
    apiVersion: shifu.edgenesis.io/v1alpha1
    kind: EdgeDevice
    metadata:
      name: edgedevice-helloworld
      namespace: devices
    spec:
      sku: "Hello World"
      connection: Ethernet
      address: helloworld.devices.svc.cluster.local:11111
      protocol: HTTP
    status:
      edgedevicephase: "Pending"
    ```

    Service for the EdgeDevice:\
    **helloworld-service.yaml**
   ```
   apiVersion: v1
   kind: Service
   metadata:
     labels:
       app: helloworld
     name: helloworld
     namespace: devices
   spec:
     ports:
       - port: 11111
         protocol: TCP
         targetPort: 11111
     selector:
       app: helloworld
     type: LoadBalancer
   ```
3. ### Prepare the configurations for Shifu to create the DeviceShifu
   With the following configurations, Shifu is able to create a DeviceShifu automatically for the device.\
   Assuming all configurations are saved in `<working_dir>/helloworld-device/configuration`.

   ConfigMap for the DeviceShifu:\
   **deviceshifu-helloworld-configmap.yaml**
    ```
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: helloworld-configmap-0.0.1
      namespace: default
    data:
    #    device name and image address
      driverProperties: |
        driverSku: Hello World
        driverImage: helloworld-device:v0.0.1
    #    available instructions
      instructions: |
        hello:
    #    telemetry retrieval methods
    #    in this example, a device_health telemetry is collected by calling hello instruction every 1 second
      telemetries: |
        device_health:
          properties:
            instruction: hello
            initialDelayMs: 1000
            intervalMs: 1000
   ```
   Deployment for the DeviceShifu:\
   **deviceshifu-helloworld-deployment.yaml**
    ```
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      labels:
        app: edgedevice-helloworld-deployment
      name: edgedevice-helloworld-deployment
      namespace: default
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: edgedevice-helloworld-deployment
      template:
        metadata:
          labels:
            app: edgedevice-helloworld-deployment
        spec:
          containers:
            - image: edgehub/deviceshifu-http:v0.0.1
              name: deviceshifu-http
              ports:
                - containerPort: 8080
              volumeMounts:
                - name: edgedevice-config
                  mountPath: "/etc/edgedevice/config"
                  readOnly: true
              env:
                - name: EDGEDEVICE_NAME
                  value: "edgedevice-helloworld"
                - name: EDGEDEVICE_NAMESPACE
                  value: "devices"
          volumes:
            - name: edgedevice-config
              configMap:
                name: helloworld-configmap-0.0.1
          serviceAccountName: edgedevice-mockdevice-sa   
   ```
    Service for the DeviceShifu:\
    **deviceshifu-helloworld-service.yaml**
    ```
    apiVersion: v1
    kind: Service
    metadata:
      labels:
        app: edgedevice-helloworld-deployment
      name: edgedevice-helloworld-service
      namespace: default
    spec:
      ports:
        - port: 80
          protocol: TCP
          targetPort: 8080
      selector:
        app: edgedevice-helloworld-deployment
      type: LoadBalancer
   ```

4. ### Create new DeviceShifu
   The following steps require Shifu already started and running, which is covered in [Getting started: installation](./getting_started_installation.md)
   1. load the docker image of the helloworld device
       ```
       kind load docker-image helloworld-device:v0.0.1
       ```
   2. let Shifu create the DeviceShifu from the configurations
       ```
       kubectl apply -f <working_dir>/helloworld-device/configuration
       ```
   3. start a nginx server
       ```
       kubectl run nginx --image=nginx:1.21
       ```
      So far, the following pods should be generated:
        ```
        kubectl get pods --all-namespaces
        NAMESPACE            NAME                                                READY   STATUS    RESTARTS   AGE
        crd-system           crd-controller-manager-7bc78896b9-sq72b             2/2     Running   0          28m
        default              edgedevice-helloworld-deployment-6464b55979-hbdhr   1/1     Running   0          27m
        default              nginx                                               1/1     Running   0          8s
        devices              helloworld-5f467bf5db-f5hxh                         1/1     Running   0          25m
        kube-system          coredns-558bd4d5db-qqx92                            1/1     Running   0          30m
        kube-system          coredns-558bd4d5db-zlw8b                            1/1     Running   0          30m
        kube-system          etcd-kind-control-plane                             1/1     Running   0          30m
        kube-system          kindnet-ndrnh                                       1/1     Running   0          30m
        kube-system          kube-apiserver-kind-control-plane                   1/1     Running   0          30m
        kube-system          kube-controller-manager-kind-control-plane          1/1     Running   0          30m
        kube-system          kube-proxy-qkswm                                    1/1     Running   0          30m
        kube-system          kube-scheduler-kind-control-plane                   1/1     Running   0          30m
        local-path-storage   local-path-provisioner-547f784dff-44xnv             1/1     Running   0          30m
       ```
       Check the EdgeDevice instance created:

        ```
        kubectl get edgedevice --namespace devices edgedevice-helloworld

        NAME                    AGE
        edgedevice-helloworld   22m
        ```
       Get detailed EdgeDevice connection and status information by calling `describe` on the pod: 
        ```
        kubectl describe edgedevice --namespace devices edgedevice-helloworld
        ```

   4. get into the nginx shell
       ```
       kubectl exec -it --namespace default nginx -- bash
       ```
   5. interact with the Hellow World EdgeDevice via its DeviceShifu
      ```
      /# curl http://edgedevice-helloworld-service:80/hello
      ```

      The response should be:
      ```
      Hello_world from device via shifu!
      ```
   6. check the telemetry reported by the DeviceShifu:
      ```
      kubectl logs edgedevice-helloworld-deployment-6464b55979-hbdhr
      ```
Now the Hello World EdgeDevice is fully integrated in the Shifu framework and we can interact with it via the DeviceShifu as shown above.
   
   ***To update the configmap, Shifu currently requires delete and re-apply the configuration:***

      /# kubectl delete -f <working_dir>/helloworld-device/configuration
      /# kubectl apply -f <working_dir>/helloworld-device/configuration