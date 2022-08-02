# WIP: ***DeviceShifu*** Refactor design

### What is DeviceShifu
citing From design-shifu.md: 
> ***deviceShifu*** is an structural [digital twin](https://en.wikipedia.org/wiki/Digital_twin) of the ***edgeDevice***. 
> We call it **structural** because it's not only a virtual presentation of the ***edgeDevice*** 
> but also it's capable of driving the corresponding ***edgeDevice*** towards its goal state. 
> For example, if you want your robot to move a box but it's currently busy on something else, 
> ***deviceShifu*** will cache your instructions and tell your robot to move the box whenever it's available. 
> For the basic part, ***deviceShifu*** provides some general functionalities such as ***edgeDevice*** health monitoring, state caching, etc. 
> By implementing the interface of ***deviceShifu***, your ***edgeDevice*** can achieve everything its designed for, and much more!


### Why do we need to refactor deviceShifu
Currently we implement a completely new deviceShifu struct for each new protocol, which doesn't utilize golang's feature 
and need to update every deviceShifu struct when we try to add new features. Thus, we need to decouple current deviceShifu
implementation to strip out the common part of each deviceShifu and create a base-deviceShifu struct, and whenever we 
need to create new deviceShifu, we only need to focus on the protocol specific part.
[Existing deviceShifu](https://github.com/Edgenesis/shifu/tree/main/deviceshifu/pkg)


### Design goals

1. Factor out the common parts of current existing ***shifuDevices*** and configs into a common ***base-shifuDevice*** struct to 
reduce duplicate codes and decouple some common device creation logic to the ***base-shifuDevice*** struct and config from each 
specific shifuDevice.
2. Create a ***deviceShifu*** interface as the guideline for adding new ***deviceShifu*** and let current main.go use the 
general ***deviceShifu*** interface to perform the actions needed to start a ***deviceShifu***.

### Non-Goal
Create a general universal ***deviceShifu*** to represent every edge device.


### Components

#### BaseDeviceShifuStruct

```go
type BaseDeviceShifu struct {
	Name              string
	server            *http.Server
	deviceShifuConfig *DeviceShifuConfig
	edgeDevice        *v1alpha1.EdgeDevice
	restClient        *rest.RESTClient
}

type BaseDeviceShifuMetaData struct {
    Name           string
    ConfigFilePath string
    KubeConfigPath string
    Namespace      string
}
```

We currently use ***BaseDeviceShifu*** and ***BaseDeviceShifuMetaData*** as examples of common part for each ***deviceShifu***.
For specific ***deviceShifu*** such as ***deviceShifuSocket*** we will use a composite struct like:
```go

type DeviceShifuSocket struct {
	baseDeviceShifu   *BaseDeviceShifu
	socketConnection  *net.Conn
}

```
In such a way, we can reuse ***baseDeviceShifu*** for all the common functionalities of that all edge devices share, and add fields
like ***socketConnection*** to handle protocol specific logics.

#### DeviceShifu interface:

```go
interface DeviceShifu {
	Start(stopCh <-chan struct{}) error
	Stop()                        error
}
```

The interface would only contain the basic methods that each edge device would share in common, which is just start and stop.

However, in order to create each different edge device base on the provided metadata, we need to use a factory pattern

```go
func createDeviceShifu(metaData BaseDeviceShifuMetaData) (DeviceShifu, error)
```
The factory method would take a metadata and create a concrete deviceShifu instance base on the metadata provided.