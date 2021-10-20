# Hello World 设备
本文会通过创建一个简单的***edgeDevice***并通过***deviceShifu*** (数字孪生)接入到***Shifu***来帮助开发者了解***Shifu***是如何运行的。\
***edgeDevice***可以是任意可以通过驱动沟通并执行某些任务的设备。本文例子中的***edgeDevice***可以实现一件事情：回答HTTP路径`/hello`的请求。
### 必要条件
本文中的示例要求用户安装[Go](https://golang.org/dl/), [Docker](https://docs.docker.com/get-docker/), [kind](https://kubernetes.io/docs/tasks/tools/), [kubectl](https://kubernetes.io/docs/tasks/tools/)和[kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)

### 关键词解释
***edgeDevice***:    
 - ***edgeDevice*** 是一个由***Shifu***管理的物理IoT设备

***deviceShifu***:
- ***deviceShifu*** 是***edgeDevice***的数字孪生

***Shifu***:
- ***Shifu*** 是用来管理，调和***deviceShifu***以及所有相关组件的整个边无际OS框架

### 步骤
1. ### 准备 ***edgeDevice***:  Docker 镜像
   ***edgeDevice*** 的功能: 一个以 "Hello_world from device via shifu!"为响应的HTTP服务器\
   在开发路径中，创建一个`helloworld.go`文件包含如下内容:
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
   生成go.mod:
   ```
   go mod init helloworld
   ``` 
   编写 Dockerfile:
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

   可以测试一下，但是本文不包含测试部分的内容
   
   创建镜像

    `docker build --tag helloworld-device:v0.0.1 .`

2. ### 准备***edgeDevice***的配置文件:
   ***edgeDevice***的基本信息：\
   假设所有配置文件保存在 `<working_dir>/helloworld-device/configuration`

   ***edgeDevice***的Deployment:\
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
   
   ***edgeDevice***的硬件和连接信息：\
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

    ***edgeDevice***的Service:\
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
3. ### 为***Shifu***准备***deviceShifu***的配置文件
   通过以下配置文件, ***Shifu***可以自动创建设备的***deviceShifu***

   假设所有配置文件保存在 `<working_dir>/helloworld-device/configuration`

   ***deviceShifu***的ConfigMap:\
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
   ***deviceShifu***的Deployment:\
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
    ***deviceShifu***的Service:\
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

4. ### 启动***Shifu***并建立***deviceShifu***
   现在所有的准备已经就绪，是时候开始启动***Shifu***并连接设备\
   请确保***Shifu***的源代码已经同步到本地并且为当前目录(`cd shifu` 进到 ***Shifu*** 项目的根目录)

   1. 启动***Shifu***服务
       ```
       ./test/scripts/shifu-application-demo-env-setup.sh apply deviceDemo
       ```
   2. 加载刚刚构建完成的docker镜像
       ```
       kind load docker-image helloworld-device:v0.0.1
       ```
   3. 让***Shifu***通过配置创建***deviceShifu***
       ```
       kubectl apply -f <working_dir>/helloworld-device/configuration
       ```
   4. 启动一个 nginx 服务器
       ```
       kubectl run nginx --image=nginx:1.21
       ```
      现在集群中应该有以下Pod：
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
       查看创建的***edgeDevice***：

        ```
        kubectl get edgedevice --namespace devices edgedevice-helloworld

        NAME                    AGE
        edgedevice-helloworld   22m
        ```
       ***edgeDevice***的细节信息以及状态可以通过`describe`命令获取: 
        ```
        kubectl describe edgedevice --namespace devices edgedevice-helloworld
        ```

   5. 使用nginx的shell：
       ```
       kubectl exec -it --namespace default nginx -- bash
       ```
   6. 和 Hellow World ***edgeDevice***通过***deviceShifu***来进行交互：
      ```
      /# curl http://edgedevice-helloworld-service:80/hello
      ```

      应该得到以下输出:
      ```
      Hello_world from device via shifu!
      ```

现在Hello World ***edgeDevice***已经完全整合到***Shifu***框架中，可以通过上述方式来通过***deviceShifu***与其交互
