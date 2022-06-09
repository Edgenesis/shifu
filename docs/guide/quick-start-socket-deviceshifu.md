**Note**: currently encoding does not work, command will be proxied directly

**Note**: currently edgedevice's status for socket type connection will fail

**Note**: deviceShifu currently will expect a `0x0A` character when receiving from TCP socket. Otherwise you may expect no return from the device.

## To create a socket type deviceShifu, use image:

```
edgehub/deviceshifu-http-socket:v0.0.1
```

### Sample deployment file, all files are available in `/shifu/examples/socketDeviceShifu`:

```
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: edgedevice-socket-deployment
  name: edgedevice-socket-deployment
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: edgedevice-socket-deployment
  template:
    metadata:
      labels:
        app: edgedevice-socket-deployment
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
          value: "edgedevice-socket"
        - name: EDGEDEVICE_NAMESPACE
          value: "devices"
      volumes:
      - name: edgedevice-config
        configMap:
          name: socket-configmap-0.0.1
      serviceAccountName: edgedevice-sa
```

### Sample service file:

```
apiVersion: v1
kind: Service
metadata:
  labels:
    app: edgedevice-socket-deployment
  name: edgedevice-socket
  namespace: default
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 8080
  selector:
    app: edgedevice-socket-deployment
  type: LoadBalancer
```

### sample configmap file:

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: socket-configmap-0.0.1
  namespace: default
data:
  driverProperties: |
    driverSku: testSocket
    driverImage: 
  instructions: |
    cmd:
  telemetries: |
```

### sample edgedevice file:

```
apiVersion: shifu.edgenesis.io/v1alpha1
kind: EdgeDevice
metadata:
  name: edgedevice-socket
  namespace: devices
spec:
  sku: "testSocket" 
  connection: Ethernet
  address: 192.168.15.248:11122 #change this accordingly
  protocol: Socket
  protocolSettings:
    SocketSetting:
      encoding: utf-8
      networkType: tcp
```

## To interact with the device:

```
curl -XPOST -H 'Content-Type: application/json' -d '{"command": "test", "timeout":123}' http://edgedevice-led/cmd  
```

Where `command` is the string being proxied to the actual device

An `\n` character will be appended at the end of the `command` value.

Return from an "echo" server:

```
{"message":"test123\n","status":200}
```

Where `message` is the string returned from the device
