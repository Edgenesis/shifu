# Welcome to Shifu
Shifu is an open-source, [Kubernetes](https://kubernetes.io/)-based Internet of Things (IoT) development and management platform to connect, monitor and control all the IoT devices in a place. Shifu allows users to create software applications to interact with different types of IoT devices in an easier way.

## IoT devices
An IoT device is a device that can connect and exchange data with other devices and systems over the Internet or other communications networks. For example: 
- A robot arm in a manufacturing plant, receiving commands from the automation software in local server.
- An automated guided vehicle (AGV) in a lab, controlled by a remote operator.
- A temperature sensor in a car, instructing the air conditioner to turn up or down, as well as sending the live temperature data to cloud server for monitoring.

There is a wide variety of devices available from different manufacturers that can be used in your project. Shifu has integrated many different protocols and drivers, and has the feature to allow users to integrate new drivers, to ensure a high compatibility.

## Communication
Shifu accepts different types of communication from devices, and converts them into a simpler one such as HTTP, to help user interact with the devices more easily.

**Between Shifu and IoT devices:**
Shifu is constantly integrating new protocols and drivers. For a list of supported device protocols and drivers, see the page of [Supported Device Protocols and Drivers](./supported_device_protocols_and_drivers.md).

**Between Shifu and user:**
Shifu is constantly adding new protocols for user to choose from. For a list of supported user-facing protocols, see the page of [Supported User Protocols and Drivers](./supported_user_protocols_and_drivers.md).  

## Functionality
As an IoT development and management platform, Shifu provides functionality such as: 
- Gathering telemetries from devices.
- Sending requests to devices.
