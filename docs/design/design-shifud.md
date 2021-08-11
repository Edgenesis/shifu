# ***shifud*** High-level design

## Introduction:
This is a high-level design document for ***shifud*** component of the ***Shifu*** system for edge devices. ***shifud*** is a daemonset that runs on every ***edgeNode***. It discovers devices from a list of devices sent by ***shifuController*** and spawn relative resouces to kube-apiserver.    

## Design principles
[TODO]

## Design goals and non-goals
[TODO]

## Design overview
  

### Components

#### Software components

##### 1. ***deviceDiscoverer***
***deviceDiscoverer*** is a process that continuously scans for device events on ***edgeNode***, including but not limited to network reachability, USB event.

##### 2. ***deviceVerifier***
***deviceVerifier*** is a process that interacts with ***edgeDevice*** and try to populate and verify their metadata to match ***shifuController***'s list

##### 3. ***deviceShifuGenerator***
***deviceShifuGenerator*** generates and creates a ***deviceShifu*** resource to ***kube-apiserver*** based on the ***edgeDevice***

### ***shifud*** input & output
The overall input and output of ***shifud*** can be summarized in the following graph:
[![shifud input and output overview](/img/shifud-input-output.PNG)](/img/shifud-input-output.PNG)    
The input to ***shifud*** from shifuController should be a list of edge devices in the following format:    
```
#deviceName, connection, address, type, brand, protocol
deviceA, USB, /tty/USB1, t_sensor, Xiaomi, MQTT
deviceB, IP, 10.0.0.1, IP_camera, Yunmi, ONVIF
......
```

### Architecture diagrams
[![shifud design overview](/img/shifud-design-overview.PNG)](/img/shifud-design-overview.PNG)    

#### ***shifud***'s execution flow:
1. Upon receiving the list of devices, ***deviceDiscoverer*** starts local scanning using different protocols. The following protocols should be supported:
   ```
   udev
   ONVIF
   SNMP
   MQTT
   MODBUS
   OPC UA
   ```
2. The discovery process depends on the protocol:
    1. For TCP/IP type of edge devices, Ping/TCP connect can be use directly
    2. For udev/USB type of edge devices, ***deviceDiscoverer*** will utilize Linux's [udev](https://man7.org/linux/man-pages/man7/udev.7.html) tool    
3. Once the device has been discovered, ***deviceVerifier*** will then start matching the device metatdata with device list through its connection protocol.
    ```
    sample udevadm output:
    E: DEVNAME=/dev/video3
    E: SUBSYSTEM=video4linux
    E: ID_SERIAL=Sonix_Technology_Co.__Ltd._USB_2.0_Camera_SN0001
    ```
4. Once the verification is done, ***deviceShifuGenerator*** will send out the deviceShifu's deployment YAML file to the controller for spawning the actual deviceShifu:
   ```
   apiVersion: v1
   kind: deviceShifu
   metadata:
       name: shifu-deviceA
       labels:
           connection: USB
           location: /tty/USB1
           protocol: MQTT
   spec:
       containers:
           -  name: shifu-deviceA
              image: shifu-t-sensor
    ......
    ```
