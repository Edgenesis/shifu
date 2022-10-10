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

***Shifu*** is a cloud native framework to help IoT developers interact with physical devices. ***Shifu*** virtualizes the management and monitoring of devices, giving them the ability to run automatically. It solves the troubles of protocol and driver integration and enables devices to be plug-and-play.

## Why ***Shifu***

- ***Shifu*** is a cloud native framework that achieves aerospace level stability (up to 99.9999%).
- Device protocols and drivers are compiled from configuration files, which are efficient and reusable.
- Modular deployment of devices, developers can simply load them on demand.

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

## Stargazers over time

[![Stargazers over time](https://starchart.cc/Edgenesis/shifu.svg)](https://starchart.cc/Edgenesis/shifu)
