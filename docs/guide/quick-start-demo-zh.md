# 快速试玩

我们准备了一个完整的包含所有Demo所需要文件的Docker镜像
只需安装[Docker](https://docs.docker.com/get-docker/)，便可以开启你的`Shifu`体验之旅！


1. **启动Docker镜像：**

    ```
    docker run --network host -it -v /var/run/docker.sock:/var/run/docker.sock edgehub/demo-image-alpine:v0.0.1
    ```

2. **建立Kubernetes集群，开启Shifu服务：**

    在shifu根目录执行以下脚本会创建一个包含预定义CRD的Kubernetes集群，以及开启一个最小化的Shifu服务：
    ```
    ./test/scripts/deviceshifu-setup.sh apply
    ```

    上一步完成后，尝试： 
    ```
    kubectl get pod --all-namespaces
    ```

    这时集群中应该有以下运行的Pods：
    ```
    NAMESPACE            NAME                                         READY   STATUS    RESTARTS   AGE
    crd-system           crd-controller-manager-7bc78896b9-cpk7d      2/2     Running   0          11m
    kube-system          coredns-558bd4d5db-khlqs                     1/1     Running   0          13m
    kube-system          coredns-558bd4d5db-w4tvl                     1/1     Running   0          13m
    kube-system          etcd-kind-control-plane                      1/1     Running   0          13m
    kube-system          kindnet-75547                                1/1     Running   0          13m
    kube-system          kube-apiserver-kind-control-plane            1/1     Running   0          13m
    kube-system          kube-controller-manager-kind-control-plane   1/1     Running   0          13m
    kube-system          kube-proxy-g5kbl                             1/1     Running   0          13m
    kube-system          kube-scheduler-kind-control-plane            1/1     Running   0          13m
    local-path-storage   local-path-provisioner-547f784dff-wspb2      1/1     Running   0          13m
    ```

    我们可以通过查看日志来确认运行状态：
    ```
    kubectl --namespace crd-system logs crd-controller-manager-7bc78896b9-cpk7d -c manager
    ```

3. **启动演示的deviceShifu（数字孪生）:**
    
    在`examples/deviceshifu/demo_device`目录下，我们有4个演示的设备来创建 ***deviceShifu***（虚拟设备）。所有的设备都有 `get_status`命令来获取当前设备的状态，如Busy, Error, Idle等
    除了`get_status`，每一台设备有一个自己的命令：
    * **thermometer**: 一个获取当前温度的温度计，命令`read_value`会返回当前温度计的读数
    * **agv**: 一个自动引导车，命令 `get_position`会返回以x, y轴为坐标的设备当前位置
    * **robotarm**: 一个实验室用的机械臂，命令`get_coordinate`会返回机械臂当前的x, y, z轴坐标
    * **plate-reader**: 一个实验室用的酶标仪，命令`get_measurement`会返回每一个样本中光谱分析扫描的结果数值，样本为8*12个正方矩阵排列

    在shifu根目录执行以下命令，以运行4个设备的deviceShifu:
    ```
    kubectl apply -f examples/deviceshifu/demo_device/edgedevice-thermometer
    kubectl apply -f examples/deviceshifu/demo_device/edgedevice-agv
    kubectl apply -f examples/deviceshifu/demo_device/edgedevice-robot-arm
    kubectl apply -f examples/deviceshifu/demo_device/edgedevice-plate-reader
    ```
    通过命令来获取devices域中所有的Pods：
    ```
    kubectl get pod --namespace devices
    ```
    输出应该为如下：
    ```
    NAME                           READY   STATUS    RESTARTS   AGE
    agv-5944698b79-qxdmk           1/1     Running   0          86s
    robotarm-5478f86fc8-s5kmg      1/1     Running   0          85s
    thermometer-6d6d8f759f-4hd6l   1/1     Running   0          28m
    plate-reader-6859f67bc5-htxpp         1/1     Running   0          86s
    ```
    我们可以通过`kubectl describe pods`查看每一个deviceShifu的信息，如：
    ```
    kubectl describe pods edgedevice-thermometer-deployment-b648d5c6c-rf88p
    ```
4. **运行一个nginx服务来与deviceShifu交互:**
    
    如果要和deviceShifu直接交互，我们可以运行一个简易nginx服务，并通过命令 `kubectl exec`来进入Pods里。
    本演示提供了一个nginx镜像，我们可以直接运行该Pod：
    ```
    kubectl run nginx --image=nginx:1.21 -n deviceshifu
    ```
    通过命令来进入nginx Pod的命令行：
    ```
    kubectl -n deviceshifu exec -it nginx -- bash
    ```
    之后，我们可以呼叫每一个deviceShifu内置的命令来查看返回值（每一个deviceShifu的命令定义在该设备的ConfigMap文件中）。
    注意，以下返回值均为随机产生：
    ```
    / # curl http://deviceshifu-thermometer/get_status
    Busy
    / # curl http://deviceshifu-thermometer/read_value
    27
    / # curl http://deviceshifu-agv/get_status
    Busy
    / # curl http://deviceshifu-agv/get_position
    xpos: 54, ypos: 34
    / # curl http://deviceshifu-robotarm/get_status
    Busy
    / # curl http://deviceshifu-robotarm/get_coordinate
    xpos: 55, ypos: 140, zpos: 135
    / # curl http://deviceshifu-plate-reader/get_status
    Idle
    / # curl http://deviceshifu-plate-reader/get_measurement
    0.75 0.50 1.34 0.95 2.79 2.66 2.68 0.59 0.97 0.93 0.70 0.62 
    0.61 1.47 1.68 1.65 1.05 1.59 0.78 2.92 1.22 1.12 2.86 0.29 
    2.15 2.45 1.99 0.36 1.47 0.18 2.47 0.61 2.43 1.53 0.14 2.41 
    2.80 2.49 0.63 2.61 1.09 1.46 0.22 1.99 1.46 2.30 0.51 0.41 
    1.24 0.78 0.34 2.83 2.76 1.89 2.64 1.79 1.24 1.68 2.84 2.92 
    2.09 2.38 0.02 0.47 0.38 1.62 2.65 0.58 2.17 2.70 0.97 2.18 
    1.47 0.66 0.61 0.10 2.91 1.61 0.30 2.21 0.46 1.74 1.62 1.01 
    1.28 2.27 1.04 0.44 2.47 1.83 0.59 2.09 1.30 2.24 2.87 2.78 
    ```
