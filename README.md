<div align="right">

[中文](README-zh.md) | English
</div>

# Shifu

[![Build Status](https://dev.azure.com/Edgenesis/shifu/_apis/build/status/Edgenesis.shifu?branchName=main)](https://dev.azure.com/Edgenesis/shifu/_build/latest?definitionId=1&branchName=main)

Shifu is a [Kubernetes](https://k8s.io) native framework designed to abstract out the complexity of interacting with IoT devices. Shifu aims to achieve TRUE plug'n'play IoT device virtualization, management, control and automation.

## Why use Shifu?

- Shifu let you manage and control your IoT devices extremely easily. Whenever you connect your device, Shifu will recognize it and spawn an augmented digital twin called ***deviceShifu*** for it. 
- ***deviceShifu*** provides you with a high-level abstraction to interact with your device. 
  - By implementing the interface of the ***deviceShifu***, your IoT device can achieve everything its designed for, and much more! For example, your device's state can be rolled back with a single line of command. (If physically permitted, of course.) 
  - Shifu is able to abstract ***deviceShifu*** horizontally (grouping, batch execute) and vertically (layers, allow high level command to be executed. e.g.: `factory start`). For example, we can group ***deviceShifu*** of machines into a ***factoryShifu***, and then high-level commands like `factory start` will make the whole factory start manufacturing.
- A simulation feature which allows developer to simulate a scenario before actually running will be available later.

## Shifu and Thing

Shifu utilized the Web of Things (WoT)' conception of [Thing](https://www.w3.org/TR/wot-thing-description/) to describe a device to be connected. 
- Within the Shifu framework, user can connect a device to the framework by simply creating a configuration about the device. After the connection is established, Shifu will automatically start managing the device. 
- Shifu will need 3 types of descriptions: 
  - the connection type and driver, which is the "property" of the device; 
  - the available instructions, which is the "actions" or "services" of the device;
  - the telemetries we expect to get from the device for monitoring, which is the "events" of the device.

## How to use Shifu?

Currently, Shifu runs on [Kubernetes](https://k8s.io). We will provide more deployment methods including standalone deployment in the future.

### [Install](docs/guide/install.md) Shifu on Kubernetes cluster

---

### Demo:
We prepared a demo for developers to intuitively show how `Shifu` is able to create and manage digital twins of any physical devices in real world.
- [Quick Start Guide with Demo](docs/guide/quick-start-demo.md)
- [Online interactive Demo (Katacoda)](https://www.katacoda.com/xqin/scenarios/shifu-demo)

# Our Roadmap
## Protocols
### Supported
- HTTP
- Driver implementation w/ command line execution
- ... More on the way
## Features
### Supported
- Telemetry collection: shifu supports periodic collection of any telemetries from device. What telemetries to collect, how to collect, how frequent is the collection are all customizable in one single configuration file.
- Integration with Kubernetes with CRD: shifu allows any types or forms of configurations for your devices.
### Not yet supported
- Auto ***deviceShifu*** generation
- [Declarative API](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#declarative-apis)
- [Advanced Shifu Controller](docs/design/design-shifuController.md)
- [shifud](docs/design/design-shifud.md)
- Abstraction:
  - Horizontal
  - Vertical
- Simulation

---

## Milestone

| By      | Protocol                                     | Features                                                 |
|---------|----------------------------------------------|----------------------------------------------------------|
| Q4 2021 | HTTP<br>Driver w/ command line                  | Telemetry<br>Command proxy<br>CRD integration<br>Basic Controller |
| Q1 2022 | At least:<br>MQTT<br>Modbus<br>ONVIF<br>国标GB28181<br>USB  | Declarative API<br>Advanced Controller<br>shifud               |
| Q2 2022 | At least:<br>OPC UA<br>Serial<br>Zigbee<br>LoRa<br>PROFINET | Abstraction                                              |
| Q3 2022 | TBD                                          | Security Features                                        |
| Q3 2023 | TBD                                          | Simulation                                               |

# Shifu's vision

## Make developers and operators happy again

Developers and operators should 100% focus on using their creativity to make huge impacts, not fixing infrastructures and re-inventing the wheels day in and day out. As a developer and an operator, Shifu knows your pain!!! Shifu wants to take out all the underlying mess, and make us developers and operators happy again!

## Software Defined World (SDW)

If every IoT device has a Shifu with it, we can leverage software to manage our surroundings, and make the whole world software defined. In a SDW, everything is intelligent, and the world around you will automatically change itself to serve your needs. After all, all the technology we built are designed to serve human beings. 

# Community 

## Contact

<<<<<<< HEAD
Feel free to open a [GitHub issue](https://github.com/Edgenesis/shifu/issues/new) or contact us in the following ways:
=======
Feel free to [open a GitHub issue ](https://github.com/Edgenesis/shifu/issues/new)or contact us in the following ways:
>>>>>>> 22aa156 (Modify some old documents)
- Email: info@edgenesis.com
- Wechat:
  - zhengkaiwen196649 
  - if7369
