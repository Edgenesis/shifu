# Getting started: connect a PLC
Shifu's Siemens S7 Suite provides the ability to edit the PLC memory area through HTTP requests. This article will provide an example on connecting a Siemens S7-1200 1214C PLC to Shifu and using Shifu to interact with it.

## Connecting S7 PLC
Before connected to Shifu, an S7 PLC needs to be physically connected over Ethernet with an IP address, for example, 192.168.0.1. 


In this scenario, Shifu needs the deployment config to tell Shifu the address and type of the device:  
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

Other configuration files should also be prepared with necessary information:  
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

Add the PLC device to Shifu to start:
```
kubectl apply -f plc_configuration_directory
```

## Operations
Shifu provides the following HTTP APIs to edit memory area.  

**sendsinglebit**:  
Edit a single bit of a given memory area. It takes the following parameters:
- **rootaddress**: the name of the root memory area, e.g., M for Merker, Q for Digital Output Process Image, etc.
- **address**: the address in the memory area.
- **start**: the starting position in the address.
- **digit**: the nth digit of bit to be edited.
- **value**: the value to be set to the nth digit of bit.

For example, `plc-device/sendsinglebit?rootaddress=M&address=0&start=2&digit=2&value=1` will make M0.2's second digit to 1.  

**getcontent**:  
Get the current value of a given memory area in byte. It takes the following parameters:  
- **rootaddress**: the name of the root memory area, e.g., M for Merker, Q for Digital Output Process Image, etc.
- **address**: the address in the memory area.
- **start**: the starting position in the address.

For example, `plc-device/sendsinglebit?rootaddress=M&address=0&start=2` will get M0.2's value.

**getcpuordercode**:  
Get the S7 PLC static information.
