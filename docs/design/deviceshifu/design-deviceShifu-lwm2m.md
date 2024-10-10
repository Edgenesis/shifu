# deviceShifu LwM2M Design

deviceShifu LwM2M allow shifu connect device using [LwM2M protocol](https://omaspecworks.org/what-is-oma-specworks/iot/lightweight-m2m-lwm2m/).

This document outlines the design for integrating the [Lightweight Machine to Machine (LwM2M) protocol](https://omaspecworks.org/what-is-oma-specworks/iot/lightweight-m2m-lwm2m/) into the deviceShifu framework to allow for the management of IoT devices.

## Goals

### Features

- Create a deviceShifu-lwm2m type to enable users to connect devices using the LwM2M protocol.
- LwM2M protocol supports both `read` and `write` requests.
- Datagram Transport Layer Security (DTLS) support.
- Resource observation and notification.
- [LwM2M protocol using v1.0.x version](https://www.openmobilealliance.org/release/LightweightM2M/V1_0-20170208-A/OMA-TS-LightweightM2M-V1_0-20170208-A.pdf).
- LwM2M protocol under UDP.
  
### Unsupported Features

- Support LwM2M v1.1.x or later version.
- Over TCP or other protocol.
- Bootstrap Server.

## Detail Design

```mermaid
flowchart LR
subgraph kubernetes[Kubernetes]
  subgraph ds[deviceShifu-LwM2M]
          direction TB
          lws[LwM2M Server]
          HTTP[HTTP Server]
          HTTP -->|call interface| lws
  end
  subgraph ds1[deviceShifu-LwM2M 1]
          direction TB
          lws1[LwM2M Server]
          http1[HTTP Server]
          http1 -->|call interface| lws1
  end
end
ap[application]
d[device]
d1[device1]


ap <-->|RESTful API| HTTP
ap <-->|RESTful API| http1
lws <-->|LwM2M| d
lws1 <-->|LwM2M| d1
```

The `deviceShifu-LwM2M` will represent each LwM2M object as an instruction that can be accessed through a RESTful API. The API will use the PUT method for write operations and the GET method for read operations. 

Supported data formats:

- PlainText
- JSON

deviceShifu-LwM2M's LwM2M server will handle the following functions:
- Register
- Update
- De-register
- Read
- Write
- Observe
- Notify

Each deviceshifu-LwM2M should expose a UDP port for LwM2M communication.

deviceShifu-LwM2M supports two modes: normal and observe mode.
- normal mode: just like the other deviceShifu, when call the instruction, deviceShifu will send Read or Write Request to deviceShifu and return response
- [observe mode](https://guidelines.openmobilealliance.org/object-support/#observe-and-notify-multiple-resources): this mode will set the min and max notify time, then device will notify data when data changed or timeout. deviceShifu will record data, and return the data store in the memory cache when call the instruction. and this kind of mode also support read and write operation.

DeviceShifu will host an LwM2M server, listening on UDP port 3683 (the default LwM2M port), and an HTTP server on port 8080. When an object is in observe mode, the LwM2M server must handle notifications from the device and update the data in the memory cache. If DeviceShifu receives an instruction before the device is registered, it will return an error message.

```mermaid
sequenceDiagram
    participant d as device
    participant ds as deviceShifu
    participant ap as Application
    participant bs as BootStrap Server
    
    alt bootstrap is enable
    d ->> bs: Get Server Info
    bs ->> d: Reply Server Info
    end

    d ->> ds: Register
    ds ->> d: Created

    alt observe mode
    ds ->> d: Observe[Write the Notification Attributes Settings]
    d ->> ds: Success
    ds ->> d: Observe
    loop Device Data Change or Timeout
        d ->> ds: Notify
        ds ->> d: Update Device Data
    end
    else normal mode
        alt read
            ap ->> ds: Read Object
            ds ->> d: Read
            d ->> ds: Data
            ds ->> ap: Data
        else write
            ap ->> ds: Write Data
            ds ->> d: Write Object Data
            d ->> ds: Changed
            ds ->> ap: Changed
        end
    end
```

### Protocol Specification

#### Connection Settings

In EdgeDevice, a new protocol LwM2M will be added, and add a new field `LwM2MSetting` in protocolSettings to store the LwM2M configuration. For deviceshifu-LwM2M is a LwM2M server, so the `address` field is not required.

In LwM2MSetting structure will store the following fields:
- `EndpointName`: the name of the endpoint, which is used to identify the device, deviceShifu only support the device with the same endpoint name.
- `SecurityMode`: the security mode of the device, support `None` and `DTLS`. default is `None`.
- `DTLSMode`: the DTLS mode of the device, support `PSK`. `RPK` ans `X509` are not supported.
- `CipherSuites`: the cipher suites of the device, reference to [IANA](https://www.iana.org/assignments/tls-parameters/tls-parameters.xhtml#tls-parameters-4) for the cipher suites. And at least support [Basic DTLS CipherSuites in Pion](https://github.com/pion/dtls/blob/98a05d681d3affae2d055a70d3273cbb35425b5a/cipher_suite.go#L25-L45)
- `PSKIdentity`: the pre-shared key identity of the device.
- `PSKKey`: the pre-shared key of the device.

```go
type LwM2MSetting struct {
	EndpointName string `json:"endpointName,omitempty"`
  	// +kubebuilder:default="None"
	SecurityMode *SecurityMode `json:"securityMode,omitempty"`
	DTLSMode     *DTLSMode     `json:"dtlsMode,omitempty"`

	CipherSuites []CiperSuite `json:"cipherSuites,omitempty"`
	PSKIdentity  *string      `json:"pskIdentity,omitempty"`
	PSKKey       *string      `json:"pskKey,omitempty"`
}

type DTLSMode string

const (
    DTLSModePSK DTLSMode = "PSK"
    // ...
)

type CiperSuite string

const (
	CiperSuite_TLS_ECDHE_ECDSA_WITH_AES_128_CCM   CiperSuite = "TLS_ECDHE_ECDSA_WITH_AES_128_CCM"
    // ...
)
```

example yaml:

```yaml
apiVersion: shifu.edgenesis.io/v1alpha1
kind: EdgeDevice
metadata:
  name: edgedevice-lwm2m
  namespace: devices
spec:
  sku: "LwM2M Device"
  connection: Ethernet
  address: -- 
  protocol: LwM2M
  protocolSettings:
    LwM2MSettings:
      endpointName: test
      securityMode: DTLS
      dtlsMode: PSK
      cipherSuites:
        - TLS_PSK_WITH_AES_128_CCM_8
      pskIdentity: hint
      pskKey: ABC123
```

#### Instruction Properties

Since LwM2M is Object and Resource-based, so the instruction properties will store the ObjectID  which is used to identify the data in the device to make the instruction is mapped to the device's Object and Resource.

If enable Observe mode, the instruction properties will store the `EnableObserve` field to enable the observe mode. deviceShifu will cache the data when the device notify the data.

```go
type LwM2MProtocolProperty struct {
    ObjectId            string      `json:"ObjectId"` // required example /3303/0
    EnableObserve       bool        `json:"EnableObserve,omitempty"` // optional enable observe mode default false
}
```

example yaml:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: deviceshifu-lwm2m
  namespace: deviceshifu
data:
  driverProperties: |
    driverSku: LwM2M Device
    driverImage: lwm2m-device:v0.0.1
  instructions: |
    instructions:
      float:
        protocolPropertyList:
          ObjectId: /3442/0/130
          EnableObserve: false
```

### Serving requests

deviceShifu-LwM2M would take RESTful-style requests just as other deviceShifu do.

LwM2M supports both `GET` and `PUT` requests.

To read data from the device, the method signature looks like:
```
GET lwm2m.device.svc.cluster.local/{instruction_name}
```
```go
LwM2MServer.Read(properties.ObjectId)
if properties.EnableObserver {
    cache.Update(properties.ObjectId, newValue)
}
return value
```

To write data to the device, the method signature looks like:
```
PUT lwm2m.device.svc.cluster.local/{instruction_name}
```
```go
LwM2MServer.Write(properties.ObjectId, newValue)
if properties.EnableObserver {
    cache.Update(properties.ObjectId, newValue)
}
return success
```

For the LwM2MServer to handle notification requests from device, the method signature looks like:

```go
LwM2MServer := NewLwM2MServer()
LwM2MServer.Handle("/", handler)
...
go LwM2MServer.ListenAndServe()

// Refer to LwM2M documentation for function implementations
func HandleNotify(request,response) {
    cache.Update(request.ObjectId, request.Value)
}

func HandleRegister(request,response) {
    cache.Add(request.ObjectId, request.Value)
}

func HandleDeRegister(request,response) {
    cache.Delete(request.ObjectId)
}
```

## Testing Plan

- Using [Leshan](https://github.com/eclipse-leshan/leshan) as a LwM2M client use deviceShifu-LwM2M to connect to the device without bootstrap server.
- Using [Leshan](https://github.com/eclipse-leshan/leshan)'s bootstrap server and client use deviceShifu-LwM2M to connect to the device.
- Normal mode test: read and write data to/from the device.
- Observe mode test: read and write data to/from the device, and check the data change or timeout.