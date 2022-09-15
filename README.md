<div align="right">

[中文](README-zh.md) | English

[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat&logo=github&color=2370ff&labelColor=454545)](http://makeapullrequest.com)
[![Go Report Card](https://goreportcard.com/badge/github.com/Edgenesis/shifu)](https://goreportcard.com/report/github.com/Edgenesis/shifu)
[![codecov](https://codecov.io/gh/Edgenesis/shifu/branch/main/graph/badge.svg?token=OX2UN22O3Z)](https://codecov.io/gh/Edgenesis/shifu)
[![Build Status](https://dev.azure.com/Edgenesis/shifu/_apis/build/status/shifu-build-muiltistage?branchName=main)](https://dev.azure.com/Edgenesis/shifu/_build/latest?definitionId=19&branchName=main)
[![golangci-lint](https://github.com/Edgenesis/shifu/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/Edgenesis/shifu/actions/workflows/golangci-lint.yml)

</div>

# ***Shifu***

<div align="center">

<img width="200px" src="./img/shifu-logo.svg"></img>

</div>

***Shifu*** is a [Kubernetes](https://kubernetes.io) native framework designed to abstract out the complexity of interacting with IoT devices. ***Shifu*** aims to achieve TRUE plug-and-play IoT device virtualization, management, control, and automation.

## Why ***Shifu***

***Shifu*** makes it extremely simple to manage and control your IoT devices via ***deviceShifu***, a digital twin for your device. When you connect an IoT device, ***Shifu*** will recognize it and start a ***deviceShifu*** of the device as a [Kubernetes Pod](https://kubernetes.io/docs/concepts/workloads/pods/).

***deviceShifu*** provides you with a high-level abstraction to interact with your device. By implementing the interface of the ***deviceShifu***, your IoT device can achieve everything it is designed for, and much more! For example, you can have your device actively push its telemetry to any endpoint of your choice.

## Start ***Shifu***

### Installation

***Shifu*** provides `shifu_install.yml`. If you have started a Kubernetes cluster, use the command `kubectl aply` to install ***Shifu*** in your cluster:

```sh
cd shifu
kubectl apply -f pkg/k8s/crd/install/shifu_install.yml
```

### Demo

If you are not familiar with Kubernetes, we provide [***Shifu*** Demo](https://shifu.run/disclaimer), which will intuitively show how ***Shifu*** creates and manages digital twins of any physical device in the real world.

### Documentation

See documentation on <https://shifu.run/docs/>.

## Look into ***Shifu***

Check the Markdown files in [`docs/`](./docs/) to learn about the [design](./docs/design/) and [development guides](./docs/guide/) of ***Shifu***.

# ***Shifu***'s Milestone

- features supported
    - **Telemetry collection**: ***Shifu*** supports the collection of telemetry from devices on a regular basis. What telemetry to collect, how to collect, how frequently to collect are all customizable in one single configuration file.
    - **Integration with Kubernetes with CRD**: ***Shifu*** allows any type or form of configuration for your devices.
- features not yet supported
    - Automatic ***deviceShifu*** creation
    - [Declarative API](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#declarative-apis)
    - [Advanced ***shifuController***](docs/design/design-shifuController.md)
    - [***shifud***](docs/design/design-shifud.md)
    - ***deviceShifu*** grouping
    - Abstraction
    - Simulation

| By | Protocols | Features |
| --- | --- | --- |
| Q1 2022 | HTTP <br> command line driver | telemetry <br> command proxy <br> CRD integration |
| Q2 2022 | MQTT <br> TCP Socket <br> RTSP <br> Siemens S7 <br> OPC UA | state machine |
| Q3 2022 | ONVIF | ***Shifu Cloud*** |
| Q4 2022 | gRPC | abstraction <br> [***shifuController***](docs/design/design-shifuController.md) <br> [***shifud***](docs/design/design-shifud.md) <br> simulation |

If you want more features and protocols supported, please [open an issue](https://github.com/Edgenesis/shifu/issues)!

# ***Shifu***'s vision

## Make developers and operators happy again

Developers and operators should focus entirely on using their creativity to make huge impacts, rather than fixing infrastructure and reinventing the wheel on a daily basis. As a developer and an operator, ***Shifu*** knows your pain!!! ***Shifu*** wants to take out all the underlying mess, and make us developers and operators happy again!

## Software-Defined World

If every IoT device has a ***deviceShifu*** with it, we can leverage software to manage our surroundings, and make the whole world software defined. In a software-defined world, everything is intelligent, and the world around you will automatically change itself to serve your needs -- after all, all the technology we built is designed to serve human beings.

# Community 

## Contribute

Welcome to [open an issue](https://github.com/Edgenesis/shifu/issues/new) or [create a PR](https://github.com/Edgenesis/shifu/pulls)!

## Contact Us

- Email
    - info@edgenesis.com
- WeChat
    - Donoteattoomuchla 
    - if7369

## Stargazers over time

[![Stargazers over time](https://starchart.cc/Edgenesis/shifu.svg)](https://starchart.cc/Edgenesis/shifu)
