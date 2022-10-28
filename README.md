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

***Shifu*** is the next-generation open source IoT development framework, makes developing an IoT app as simple as developing a Web app. The high extensibility of ***Shifu*** allows it to integrate almost all devices and protocols. Once a device is integrated, ***Shifu*** generates a digital twin of the device as a microservice in the system and opens up device data and capabilities in the form of APIs. In this way, traditional IoT application development is simplified to web development, which greatly improves the efficiency, quality and reusability of IoT application development.

|Features<div style="width: 100pt">|  |
|---|---|
|💻 Blazing-fast|From thermohydrometers using standard protocols to complex machinery using proprietary drivers, Shifu is capable of integrating all kinds of heterogeneous devices.|
|▶️ Modularized|All devices and Apps integrated into Shifu are packaged as modules that can be loaded on demand.|
|👨‍💻 Efficient|Once a device is integrated, Shifu automatically abstracts its capabilities into APIs, completely decoupling your App from the hardware, making IoT App development simple and efficient.|
|🚀 Stable|Shifu is running in multiple production scenarios with 99.9999% stability, relieving you from the operational mess.|
|🛡️ Safe |Designed by ex-UN cloud native security team. Shifu can easily enforce data encryption, network security and much more.|
|🌐 Global Community|Benefiting from its Kubernetes' native architecture, Shifu can seamless integrate with the powerful cloud native software eco-system that allows developers around the world to help you with your problems.|

## Guide

### Install

If you have started a Kubernetes cluster, use the command `kubectl apply` to install ***Shifu*** in your cluster:

```sh
cd shifu
kubectl apply -f pkg/k8s/crd/install/shifu_install.yml
```

### Try it out

We present you [***Shifu*** Demo](https://shifu.run/disclaimer). 

The Demo will show you how ***Shifu*** creates and manages digital twins of any physical devices.

### Documentation

For more information please visit our [documentation](https://shifu.run/docs/).

## Design principle

Our job is to make developers and operators happy. Which is why all our designs are developer-oriented:
#### 📡Easy to deploy
Shifu can be deployed with a single command.
#### 🤖Plug'n'Play
Shifu will automatically recognize and provide basic functionalities to a supported IoT device. Once the developer completes Shifu's template, all features of the device should be immediately available.
#### 🪄High extensibility
Developer can further implement Shifu's interface/SDKs to create custom features and unleash endless possibilities.
#### 🔧Zero maintenance
Shifu aims to achieve zero maintenance by adopting cutting-edge cloud native best practices. After all, Shifu should take care of himself and make DevOps' lives easier!

# Community 

## Contribute

Welcome to [open an issue](https://github.com/Edgenesis/shifu/issues/new) or [create a PR](https://github.com/Edgenesis/shifu/pulls)!

## Contact Us

- Email
  - info@edgenesis.com

## Stargazers over time

[![Stargazers over time](https://starchart.cc/Edgenesis/shifu.svg)](https://starchart.cc/Edgenesis/shifu)
