<div align="right">

中文 | [English](README.md)
</div>

# Shifu

[![Build Status](https://dev.azure.com/Edgenesis/shifu/_apis/build/status/Edgenesis.shifu?branchName=main)](https://dev.azure.com/Edgenesis/shifu/_build/latest?definitionId=1&branchName=main)

Shifu是一个k8s原生的IoT设备虚拟化框架。 Shifu希望帮助IoT开发者以即插即用的方式实现IoT设备的虚拟化、监视、管控和自动化

## 为什么用 Shifu?

- Shifu让管理和控制IoT设备变得极其简单。当你连接设备的时候，Shifu会识别并以一个k8s Pod的方式启动一个该设备的虚拟设备 ***deviceShifu***
- ***deviceShifu*** 提供给用户了高层的交互抽象
  - 开发者通过接入 ***deviceShifu*** 的接口，不仅可以实现IoT设备的所有设计功能，还可以实现原本设备所不具备的功能！例如：在设备允许的状况下，通过一行命令来回滚设备的状态
  - Shifu 还可以实现设备分组以及多层封装，来自动执行更高级的命令。比如通过把各种机器的***deviceShifu***进一步封装到***factoryShifu***里面，再通过类似`Factory start`这样的命令，可以让一个工业4.0的工厂自动开始生产。
- 之后实现的 simulation 功能可以使开发人员在执行命令前演算一遍，通过模拟现实来避免实际执行时可能遇到的问题

## Shifu 与 Thing
Shifu框架利用Web of Things (WoT)提出的[Thing](https://www.w3.org/TR/wot-thing-description/)的理念对连接的设备进行定义。
- Shifu框架内，用户添加关于设备的描述即可完成连接。在连接完成后，Shifu框架会开始管控该设备。
- Shifu所需要的描述包括：
  - 设备的连接方式和驱动，即设备在物联网中的属性
  - 设备支持的操作，即设备可执行的服务
  - 设备的监控数据种类以及监控方式，即设备汇报的事件

## 如何使用 Shifu?

当前，Shifu运行在[Kubernetes](https://k8s.io) 上。我们将来会提供包含单独部署在内的更多部署方式

### 在现有 Kubernetes 集群中[安装](docs/guide/install-zh.md) Shifu

---

### 演示：
我们为开发者准备了一个 demo 来更直观地展示 `Shifu`是如何建立管理IoT设备的
- [带演示的快速上手指南](docs/guide/quick-start-demo-zh.md)
- [在线演示(Katacoda)](https://www.katacoda.com/xqin/scenarios/shifu-demo)

# 我们的路线图
## 协议
### 已支持
- HTTP
- 通过命令行调用的驱动程序
- ... 更多正在开发中
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
- 演算

---

## 里程碑

如果您想要shifu支持更多的协议/功能，请在[这里](https://github.com/Edgenesis/shifu/issues)提交一个issue!

| 时间      | 协议                                     | 功能                                                 |
|---------|----------------------------------------------|---------------------------------------------------------|
| Q1 2022 | HTTP<br>命令行驱动 | 监测<br>命令代理<br>CRD 整合 |
| Q2 2022 | MQTT<br>TCP Socket<br>RTSP<br>Siemens S7<br>OPC UA | 状态机<br>shifu portal（前端） |
| Q3 2022 | ONVIF<br>OPC UA | shifu抽象<br>[shifuController](docs/design/design-shifuController-zh.md)<br>[shifud](docs/design/design-shifud-zh.md) |
| Q4 2022 | gRPC | 仿真 |

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
  - zhengkaiwen196649 
  - if7369
