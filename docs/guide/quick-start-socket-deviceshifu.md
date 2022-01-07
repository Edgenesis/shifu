**Note**: currently encoding does not work, command will be proxied directly

**Note**: currently edgedevice's status for socket type connection will fail

## To create a socket type deviceShifu, use image:

```
edgehub/deviceshifu-http-socket:v0.0.1
```

### Sample deployment file:

```
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: edgedevice-led-deployment
  name: edgedevice-led-deployment
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: edgedevice-led-deployment
  template:
    metadata:
      labels:
        app: edgedevice-led-deployment
    spec:
      containers:
      - image: edgehub/deviceshifu-http-socket:v0.0.1
        name: deviceshifu-http
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: edgedevice-config
          mountPath: "/etc/edgedevice/config"
          readOnly: true
        env:
        - name: EDGEDEVICE_NAME
          value: "edgedevice-led"
        - name: EDGEDEVICE_NAMESPACE
          value: "devices"
      volumes:
      - name: edgedevice-config
        configMap:
          name: led-configmap-0.0.1
      serviceAccountName: edgedevice-sa

```

### sample configmap file:

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: led-configmap-0.0.1
  namespace: default
data:
  driverProperties: |
    driverSku: rpi
    driverImage: edgenesis/rpi:v0.0.1
  instructions: |
    cmd:
# Telemetries are configurable health checks of the EdgeDevice
# Developer/user can configure certain instructions to be used as health check
# of the device. In this example, the device_health telemetry is mapped to
# "get_status" instruction, executed every 1000 ms
  telemetries: |
    device_health:
      properties:
        instruction: led_status
        initialDelayMs: 1000
        intervalMs: 1000
```

### sample edgedevice file:

```
apiVersion: shifu.edgenesis.io/v1alpha1
kind: EdgeDevice
metadata:
  name: edgedevice-led
  namespace: devices
spec:
  sku: "rpi" 
  connection: Ethernet
  address: 192.168.14.208:11122
  protocol: Socket
  protocolSettings:
    encoding: utf-8 // currently this does not work
    networkType: tcp // currently we only support TCP socket
```

## To interact with the device:

```
curl -XPOST -H 'Content-Type: application/json' -d '{"command": "test", "timeout":123}' http://edgedevice-led/cmd  
```

Where `command` is the string being proxied to the actual device

Return from an "echo" server:

```
{"message":"test123\n","status":200}
```

Where `message` is the string returned from the device
