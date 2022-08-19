<div align="right">

[中文](README-zh.md) | English
</div>

# Shifu

[![Build Status](https://dev.azure.com/Edgenesis/shifu/_apis/build/status/shifu-build-muiltistage?branchName=main)](https://dev.azure.com/Edgenesis/shifu/_build/latest?definitionId=19&branchName=main)

![](./img/shifu-logo.svg)

Shifu is a [Kubernetes](https://k8s.io) native framework designed to abstract out the complexity of interacting with IoT devices. Shifu aims to achieve TRUE plug'n'play IoT device virtualization, management, control and automation.

## Why use Shifu?

- Shifu let you manage and control your IoT device extremely easily through ***deviceShifu***, a digital twin for your device.
- ***deviceShifu*** provides you with a high-level abstraction to interact with your device. By implementing the interface of the ***deviceShifu***, your IoT device can achieve everything its designed for, and much more! For example, you can have your device actively push its telemetries to any endpoint of your choice.

## How to use Shifu?
Please refer to [shifu doc](https://shifu.run/docs/).

### Demo:
We have prepared a demo for developers to intuitively show how `Shifu` is able to create and manage digital twins of any physical devices in real world.
- [Shifu Demo](https://demo.shifu.run/)

# Our Roadmap
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
If you want more features/protocols supported, please open an issue [here](https://github.com/Edgenesis/shifu/issues)!

| By      | Protocol                                     | Features                                                 |
|---------|----------------------------------------------|---------------------------------------------------------|
| Q1 2022 | HTTP<br>Driver w/ command line | telemetry<br>command proxy<br>CRD integration |
| Q2 2022 | MQTT<br>TCP Socket<br>RTSP<br>Siemens S7<br>OPC UA | state machine |
| Q3 2022 | ONVIF | shifu portal(frontend) |
| Q4 2022 | gRPC | abstraction<br>[shifuController](docs/design/design-shifuController.md)<br>[shifud](docs/design/design-shifud.md)<br>simulation |

# Shifu's vision

## Make developers and operators happy again

Developers and operators should 100% focus on using their creativity to make huge impacts, not fixing infrastructures and re-inventing the wheels day in and day out. As a developer and an operator, Shifu knows your pain!!! Shifu wants to take out all the underlying mess, and make us developers and operators happy again!

## Software Defined World (SDW)

If every IoT device has a Shifu with it, we can leverage software to manage our surroundings, and make the whole world software defined. In a SDW, everything is intelligent, and the world around you will automatically change itself to serve your needs. After all, all the technology we built are designed to serve human beings. 

# Community 

## Contact

Feel free to open a [GitHub issue](https://github.com/Edgenesis/shifu/issues/new) or contact us in the following ways:
- Email: info@edgenesis.com
- Wechat:
  - Donoteattoomuchla 
  - if7369
