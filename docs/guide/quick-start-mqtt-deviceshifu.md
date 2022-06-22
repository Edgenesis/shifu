
### To create a MQTT type deviceShifu, use image:

```
edgehub/deviceshifu-http-mqtt:v0.0.1
```

### Sample deployment file, all files are available in `/shifu/examples/mqttDeviceShifu`:

```
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: deviceshifu-mqtt-deployment
  name: deviceshifu-mqtt-deployment
  namespace: deviceshifu
spec:
  replicas: 1
  selector:
    matchLabels:
      app: deviceshifu-mqtt-deployment
  template:
    metadata:
      labels:
        app: deviceshifu-mqtt-deployment
    spec:
      containers:
      - image: edgehub/deviceshifu-http-mqtt:v0.0.1
        name: deviceshifu-http
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: deviceshifu-config
          mountPath: "/etc/edgedevice/config"
          readOnly: true
        env:
        - name: EDGEDEVICE_NAME
          value: "edgedevice-mqtt"
        - name: EDGEDEVICE_NAMESPACE
          value: "devices"
      volumes:
      - name: deviceshifu-config
        configMap:
          name: mqtt-configmap-0.0.1
      serviceAccountName: edgedevice-sa
```

### Sample service file:

```
apiVersion: v1
kind: Service
metadata:
  labels:
    app: deviceshifu-mqtt-deployment
  name: deviceshifu-mqtt
  namespace: deviceshifu
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 8080
  selector:
    app: deviceshifu-mqtt-deployment
  type: LoadBalancer
```

### sample configmap file:

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: mqtt-configmap-0.0.1
  namespace: deviceshifu
data:
  driverProperties: |
    driverSku: testMQTT
    driverImage: 
  instructions: |
  telemetries: |
    device_health:
      properties:

```

### sample edgedevice file:

```
apiVersion: shifu.edgenesis.io/v1alpha1
kind: EdgeDevice
metadata:
  name: edgedevice-mqtt
  namespace: devices
spec:
  sku: "testMQTT" 
  connection: Ethernet
  address: 192.168.62.231:1883 # change this accordingly
  protocol: MQTT
  protocolSettings:
    MQTTSetting:
      MQTTTopic: /test/test
```

## To get the latest MQTT message from device:

```
curl deviceshifu-mqtt/mqtt_data
```

Where `mqtt_data` is the embedded query string

Return from MQTT deviceShifu:

```
{"mqtt_message":"test2333","mqtt_receive_timestamp":"2022-04-29 08:57:49.9492744 +0000 UTC m=+75.407609501"}
```

Where `mqtt_message` is the latest data string from device, `mqtt_receive_timestamp` is the timestamp when the message was received.
