# DeviceShifu PLC4X Design

DeviceShifu-PLC4X allows shifu utilize [apache plc4x](https://plc4x.apache.org/) project to integrate PLC devices with various protocols such as S7, Modbus, etc.

## Goal
- Create a deviceShifu-PLC4X type using PLC4X library and allow user to connect PLC devices with all the [supported protocols](https://plc4x.apache.org/users/protocols/index.html) with only this single deviceShifu-PLC4X.

## General Design

deviceShifu-PLC4X will receive RESTful style request like other, then transfer the request into valid PLC4X method calls and send it to connected devices.

## Detailed Design

### Protocol Specification
PLC4X can support various protocols, thus when initializing deviceShifu-PLC4X, user should specify which protocol to use. In order to support protocol specification,
we can add new settings to edgedevice_types.go to support it.

```go
type ProtocolSettings struct {
	MQTTSetting   *MQTTSetting   `json:"MQTTSetting,omitempty"`
	OPCUASetting  *OPCUASetting  `json:"OPCUASetting,omitempty"`
	SocketSetting *SocketSetting `json:"SocketSetting,omitempty"`
+	PLC4XSetting  *PLC4XSetting  `json:"PLC4XSetting,omitempty"`
}

// PLC4XSetting defines PLC4X specific settings when connecting to an EdgeDevice
type PLC4XSetting struct {
    //DrvierType is an enum of all the supported protocols
    DriverType *DriverType `json:"driverType,omitempty"`
}
```
DeviceShifu-PLC4X will maintain an enum of protocols and match the driver base on the setting:
```go
 swtich plc4xSetting.DrvierType:
	 case S7:
		 drviers.RegsiterS7(drvierManager)
		 ...
```

### Serving requests
deviceShifu-PLC4X would take RESTful-style requests just as other deviceShifu do. 
PLC4X supports both `read` and `write` requests. 

For read, the method signature looks like:
```go
// Prepare a read-request
readRequest, err := connection.ReadRequestBuilder().
	AddQuery("field", "holding-register:26:REAL").
	AddQuery("field", "holding-register:3:REAL").
	Build()
```
For write, the method signature looks like:
```go
// Prepare a write-request
writeRequest, err := connection.WriteRequestBuilder().
	AddQuery("field", "holding-register:26:REAL", 2.7182818284).
	AddQuery("field", "holding-register:3:REAL", 3.141592657).
	Build()
```

Thus, we can construct the REST request as the following format:
```
For read:
http://device-plc/read?${field1}?${field2}
e.g: http://device-plc/read?holding-register:26:REAL?holding-register:3:REAL
For write:
http://device-plc/write?${field1}=${value1}?${field2}=${value2}
e.g: http://device-plc/write?holding-register:26:REAL=2.7182818284?holding-register:3:REAL=3.141592657
```
The `read` , `write` after `http://device-plc/` will be served as an instruction router.
`?` would be served as query delimiter, whatever comes between `?` will be seen as the query.

For `read`, the query would only contain the fields requests reading from. 
For `write`, the query would contain the fields requests trying to write and the value it tries to write. 

## Testing Plan
We can use existing mock-plc device, create a deviceShifu-PLC4X image add a e2e test to current azure-pipeline to run test against deviceShifu-PLC4X.

