- [***shifud*** High-level design](#shifud-high-level-design)
  - [Introduction:](#introduction)
  - [Design principles](#design-principles)
    - [Automatic & Autonomous](#automatic--autonomous)
      - [1. Automatic discovery for discoverable ***edgeDevices***:](#1-automatic-discovery-for-discoverable-edgedevices)
      - [2. Minimal information needed for non-automatic ***edgeDevice*** device discovery:](#2-minimal-information-needed-for-non-automatic-edgedevice-device-discovery)
  - [Design goals and non-goals](#design-goals-and-non-goals)
    - [Design goals](#design-goals)
      - [Autonomous](#autonomous)
      - [Lightweight](#lightweight)
      - [Flexible](#flexible)
    - [Design non-goals](#design-non-goals)
  - [Design overview](#design-overview)
    - [Components](#components)
      - [Software components](#software-components)
        - [1. ***deviceDiscoverer***](#1-devicediscoverer)
        - [2. ***deviceVerifier***](#2-deviceverifier)
        - [3. ***deviceUpdater***](#3-deviceupdater)
    - [***shifud*** input & output](#shifud-input--output)
      - [Architecture diagrams](#architecture-diagrams)
      - [***shifud***'s execution flow(cluter):](#shifuds-execution-flowcluter)
      - [***shifud***'s execution flow(edgeNode):](#shifuds-execution-flowedgenode)

# ***shifud*** High-level design

## Introduction:
This is a high-level design document for ***shifud*** component of the ***Shifu*** framework. ***shifud*** is a daemonset that runs on every ***edgeNode***. It discovers devices from a list of devices configured in Kubernetes' ***edgeDevice*** resource and update relative resources to kube-apiserver.    

## Design principles
### Automatic & Autonomous
***shifud***'s foremost job is to make ***edgeDevice*** discovery and verification as easy as possible. Developers should't need much configuration to make their ***edgeDevice*** available in ***Shifu***. Here are a few design requirements:

#### 1. Automatic discovery for discoverable ***edgeDevices***:
***shifud*** should be able to discover ***edgeDevices*** with ONVIF or similar protocols automatically, without much user/developer intervention.

#### 2. Minimal information needed for non-automatic ***edgeDevice*** device discovery:
Developer should only need to provide necessary information in order for ***shifud*** to discover a specific devices.

## Design goals and non-goals
### Design goals
#### Autonomous
***shifud*** should be able to run on its own as soon as ***Shifu*** framework is up.

#### Lightweight
***shifud*** should consume minimum amount of resource on each ***edgeNode*** and across the cluster.

#### Flexible
***shifud*** should be able to handle the majority of IoT protocols.

### Design non-goals
[None]

## Design overview
  

### Components

#### Software components

##### 1. ***deviceDiscoverer***
***deviceDiscoverer*** is a process that continuously scans for device events on ***edgeNode***, including but not limited to network reachability, USB event.

##### 2. ***deviceVerifier***
***deviceVerifier*** is a process that interacts with ***edgeDevice*** and tries to populate and verify their metadata to match Kubernetes' ***edgeDevice*** resources.

##### 3. ***deviceUpdater***
***deviceUpdater*** updates ***edgeDevice*** resource to ***kube-apiserver*** based on the ***edgeDevice***'s verification status.

### ***shifud*** input & output
The overall input and output of ***shifud*** can be summarized in the following graph:
[![shifud input and output overview](/img/shifud-input-output.svg)](/img/shifud-input-output.svg)    
The input to ***shifud*** from Kubernetes ***edgeDevice*** resource should be a list of ***edgeDevices***:    
```
apiVersion: v1
kind: edgeDevice
metadata:
  name: franka-emika-1
spec:
- sku: "Franka Emika"
  connection: Ethernet
  status: offline
  address: 10.0.0.1:80
  protocol: HTTP
  disconnectTimeoutInSeconds:600 # optional
  group:["room1", "robot"] # optional
  driverSpec: # optional when no driver is required
  - instructionMap:  # optional
      move_to:
      - api: absolute_move # API of the driver
......
```

#### Architecture diagrams
[![shifud design overview](/img/shifud-design-overview.svg)](/img/shifud-design-overview.svg)    


#### ***shifud***'s execution flow(cluter):
1. Upon querying the list of devices, ***deviceDiscoverer*** starts local scanning using ethernet protocols. The following protocols should be supported:
   ```
   ONVIF
   SNMP
   MQTT
   OPC UA
   PROFINET
   ```

#### ***shifud***'s execution flow(edgeNode):
1. Upon receiving the list of devices, ***deviceDiscoverer*** starts local scanning using different protocols. The following protocols should be supported:
   ```
   udev
   MODBUS
   ```
2. The discovery process depends on the protocol:
    1. For TCP/IP type of edge devices, ping/TCP connect can be use directly.
    2. For udev/USB type of edge devices, ***deviceDiscoverer*** will utilize Linux's [udev](https://man7.org/linux/man-pages/man7/udev.7.html) tool.    
3. Once the device has been discovered, ***deviceVerifier*** will then start matching the device metatdata with device list through its connection protocol.
    ```
    sample udevadm output:
    E: DEVNAME=/dev/video3
    E: SUBSYSTEM=video4linux
    E: ID_SERIAL=Sonix_Technology_Co.__Ltd._USB_2.0_Camera_SN0001
    ```
4. Once the verification is done, ***deviceUpdater*** will update the ***edgeDevice***'s resource to Kubernetes API server.
    ```
   apiVersion: v1
   kind: edgeDevice
   metadata:
     name: franka-emika-1
   spec:
   - sku: "Franka Emika"
     connection: Ethernet
     status: online
     address: 10.0.0.1:80
     protocol: HTTP
     disconnectTimeoutInSeconds:600 # optional
     group:["room1", "robot"] # optional
     driverSpec: # optional when no driver is required
     - instructionMap:  # optional
         move_to:
         - api: absolute_move # API of the driver
   ......
    ```
