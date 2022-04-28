# 快速上手: 接入一个PLC
Shifu对西门子S7系列PLC提供了兼容。用户可以使用Shifu，通过HTTP请求对S7 PLC的内存进行修改。本文将介绍如何接入一台西门子S7-1200 1214C PLC并且与它交互。

## 连接
在接入Shifu之前，PLC应当已经通过以太网完成物理连接，并且拥有一个IP地址。这里我们使用192.168.0.1。

Shifu需要如下例所示的配置文件来获取IP地址与设备类型：  
**plc-deployment.yaml**
```
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: plc
  name: plc
  namespace: devices
spec:
  replicas: 1
  selector:
    matchLabels:
      app: plc
  template:
    metadata:
      labels:
        app: plc
    spec:
      containers:
        - image: edgehub/plc-device:v0.0.1
          name: plc
          ports:
            - containerPort: 11111
          env:
            - name: PLC_ADDR
              value: "192.168.0.1"
            - name: PLC_RACK
              value: "0"        
            - name: PLC_SLOT
              value: "1"
```

同时，Shifu还需要一些通用的配置文件:  
**deviceshifu-plc-configmap.yaml**
```
apiVersion: v1
kind: ConfigMap
metadata:
  name: plc-configmap-0.0.1
  namespace: default
data:
#    device name and image address
  driverProperties: |
    driverSku: PLC
    driverImage: plc-device:v0.0.1
    driverExecution: " "
#    available instructions
  instructions: |
    sendsinglebit:
    getcpuordercode:
#    telemetry retrieval methods
#    in this example, a device_health telemetry is collected by calling hello instruction every 1 second
  telemetries: |
    device_health:
      properties:
        instruction: getcpuordercode
        initialDelayMs: 1000
        intervalMs: 1000
```
**deviceshifu-plc-deployment.yaml**
```
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: edgedevice-plc-deployment
  name: edgedevice-plc-deployment
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: edgedevice-plc-deployment
  template:
    metadata:
      labels:
        app: edgedevice-plc-deployment
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
              value: "edgedevice-plc"
            - name: EDGEDEVICE_NAMESPACE
              value: "devices"
      volumes:
        - name: edgedevice-config
          configMap:
            name: plc-configmap-0.0.1
      serviceAccountName: edgedevice-sa
```
**deviceshifu-plc-service.yaml**
```
apiVersion: v1
kind: Service
metadata:
  labels:
    app: edgedevice-plc-deployment
  name: edgedevice-plc-service
  namespace: default
spec:
  ports:
    - port: 80
      protocol: TCP
      targetPort: 8080
  selector:
    app: edgedevice-plc-deployment
  type: LoadBalancer
  ```
**plc-edgedevice.yaml**
```
apiVersion: shifu.edgenesis.io/v1alpha1
kind: EdgeDevice
metadata:
  name: edgedevice-plc
  namespace: devices
spec:
  sku: "PLC"
  connection: Ethernet
  address: plc.devices.svc.cluster.local:11111
  protocol: HTTP
status:
  edgedevicephase: "Pending"
```
**plc-service.yaml**
```
apiVersion: v1
kind: Service
metadata:
  labels:
    app: plc
  name: plc
  namespace: devices
spec:
  ports:
    - port: 11111
      protocol: TCP
      targetPort: 11111
  selector:
    app: plc
  type: LoadBalancer
```

向Shifu添加PLC设备，创建和启动DeviceShifu:
```
kubectl apply -f plc_configuration_directory
```

## 操作
Shifu支持通过HTTP请求来编辑PLC内存。  

**sendsinglebit**:  
修改一个bit，它需要下列参数:
- **rootaddress**: 内存区域名称，比如M代表Merker，Q代表Digital Output。
- **address**: 内存区域中的地址。
- **start**: 开始位置。
- **digit**: 从开始位置起第几个bit。
- **value**: 需要修改成为的数值.

比如，`plc-device/sendsinglebit?rootaddress=M&address=0&start=2&digit=2&value=1` 会将 M0.2 的第二个 bit 修改为1.  

**getcontent**:  
得到内存区域中一个byte的值，它需要下列参数:  
- **rootaddress**: 内存区域名称，比如M代表Merker，Q代表Digital Output。
- **address**: 内存区域中的地址。
- **start**: 开始位置。

比如 `plc-device/sendsinglebit?rootaddress=M&address=0&start=2` 会返回 M0.2 的一个 byte 的值.

**getcpuordercode**:  
得到PLC的静态信息。
