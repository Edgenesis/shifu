# ***shifud*** development design

## Introduction:
This is a development design document for ***shifud*** component of the ***Shifu*** system for edge devices.    

## Context:

### Goal:
The goal of ***shifud*** is to achieve the following:
- ***deviceDiscoverer***:
    1. Automatically detect USB devices local to the edgeNode.
    2. Automatically detect ONVIF devices in the network subnet.
    3. Given the connection type and address, detect if the device is live.
    4. Given the server url, discover devices registered in the OPC UA server.
- ***deviceVerifier***:
    1. Verify USB connected devices' metadata through udev.
    2. Verify network connected devices' metadata using SNMP/ONVIF/OPC UA.
- ***deviceShifuGenerator***:
    1. Generate the device resource to Kubernetes api-server.

### Input & Output:

#### ***deviceDiscoverer*** (per Node):
- Input: ConfigMap from Kubernetes cluster, mounted to a particular path.
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
- Output: List of discovered ONVIF cameras in JSON format(same as above).

#### ***deviceVerifier***
- Input: ConfigMap from Kubernetes cluster, deviceList from ***deviceDiscoverer***.
- Output: List of discovered devices that are verified matching ConfigMap.

#### ***deviceShifuGenerator***
- Input: List of devices from ***deviceVerifier***.

## Implementation

### Functions:

#### ***deviceDiscoverer***
1. A continuous loop that scans for ONVIF (Per cluster/subnet).    
    - *discoverONVIF()*
2. A continuous loop that scans udev.
    - *discoverUDEV()*
3. On demand loop that scans OPC UA servers.
    - *discoverOPCUA(str url)*
4. On demand TCP pinger for TCP/IP based alive detection.
    - *discoverTCP(str address, int port)*

#### ***deviceVerifier***
1. Listens for the device list sent from ***deviceDiscoverer***.
2. Upon receiving, query for metadata of the device.
    - *queryMetadata(str conn_type, str address, int port=None)*
        - *conn_type* can be: `SNMP`, `ONVIF`, `UDEV`, `OPCUA`
3. Compare the metadata of the device to ConfigMap.
    - *verifyDeviceMetadata(device configDevice, device discoveredDevice)*
4. If matches, output to ***deviceShifuGenerator***.


#### ***deviceShifuGenerator***
1. Listens for the device list sent from ***deviceVerifier***
2. Upon receiving, create Kubernetes resources.
    - *generateDeviceShifuResource(device discoveredDevice)*

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

#### ***deviceDiscoverer***
[![shifud deviceDiscoverer call stack](/img/shifud-deviceDiscoverer-call-stack.svg)](/img/shifud-deviceDiscoverer-call-stack.svg)    

#### ***deviceVerifier***
[![shifud deviceVerifier call stack](/img/shifud-deviceVerifier-call-stack.svg)](/img/shifud-deviceVerifier-call-stack.svg)    

#### ***deviceShifuGenerator***
[![shifud deviceShifuGenerator call stack](/img/shifud-deviceShifuGenerator-call-stack.svg)](/img/shifud-deviceShifuGenerator-call-stack.svg)    