- [Shifu 高层设计](#shifu-高层设计)
  - [设计理念](#设计理念)
    - [以人为本](#以人为本)
      - [1. 容易部署](#1-容易部署)
      - [2. 即插即用](#2-即插即用)
      - [3. 零维护](#3-零维护)
      - [4. 易用的 SDK](#4-易用的-sdk)
  - [设计目标与非目标](#设计目标与非目标)
    - [设计目标](#设计目标)
      - [简单](#简单)
      - [高可用](#高可用)
        - [1. 自我修复](#1-自我修复)
        - [2. 稳定](#2-稳定)
      - [跨平台](#跨平台)
      - [轻量](#轻量)
      - [可扩展性和可变性](#可扩展性和可变性)
    - [设计非目标](#设计非目标)
      - [除 Kubernetes 的其他平台](#除-kubernetes-的其他平台)
      - [100% up-time](#100-up-time)
  - [设计总览](#设计总览)
    - [组件](#组件)
      - [物理组件](#物理组件)
        - [1. ***edgeDevice***](#1-edgedevice)
        - [2. ***edgeNode***](#2-edgenode)
      - [软件组件（控制面）](#软件组件控制面)
        - [1. ***shifud***](#1-shifud)
        - [2. ***shifuController***](#2-shifucontroller)
      - [软件组件（数据面）](#软件组件数据面)
        - [1. ***deviceShifu***](#1-deviceshifu)
    - [架构图](#架构图)
      - [***deviceShifu*** 的生命周期](#deviceshifu-的生命周期)

# Shifu 高层设计

## 设计理念

### 以人为本

***Shifu*** 的最主要任务是让开发者和运维人员开心。所有 ***Shifu*** 的设计都应该以人为本，需要一些设计需求来达到这样的目标：

#### 1. 容易部署

***Shifu*** 永远可以被一条命令召唤（部署）。

#### 2. 即插即用

***Shifu*** 可以自动的认知并提供新加入的 IoT 设备的基本功能。当开发者完善 ***Shifu*** 接口之后所有设备设计的功能会立即可用。开发者可以进一步去完善 ***Shifu*** 的接口来创建定制化的功能来实现无限可能。

#### 3. 零维护

***Shifu*** 的目标是零维护。***Shifu*** 可以解决自己的问题来使运维人员更加轻松。

#### 4. 易用的 SDK

***Shifu*** 一直会向开发者提供非常易用的 SDK，因为 ***Shifu*** 想让开发者快乐！

## 设计目标与非目标

### 设计目标

#### 简单

拥有简单的架构和逻辑来保持高标准的可读性和可维护性至关重要。

易读和易更改的代码可以使开发者来使 ***Shifu*** 变得更好甚至创建他们自己版本的 ***Shifu*** 。

#### 高可用

作为 ***Shifu*** ，稳定性是必须的。一个适应力强的 ***Shifu*** 对于运维人员的精神健康非常重要，没有人会想每天面临崩溃的局面。

##### 1. 自我修复

***Shifu*** 在遇到突发事件时可以自我修复并将自己调整到目标状态。

##### 2. 稳定

***Shifu*** 目标是达到高可用性并降低运维开销。

#### 跨平台

***Shifu*** 可以运行在所有主要平台上，包括但不限于 x86/64, ARM64 等。

#### 轻量

作为一个可以运行在边缘的 IoT 框架，***Shifu*** 必须要轻量。***Shifu*** ，该减肥了！

#### 可扩展性和可变性

世界上有着太多异构的 IoT 设备。***Shifu*** 需要可扩展和可改变来应对不同情景。

### 设计非目标

#### 除 Kubernetes 的其他平台

我们的首要目标是保证 ***shifu*** 平稳的运行在 Kubernetes 上。未来我们会提供单独部署的方式。

#### 100% up-time

***Shifu*** 的设计内置了容错。***Shifu*** 努力达到 >99.9999% 的 up-time，但不是 100% 。

## 设计总览

当前版本的 ***Shifu*** 类似 [Kubernetes Operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)。因为利用了 Operator 的模式，因此 ***Shifu*** 拥有所有 Kubernetes 带来的好处。

### 组件

#### 物理组件

##### 1. ***edgeDevice*** 

***edgeDevice*** 是一个由 ***Shifu*** 管理的 IoT 设备。

##### 2. ***edgeNode***

***edgeNode*** 是一个可以连接到多个 ***edgeDevices*** 的 [Kubernetes 节点](https://kubernetes.io/docs/concepts/architecture/nodes/)。默认情况下，所有 Kubernetes 集群中的 worker 节点是 [taint](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) 了 ***edgeNode***。举例，用户可以配置节点作为一个非 ***edgeNode***, 来隔离应用 Pods 和 ***deviceShifu*** Pods。

#### 软件组件（控制面）

##### 1. ***shifud***

***shifud*** 是一个运行在每一个 ***edgeNode*** 上监控硬件变化的 daemonset，如 ***edgeDevice*** 连接/断开事件。 ***shifud*** 被 ***shifuController*** 管理。

##### 2. ***shifuController***

***shifuController*** 是一个接收 ***shifud*** 发送来的硬件事件，并做出相应动作来管理 ***deviceShifu*** 生命周期的 [Kubernetes controller](https://kubernetes.io/docs/concepts/architecture/controller/) 。

#### 软件组件（数据面）
##### 1. ***deviceShifu***

***deviceShifu*** 是一个 ***edgeDevice*** 的结构性[数字孪生](https://zh.wikipedia.org/wiki/%E6%95%B0%E5%AD%97%E6%98%A0%E5%B0%84)。之所以称其为**结构性**是因为它不光是一个 ***edgeDevice*** 的虚拟化表达，更可以将相应的 ***edgeDevice*** 驱动到目标状态。比如如果需要机械手臂来搬运一个箱子，但是当前它的状态为繁忙，***deviceShifu*** 会缓存你的命令并告诉机械手臂当有空时去搬运箱子。

***deviceShifu*** 会提供一些通用功能例如 ***edgeDevice*** 的健康状态检查，状态缓存等。通过实现 ***deviceShifu*** 的接口，***edgeDevice*** 可以实现它被设计的所有功能，并且更多！

***deviceShifu*** 有以下两种运行方式：
1. ***standalone mode***: ***standalone mode*** 是被设计为用来管理单个复杂的 ***edgeDevice*** ，比如机械手臂来提供高质量的一对一 ***edgeDevice*** 管理。
2. ***swarm mode***: ***swarm mode*** 是被设计用来管理多个简单且同类的 ***edgeDevice*** ，比如温度计来提供高效的一对多 ***edgeDevice*** 管理。

### 架构图

#### ***deviceShifu*** 的生命周期

1. **设备连接（用户操作不在下图中）**  
1.1 连接: ***edgeDevice*** 物理连接到 ***edgeNode*** 。  
1.2 设备连接: ***shifud*** 检测到设备连接事件，将事件发送给 ***shifuController*** 。  
1.3 创建: ***shifuController*** 创建一个 ***edgeDevice*** 的 ***standalone/swarm mode*** 的 ***deviceShifu*** 。  
1.4 管理: ***deviceShifu*** 开始管理新连接的 ***edgeDevice*** 。

当 ***edgeDevice*** 连接到 ***edgeNode*** 时, Shifu 会创建一个 ***deviceShifu*** , 通过 ***edgeDevice*** 的数字孪生来管理他。

[![shifu-device connect](/img/shifu-device-connect.svg)](/img/shifu-device-connect.svg)

2. **设备操作 | TODO: 统一 deviceShifu 接口**

在正常运行时，***shifud*** 和 ***shifuController*** 不会做太多事。用户直接和 ***deviceShifu*** 交互。比如，可以通过 ***deviceShifu*** 的API来获取设备的信息，健康状态等。因为是双向的，当开发者在 ***deviceShifu*** 的API中实现了 ***edgeDevice*** 的特定功能后，可以通过 ***deviceShifu*** 的 API 来管理 ***edgeDevice***。比如通过很少行的代码搭建摄像头的视频流。

[![shifu-device operating](/img/shifu-device-operating.svg)](/img/shifu-device-operating.svg)

3. **设备断连（用户操作不在下图中）**  
3.1 断开连接: 一个 ***edgeDevice*** 断开和 ***edgeNode*** 的物理连接。  
3.2 设备断开连接: ***shifud*** 检测到断开事件，将事件发送给 ***shifuController***。  
3.3 删除: ***shifuController*** 删除 ***edgeDevice*** 的 ***deviceShifu***。删除过程会因为清理持续一阵子。

[![shifu-device disconnect](/img/shifu-device-disconnect.svg)](/img/shifu-device-disconnect.svg)
