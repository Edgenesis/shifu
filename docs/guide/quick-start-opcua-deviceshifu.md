**Note**: currently ***edgedevice***'s status for OPC UA type connection will fail

**Note**: currently we do not support password/x509 authentication

**Note**: currently all values returned from the OPC UA server/device will be written into the HTTP body as `string`

## To create a OPC UA type deviceShifu, use image:

```
edgehub/deviceshifu-http-opcua:v0.0.1
```

### Sample deployment file, all files are available in `/shifu/examples/opcuaDeviceShifu`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: deviceshifu-opcua-deployment
  name: deviceshifu-opcua-deployment
  namespace: deviceshifu
spec:
  replicas: 1
  selector:
    matchLabels:
      app: deviceshifu-opcua-deployment
  template:
    metadata:
      labels:
        app: deviceshifu-opcua-deployment
    spec:
      containers:
      - image: edgehub/deviceshifu-http-opcua:v0.0.1
        name: deviceshifu-http
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: deviceshifu-config
          mountPath: "/etc/edgedevice/config"
          readOnly: true
        env:
        - name: EDGEDEVICE_NAME
          value: "edgedevice-opcua"
        - name: EDGEDEVICE_NAMESPACE
          value: "devices"
      volumes:
      - name: deviceshifu-config
        configMap:
          name: opcua-configmap-0.0.1
      serviceAccountName: edgedevice-sa
```

### Sample service file:

```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    app: deviceshifu-opcua-deployment
  name: deviceshifu-opcua
  namespace: deviceshifu
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 8080
  selector:
    app: deviceshifu-opcua-deployment
  type: LoadBalancer
```

### sample configmap file:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: opcua-configmap-0.0.1
  namespace: deviceshifu
data:
  driverProperties: |
    driverSku: Test OPC UA Server
    driverImage: 
  instructions: |
    get_value:
      protocolPropertyList:
        OPCUANodeID: "ns=2;i=2"
    get_time:
      protocolPropertyList:
        OPCUANodeID: "i=2258"
    get_server:
      protocolPropertyList:
        OPCUANodeID: "i=2261"
  telemetries: |
    device_health:
      properties:
        instruction: get_server
        initialDelayMs: 1000
        intervalMs: 1000
```

- Each `instruction` should have an `OPCUANodeID`, this represents the `API` to `Node ID` mapping

### sample edgedevice file:

```yaml
apiVersion: shifu.edgenesis.io/v1alpha1
kind: EdgeDevice
metadata:
  name: edgedevice-opcua
  namespace: devices
spec:
  sku: "opcua-test" 
  connection: Ethernet
  address:  opc.tcp://192.168.0.111:4840/freeopcua/server #change this accordingly
  protocol: OPCUA
  protocolSettings:
    OPCUASetting:
      SecurityMode: None
      ConnectionTimeoutInMilliseconds: 5000
```

- Currently we use the `spec.address` field for endpoint address

## To interact with the device:

```bash
root@nginx:/# curl deviceshifu-opcua/get_server;echo
FreeOpcUa Python Server
```

- `get_server` is the instrunction defined in `ConfigMap`, which has a Node ID of `i=2261`

- Therefore the request to `edgedevice-opcua`'s `get_server` API translates to `OPC UA`'s client request to "`opc.tcp://192.168.0.111:4840/freeopcua/server`"'s `Node ID` of "`i=2261`"

- The server returns the following string to the ***edgeDevice***:
`"FreeOpcUa Python Server"`
