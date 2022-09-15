<div align="right">

中文 | [English](README.md)

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

***Shifu*** 是一个 [Kubernetes](https://kubernetes.io/zh-cn/) 原生的IoT设备虚拟化框架。***Shifu*** 希望帮助IoT应用开发者以即插即用的方式实现IoT设备的虚拟化、监视、管控和自动化。

## 为什么用 ***Shifu***

***Shifu*** 通过数字孪生技术，让管理和控制IoT设备变得极其简单。当您连接设备的时候，***Shifu*** 会识别并以一个 [Kubernetes Pod](https://kubernetes.io/zh-cn/docs/concepts/workloads/pods/) 的方式启动一个该设备的数字孪生 ***deviceShifu***。

***deviceShifu*** 提供给用户了高层的交互抽象：开发者通过接入 ***deviceShifu*** 的接口，不仅可以实现IoT设备的所有设计功能，还可以实现原本设备所不具备的功能！例如：让设备主动将数据发送到某个地址或服务。

## 开始使用 ***Shifu***

### 安装

***Shifu*** 提供了`shifu_install.yml`文件。在已有Kubernetes集群的情况下，使用`kubectl apply`命令即可安装至集群：

```sh
cd shifu
kubectl apply -f pkg/k8s/crd/install/shifu_install.yml
```

### 演示

如果您不熟悉Kubernetes，我们准备了 [***Shifu*** Demo](https://shifu.run/zh-Hans/disclaimer/)。您可以直观的体验 ***Shifu*** 如何通过数字孪生来连接和管理实体设备。

### 使用文档

请在 <https://shifu.run/zh-Hans/docs/> 查看 ***Shifu*** 的使用文档。

## 深入理解 ***Shifu***

查看 [`docs/`](./docs/) 下的 Markdown文件 来了解 ***Shifu*** 的 [设计细节](./docs/design/) 和 [开发指南](./docs/guide/)。

## ***Shifu*** 里程碑

- 已支持功能
    - **Telemetry 收集**：***Shifu*** 可以定期收集设备的监测数据。监测数据的种类、收集的方式以及收集的频率都可以由用户在配置文件中自由设置。
    - **和 Kubernetes 通过 CRD 整合**：***Shifu*** 可以支持对任何设备进行任何形式的配置。
- 尚未支持功能
    - 自动生成 ***deviceShifu***
    - [声明式 API](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#declarative-apis)
    - [高级 ***shifuController***](docs/design/design-shifuController-zh.md)
    - [***shifud***](docs/design/design-shifud-zh.md)
    - 设备分组
    - 多层封装
    - 仿真

| 时间 | 协议 | 功能 |
| --- | --- | --- |
| Q1 2022 | HTTP <br> 命令行驱动 | 监测 <br> 命令代理 <br> CRD 整合 |
| Q2 2022 | MQTT <br> TCP Socket <br> RTSP <br> Siemens S7 <br> OPC UA | 状态机 |
| Q3 2022 | ONVIF | ***Shifu Cloud*** |
| Q4 2022 | gRPC | ***Shifu*** 抽象 <br> [***shifuController***](docs/design/design-shifuController-zh.md) <br> [***shifud***](docs/design/design-shifud-zh.md) <br> 仿真 |

如果您想要 ***Shifu*** 添加更多的功能和支持更多的协议，请 [新建 Issue](https://github.com/Edgenesis/shifu/issues)！

## ***Shifu*** 愿景

### 让开发者和运维人员再次开心

开发者和维护人员应100%聚焦在发明创造上，而不是修补基础设施以及重复造轮子。身为开发者和运维人员本身，***Shifu*** 的作者们深刻理解您的痛点！所以我们发自内心地想帮您解决掉底层的问题，让开发者和运维人员再次开心！

### 软件定义世界

如果每一个IoT设备都有一个 ***deviceShifu***，我们就可以借助软件来管理我们周围的世界。在一个软件定义的世界 (Software Define World) 中，所有东西都是智能的。您周围的一切会自动改变，进而更好地服务您；因为归根到底，科技以人为本。

## 社区

### 贡献

***Shifu*** 欢迎您 [新建 Issue](https://github.com/Edgenesis/shifu/issues/new) 或 [提交 PR](https://github.com/Edgenesis/shifu/pulls)。

### 联系我们

- 电子邮件
    - info@edgenesis.com 
- 微信
    - Donoteattoomuchla 
    - if7369

## GitHub Star 数量

[![Stargazers over time](https://starchart.cc/Edgenesis/shifu.svg)](https://starchart.cc/Edgenesis/shifu)
