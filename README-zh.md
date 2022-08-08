<div align="right">

中文 | [English](README.md)
</div>

# Shifu

[![Build Status](https://dev.azure.com/Edgenesis/shifu/_apis/build/status/shifu-build-muiltistage?branchName=main)](https://dev.azure.com/Edgenesis/shifu/_build/latest?definitionId=19&branchName=main)

Shifu是一个k8s原生的IoT设备虚拟化框架。 Shifu希望帮助IoT开发者以即插即用的方式实现IoT设备的虚拟化、监视、管控和自动化

## 为什么用 Shifu?

- Shifu通过数字孪生技术，让管理和控制IoT设备变得极其简单。当你连接设备的时候，Shifu会识别并以一个k8s Pod的方式启动一个该设备的数字孪生 ***deviceShifu***
- ***deviceShifu*** 提供给用户了高层的交互抽象
  - 开发者通过接入 ***deviceShifu*** 的接口，不仅可以实现IoT设备的所有设计功能，还可以实现原本设备所不具备的功能！例如：让设备主动将数据发送到某个地址或服务。

## 如何使用 Shifu?

请阅读[shifu文档](https://shifu.run/docs/)。

### 演示：
我们为开发者准备了一个 demo 来更直观地展示 `Shifu`是如何通过数字孪生来连接和管理实体设备的。
- [Shifu Demo](https://demo.shifu.run/)

# 我们的路线图
## 功能
### 已支持
- Telemetry 收集：shifu可以定期收集设备的监测数据。监测数据的种类、收集的方式以及收集的频率都可以由用户在配置文件中自由设置。
- 和 Kubernetes 通过 CRD 整合：shifu可以支持对任何设备进行任何形式的配置。
### 还未支持
- 自动生成***deviceShifu***
- [声明式 API](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#declarative-apis)
- [高级的 Shifu 控制器](docs/design/design-shifuController-zh.md)
- [shifud](docs/design/design-shifud-zh.md)
- 设备分组
- 多层封装
- 仿真

---

## 里程碑

如果您想要shifu支持更多的协议/功能，请在[这里](https://github.com/Edgenesis/shifu/issues)提交一个issue!

| 时间      | 协议                                     | 功能                                                 |
|---------|----------------------------------------------|---------------------------------------------------------|
| Q1 2022 | HTTP<br>命令行驱动 | 监测<br>命令代理<br>CRD 整合 |
| Q2 2022 | MQTT<br>TCP Socket<br>RTSP<br>Siemens S7<br>OPC UA | 状态机<br>shifu portal（前端） |
| Q3 2022 | ONVIF | shifu前端 |
| Q4 2022 | gRPC | shifu抽象<br>[shifuController](docs/design/design-shifuController-zh.md)<br>[shifud](docs/design/design-shifud-zh.md)<br>仿真 |

# Shifu的愿景

## 让开发者和运维人员再次开心

开发者和维护人员应将100%聚焦在发明创造上，而不是修补基础设施以及重复造轮子。身为开发者和运维人员本身，Shifu的作者们深刻理解你的痛点！所以我们发自内心地想帮你解决掉底层的问题，让开发者和运维人员再次开心！

## 软件定义世界 (SDW)

如果每一个IoT设备都有一个Shifu，我们就可以借助软件来管理我们周围的世界。在一个软件定义的世界中，所有东西都是智能的。你周围的一切会自动改变，进而更好的服务你。因为归根到底，科技以人为本

# 社区
## 联系
有问题？尝试[建立一个 GitHub Issue](https://github.com/Edgenesis/shifu/issues/new)，或者通过以下方式联系我们：
- 邮件: info@edgenesis.com 
- 微信:
  - Donoteattoomuchla 
  - if7369
