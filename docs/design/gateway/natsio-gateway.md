# nats Gateway Design

nats Gateway is a component that allows deviceShifu to publish data to nats server and subscribe to nats server to call deviceShifu instructions to handle it.

## Goal

### Design Goal

- Enable the device to act as an nats client.
- Provide a method to connect deviceShifu to the nats server.
- Allow deviceShifu to publish data which from device to nats server with interval.
- Allow deviceShifu to subscribe to nats server and call deviceShifu instructions to handle it.

### Non-Goal

- Support for all features in the nats protocol.
- Provide all authentication and authorization features.
- Integrate nats Server into deviceShifu.


## nats Gateway Design

The nats Gateway is a component that allows deviceShifu to publish data to nats server and subscribe to nats server to call deviceShifu instructions to handle it.

```mermaid
flowchart BT

device[Device]

subgraph EdgeNode
    subgraph Shifu
        subgraph deviceShifu
            gw[nats Gateway]
            ds[deviceShifu]

            ds -->|HTTP| gw
            gw -->|HTTP| ds
        end
    end
end

natsclient[nats Client]
nats[nats Server]

gw <-->|nats| nats
natsclient <-->|nats| nats

device <-->|Device Protocol| ds
```

### What Does the Gateway Do?

1. Start the nats client and obtain all device information from deviceShifu.
2. Subscribe to nats server and set callback function to handle the message from target topic.
3. Create a new thread for each publisher topic to publish data to nats server with interval.


## Detail Design

### Instruction Config

The instruction config is a configmap that contains the instruction name and the instruction config.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: configmap
  namespace: deviceShifu
data:
  instructions: |
    instructions:
      instruction1:
        gatewayPropertyList:
          Topic: "testTopic1"
          Mode: "publisher"
          PublisherIntervalMs: 1000
```

- Topic: The topic to publish data to nats server.
- Mode: The mode of the instruction. The mode can be "publisher" or "subscriber".
  - "publisher": The instruction is a publisher which publish data to nats server with interval.
  - "subscriber": The instruction is a subscriber which subscribe to nats server and call deviceShifu instructions to handle it.
- PublisherIntervalMs: The interval to publish data to nats server. If the mode is "publisher", the interval is the interval to publish data to nats server.


### Subscribe to nats

```mermaid
sequenceDiagram
    participant device
    participant deviceShifu
    participant natsGateway
    participant natsServer

    natsGateway->>natsServer: subscribe(topic)
    loop receive message from natsServer
        natsServer->>natsGateway: message
        natsGateway->>deviceShifu: callback(message)
        deviceShifu->>device: handle(message)
        device->>deviceShifu: response(message)
        deviceShifu->>natsGateway: response(message): OK
    end

```

After the nats Gateway starts, it will automatically subscribe the instruction which is set in the deviceShifu ConfigMap. and then the nats Gateway will set the callback function to handle the message from target topic. After the callback function is called, the nats Gateway will call the deviceShifu instructions to handle it to call the deviceShifu instructions to handle it.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: configmap
  namespace: deviceShifu
data:
  instructions: |
    instructions:
      instruction1:
        gatewayPropertyList:
          Topic: "testTopic1"
          Mode: "subscriber"
```


### Publish to nats

```mermaid
sequenceDiagram
    participant device
    participant deviceShifu
    participant natsGateway
    participant natsServer

    loop publish data to nats server(interval: 1000ms)
        natsGateway -->> deviceShifu: get data from device
        deviceShifu -->> device: get data from device
        device -->> deviceShifu: data
        deviceShifu -->> natsGateway: data
        natsGateway -->> natsServer: publish(topic, data)
        natsGateway -->> natsGateway: wait for 1000ms
    end
```


After the nats Gateway starts, it will create a new thread for each publisher topic to publish data to nats server with interval. By default, the interval is 1 second. The interval can be set in the deviceShifu ConfigMap.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: configmap
  namespace: deviceShifu
data:
  instructions: |
    instructions:
      instruction1:
        gatewayPropertyList:
          Topic: "testTopic1"
          Mode: "publisher"
          PubGetIntervalMs: 1000
```

## Test Plan

- Publish data to nats server with interval.
  - Test the nats Gateway to publish data to nats server with interval.
  - Test the nats Gateway to publish data to nats server with 1000ms interval.
  - Test the nats Gateway to publish data to nats server with 100ms interval.
  - Test the nats Gateway to publish data to nats server with 10ms interval.
  - Test the nats Gateway to publish data to nats server with 1ms interval.

- Subscribe to nats server and call deviceShifu instructions to handle it.
  - Test the nats Gateway to subscribe to nats server and call deviceShifu instructions to handle it.
  - Test the nats Gateway to handle the message from nats server with 1000ms interval.
  - Test the nats Gateway to handle the message from nats server with 100ms interval.
  - Test the nats Gateway to handle the message from nats server with 10ms interval.
  - Test the nats Gateway to handle the message from nats server with 1ms interval.

