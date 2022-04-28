# 快速上手：DeviceShifu

本文将通过一个简单的“helloworld”虚拟设备的例子，来展示如何向Shifu接入设备，生成DeviceShifu，并进行操作。


## Helloworld 设备
Helloworld设备只有一个功能：每次收到请求时，返回“hello world”信息。

### 步骤
1. ### 准备虚拟设备
   本次要创建的虚拟设备是一个软件应用，它每次收到HTTP GET请求时，都会返回“Hello_world from device via shifu!” 这条信息。另外，我们还将使用Shifu的数据收集功能对这条信息进行每秒一次的自动收集。

   在开发路径中，创建一个`helloworld.go`文件，包含如下内容:
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
    
   生成 go.mod 文件:
   ```
   go mod init helloworld
   ```

   添加 Dockerfile:
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
   创建镜像

   ```
   docker build --tag helloworld-device:v0.0.1 .
   ```

2. ### 准备 EdgeDevice: 
   EdgeDevice的基本信息:  
   假设所有配置文件都保存在 `<working_dir>/helloworld-device/configuration`

   Deployment 配置:\
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
   
   硬件和连接信息:\
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

    Service:\
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
3. ### 准备DeviceShifu
   使用下面的配置文件，Shifu将自动生成DeviceShifu的Pod。  
   假设所有配置文件都保存在 `<working_dir>/helloworld-device/configuration`.

   ConfigMap:\
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
   Deployment:\
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
    Service:\
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

4. ### 创建 DeviceShifu
   下面的步骤都需要要求Shifu平台已经启动并且正在运行, 参见 [快速上手：安装](./getting_started_installation.md)
   1. 加载刚刚构建完成的docker镜像
       ```
       kind load docker-image helloworld-device:v0.0.1
       ```
   2. 让Shifu通过配置创建DeviceShifu的Pod
       ```
       kubectl apply -f <working_dir>/helloworld-device/configuration
       ```
   3. 启动一个nginx的服务器
       ```
       kubectl run nginx --image=nginx:1.21
       ```
      现在集群中应当有以下Pod:
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
       查看创建的EdgeDevice：

        ```
        kubectl get edgedevice --namespace devices edgedevice-helloworld

        NAME                    AGE
        edgedevice-helloworld   22m
        ```
        DeviceShifu的细节信息以及状态可以通过 `describe` 命令获取： 
        ```
        kubectl describe edgedevice --namespace devices edgedevice-helloworld
        ```

   4. 使用nginx：
       ```
       kubectl exec -it --namespace default nginx -- bash
       ```
   5. 与DeviceShifu进行交互：
      ```
      /# curl http://edgedevice-helloworld-service:80/hello
      ```

      应该得到以下输出:
      ```
      Hello_world from device via shifu!
      ```
   6. 在日志中查看收集到的数据:
      ```
      kubectl logs edgedevice-helloworld-deployment-6464b55979-hbdhr
      ```
现在 helloworld 设备已经完全整合到Shifu框架中，可以通过上述方式来通过DeviceShifu与其交互。
   
   ***如果需要更新configuration，请先delete再apply configurtaion:***

      /# kubectl delete -f <working_dir>/helloworld-device/configuration
      /# kubectl apply -f <working_dir>/helloworld-device/configuration