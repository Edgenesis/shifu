- [***shifud*** development design](#shifud-development-design)
  - [Introduction:](#introduction)
  - [Context:](#context)
    - [Goal:](#goal)
    - [Input & Output:](#input--output)
      - [***deviceDiscoverer*** (per Node):](#devicediscoverer-per-node)
      - [***deviceDiscoverer*** (per Cluster/Subnet):](#devicediscoverer-per-clustersubnet)
      - [***deviceVerifier***](#deviceverifier)
      - [***deviceUpdater***](#deviceupdater)
  - [Implementation](#implementation)
    - [Functions:](#functions)
      - [***deviceDiscoverer***](#devicediscoverer)
      - [***deviceVerifier***](#deviceverifier-1)
      - [***deviceUpdater***](#deviceupdater-1)
    - [Data Types:](#data-types)
      - [***device***](#device)
    - [Call stack](#call-stack)
      - [***overall***](#overall)
      - [***deviceDiscoverer*** (per cluster)](#devicediscoverer-per-cluster)
      - [***deviceDiscoverer*** (per ***edgeNode***)](#devicediscoverer-per-edgenode)
      - [***deviceVerifier*** (per cluster)](#deviceverifier-per-cluster)
      - [***deviceVerifier*** (per ***edgeNode***)](#deviceverifier-per-edgenode)
      - [***deviceUpdater***](#deviceupdater-2)

# ***shifud*** development design

## Introduction:
This is a development design document for ***shifud*** component of the ***Shifu*** system for edge devices.    

## Context:

### Goal:
The goal of ***shifud*** is to achieve the following:
- ***deviceDiscoverer***:
    1. Automatically detect ONVIF devices in the network subnet.
    2. Given the connection type and address, detect if the device is live.
    3. Given the server url, discover devices registered in the OPC UA server.
- ***deviceVerifier***:
    1. Verify USB connected devices' metadata through udev.
    2. Verify network connected devices' metadata using SNMP/ONVIF/OPC UA.
- ***deviceUpdater***:
    1. Update the status field of ***edgeDevice*** resource to Kubernetes API server.

### Input & Output:

#### ***deviceDiscoverer*** (per Node):
- Input: ***edgeDevice*** resource from Kubernetes' ***edgeDevice*** resource.
- Output: List of discovered devices in the following JSON format:    
    ```
    [
        {
            "deviceName": "",
            "connection": "",
            "address": "",
            "type": "",
            "vendor": "",
            "protocol": ""
        },
        ...
    ]
    ```

#### ***deviceDiscoverer*** (per Cluster/Subnet):
- Output: List of discovered ONVIF cameras in JSON format (same as above).

#### ***deviceVerifier***
- Input: ***edgeDevice*** resource from Kubernetes cluster, deviceList from ***deviceDiscoverer***.
- Output: List of discovered devices that are verified matching ***edgeDevice*** resource.

#### ***deviceUpdater***
- Input: List of devices from ***deviceVerifier***.

## Implementation

### Functions:

#### ***deviceDiscoverer***
1. A continuous loop that scans for ONVIF (per cluster/subnet).    
    - *discoverONVIF()*
2. On demand loop that scans udev event (per ***edgeNode***).
    - *discoverUDEV()*
3. On demand loop that scans OPC UA servers (per cluster/subnet).
    - *discoverOPCUA(str url)*
4. On demand TCP pinger for TCP/IP based alive detection (per cluster/subnet).
    - *discoverTCP(str address, int port)*

#### ***deviceVerifier***
1. Listens for the device list sent from ***deviceDiscoverer***.
2. Upon receiving, query for metadata of the device.
    - *queryMetadata(device discoveredDevice)*
3. Compare the metadata of the device to ***edgeDevice*** resource.
    - *verifyDeviceMetadata(device edgeDevice, device discoveredDevice)*
4. If matches, output to ***deviceUpdater***.


#### ***deviceUpdater***
1. Listens for the device list sent from ***deviceVerifier***
2. Upon receiving, update Kubernetes resources.
    - *updateEdgeDeviceResource(device discoveredDevice)*

### Data Types:

#### ***device***
```
{
    str deviceName,
    str connection_type,
    str address,
    deviceType deviceType,
    str deviceVendor,
    protocol deviceProtocol
}
```

### Call stack
#### ***overall***
[![shifud overall call stack](/img/shifud-overall-call-stack.svg)](/img/shifud-overall-call-stack.svg)    

#### ***deviceDiscoverer*** (per cluster)
[![shifud deviceDiscoverer call stack](/img/shifud-deviceDiscoverer-cluster-call-stack.svg)](/img/shifud-deviceDiscoverer-cluster-call-stack.svg)    

#### ***deviceDiscoverer*** (per ***edgeNode***)
[![shifud deviceDiscoverer call stack](/img/shifud-deviceDiscoverer-edgeNode-call-stack.svg)](/img/shifud-deviceDiscoverer-edgeNode-call-stack.svg)    

#### ***deviceVerifier*** (per cluster)
[![shifud deviceVerifier call stack](/img/shifud-deviceVerifier-cluster-call-stack.svg)](/img/shifud-deviceVerifier-cluster-call-stack.svg)    

#### ***deviceVerifier*** (per ***edgeNode***)
[![shifud deviceVerifier call stack](/img/shifud-deviceVerifier-edgeNode-call-stack.svg)](/img/shifud-deviceVerifier-edgeNode-call-stack.svg)    

#### ***deviceUpdater***
[![shifud deviceUpdater call stack](/img/shifud-deviceUpdater-call-stack.svg)](/img/shifud-deviceUpdater-call-stack.svg)    