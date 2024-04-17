# DeviceShifu lwM2M Design

deviceShifu lwM2M allow shifu connect device using [lwM2M protocol].
This document outlines the design for integrating the [Lightweight Machine to Machine (lwM2M) protocol](https://omaspecworks.org/what-is-oma-specworks/iot/lightweight-m2m-lwm2m/) into the DeviceShifu framework to allow for the management of IoT devices.

## Goals

### Design Goal

- Create a deviceShifu-lwm2m type allow user to connect device using lwM2M protocol.
- lwM2M protocol support both `read` and `write` requests.
- Resource observation or Notification
- lwM2M protocol using v1.0.x version.
- lwM2M protocol under UDP.
- 
### Design Non-Goal

- Support lwM2M v1.1.x or later version.
- Datagram Transport Layer Security (DTLS) support.
- Over TCP or other protocol.
- Bootstrap Server.

## Detail Design

The `deviceShifu-lwM2M` will represent each lwM2M object as an instruction that can be accessed through a RESTful API. The API will use the POST method for write operations and the GET method for read operations. The supported data formats for lwM2M v1.0 are TLV, JSON, Plain text, and Opaque.

deviceShifu-lwM2M's lwM2M server will handle Register, Update, De-register, Read, Write, Observe and Notify.

deviceShifu-lwM2M support two kind of mode normal and observe mode.
- normal mode: just like the other deviceshifu, when call the instruction, deviceShifu will send Read or Write Request to deviceShifu and return response
- observe mode: this mode will set the min and max notify time, then device will notify data when data changed or timeout. deviceShifu will record data, and return the data store in the memory cache when call the instruction. and this kind of mode also support read and write operation.

deviceShifu will host a lwM2M server and listen on udp 3683(lwM2M default port) and http server(deviceshifu) on 8080.
the lwM2M server will handle register, update, de-register request from device and maintain the device info in the memory cache.
if the Object is in observe mode, lwM2M server also need to handle the notify request from device and update the data in the memory cache.
when the deviceShifu received the instruction before the device register, it will return error message. 

```mermaid
sequenceDiagram
    participant d as device
    participant ds as DeviceShifu
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
    d ->> ds: success
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

Define data structures and types in Go for configuring and managing lwM2M communication:

```go
type LwM2MSetting struct {
    EndpointName            string `json:"EndpointName,omitempty"`
    BootStrapServerAddress  string `json:"BootStrapServerAddress,omitempty"`
}
```

If bootstrap server address is not empty, deviceShifu will send register request to bootstrap server, and get the server info, then send register request to deviceShifu.

```go
type LwM2MType string

type Properties struct {
    ObjectId            string      `json:"ObjectId"` // required example /3303/0
    DataFormat          LwM2MType   `json:"dataFormat"` // optional TLV/JSON/PlainText/Opaque default plaintext
    EnableObserve       bool        `json:"EnableObserve,omitempty"` // optional enable observe mode default false
    ObserveMinPeriod    int         `json:"ObserveMinPeriod,omitempty"` // optional work when enable observe default 10 seconds
    ObserveMaxPeriod    int         `json:"ObserveMaxPeriod,omitempty"` // optional work when enable observe default 60 seconds
}
```

### Serving requests

deviceShifu-lwM2M would take RESTful-style requests just as other deviceShifu do.
lwM2M supports both `GET` and `PUT` requests.

For read the data from device, the method signature looks like:
```
GET lwm2m.device.svc.cluster.local/{object1_name}
```
```go
lwM2MServer.Read(properties.ObjectId)
if properties.EnableObserver {
    cache.Update(properties.ObjectId, newValue)
}
return value
```

For write the data to device, the method signature looks like:
```
PUT lwm2m.device.svc.cluster.local/{object1_name}
```
```go
lwM2MServer.Write(properties.ObjectId, newValue)
if properties.EnableObserver {
    cache.Update(properties.ObjectId, newValue)
}
return success
```

For the lwM2MServer to handle the notify request from device, the method signature looks like:

```go
lwM2MServer := NewLwM2MServer()
lwM2MServer.Handle("/", handler)
...
go lwM2MServer.ListenAndServe()

// Refer to lwM2M documentation for function implementations
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

- Using [Leshan](https://github.com/eclipse-leshan/leshan) as a lwM2M client use deviceshifu-lwM2M to connect to the device without bootstrap server.
- Using [Leshan](https://github.com/eclipse-leshan/leshan)'s bootstrap server and client use deviceshifu-lwM2M to connect to the device.
- Normal mode test: read and write data from device.
- Observe mode test: read and write data from device, and check the data change or timeout.