# Hello World Device
This section will show you how ***Shifu*** works by creating a simple ***edgeDevice*** and connect to ***Shifu*** with its ***deviceShifu*** (digital twin).\
An ***edgeDevice*** can be anything that performs some certain tasks and can communicate via a driver. The ***edgeDevice*** in this example will only do one thing: responds on HTTP endpoint `/hello`.
### Prerequisite
The following example requires [Go](https://golang.org/dl/), [Docker](https://docs.docker.com/get-docker/), [kind](https://kubernetes.io/docs/tasks/tools/), [kubectl](https://kubernetes.io/docs/tasks/tools/) and [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) installed.

### Glossary
***edgeDevice***:    
 - ***edgeDevice*** is a physical IoT device managed by ***Shifu***.

***deviceShifu***:
- ***deviceShifu*** is an augmented digital twin of the ***edgeDevice***

***Shifu***:
- ***Shifu*** is the underlying framework of Edgenesis OS that controls and reconcile everything related to ***deviceShifu***

### Steps
1. ### Prepare the ***edgeDevice***:  Docker image
   The expected ***edgeDevice***:  is a HTTP server that response "Hello_world from device via shifu!"\
   In your working directory, for example, create a `helloworld.go` with following content:
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
    
   and generate the go.mod file:
   ```
   go mod init helloworld
   ```

   and add its Dockerfile:
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
   You can test it locally but it is not covered here.

   Build the image

   ```
   docker build --tag helloworld-device:v0.0.1 .
   ```

2. ### Prepare the configuration for the ***edgeDevice***: 
   The basic information of the ***edgeDevice***: \
   Assuming all configurations are saved in `<working_dir>/helloworld-device/configuration`

   Deployment for the ***edgeDevice***:\
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
   
   Hardware and connection info for the ***edgeDevice***:\
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

    Service for the ***edgeDevice***:\
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
3. ### Prepare the configurations for ***Shifu*** to create the ***deviceShifu***
   With the following configurations, ***Shifu*** is able to create a ***deviceShifu*** automatically for the device.\
   Assuming all configurations are saved in `<working_dir>/helloworld-device/configuration`.

   ConfigMap for the ***deviceShifu***:\
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
   Deployment for the ***deviceShifu***:\
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
    Service for the ***deviceShifu***:\
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

4. ### Start ***Shifu*** and create ***deviceShifu***
   Now we have everything ready, and it is the time to start ***Shifu*** and connect the device.\
   Assuming the source code of ***Shifu*** is checked out in the working directory (`cd shifu` will go into the ***Shifu*** project root directory).

   1. start ***Shifu*** service
       ```
       ./test/scripts/shifu-application-demo-env-setup.sh apply deviceDemo
       ```
   2. load the docker image we just built
       ```
       kind load docker-image helloworld-device:v0.0.1
       ```
   3. let ***Shifu*** create the ***deviceShifu*** from the configurations
       ```
       kubectl apply -f <working_dir>/helloworld-device/configuration
       ```
   4. start a nginx server
       ```
       kubectl run nginx --image=nginx:1.21
       ```
      Now we should have the following pods:
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
       Check the ***edgeDevice*** instance created:

        ```
        kubectl get edgedevice --namespace devices edgedevice-helloworld

        NAME                    AGE
        edgedevice-helloworld   22m
        ```
       And you can get detailed ***edgeDevice*** connection and status information by `describe` it: 
        ```
        kubectl describe edgedevice --namespace devices edgedevice-helloworld
        ```

   5. get into the nginx shell
       ```
       kubectl exec -it --namespace default nginx -- bash
       ```
   6. interact with the Hellow World ***edgeDevice*** via its ***deviceShifu***
      ```
      /# curl http://edgedevice-helloworld-service:80/hello
      ```

      you should be able to see this:
      ```
      Hello_world from device via shifu!
      ```

Now the Hello World ***edgeDevice*** is fully integrated in the ***Shifu*** framework and we can interact with it via the ***deviceShifu*** as shown above.
   
   ***If you would like to update the configmap of the device, you can delete and re-apply the configuration:***

      /# kubectl delete -f <working_dir>/helloworld-device/configuration
      /# kubectl apply -f <working_dir>/helloworld-device/configuration
