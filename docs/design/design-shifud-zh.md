- [***shifud*** 高层设计](#shifud-高层设计)
  - [介绍:](#介绍)
  - [设计理念](#设计理念)
    - [自动 & 自主](#自动--自主)
      - [1. 自动发现 ***edgeDevice***:](#1-自动发现-edgedevice)
      - [2. 最简化的发现不支持自动发现的 ***edgeDevice***:](#2-最简化的发现不支持自动发现的-edgedevice)
  - [设计目标和非目标](#设计目标和非目标)
    - [设计目标](#设计目标)
      - [自主](#自主)
      - [轻量](#轻量)
      - [扩展性](#扩展性)
    - [设计非目标](#设计非目标)
  - [设计总览](#设计总览)
    - [组件](#组件)
      - [软件组件](#软件组件)
        - [1. ***deviceDiscoverer***](#1-devicediscoverer)
        - [2. ***deviceVerifier***](#2-deviceverifier)
        - [3. ***deviceUpdater***](#3-deviceupdater)
    - [***shifud*** 输入 & 输出](#shifud-输入--输出)
      - [结构图](#结构图)
      - [***shifud*** 的执行流(集群):](#shifud-的执行流集群)
      - [***shifud***的执行流(***edgeNode***):](#shifud的执行流edgenode)

# ***shifud*** 高层设计

## 介绍:
本文档是 ***Shifu*** 架构中的 ***shifud*** 组件的高层设计。 ***shifud*** 是一个运行在每个 ***edgeNode*** 上的 DaemonSet。它会从Kubernetes的 ***edgeDevice*** 资源中发现设备并将相关资源更新到apiServer中。

## 设计理念
### 自动 & 自主
***shifud*** 的首要任务是将 ***edgeDevice*** 的发现和校验尽可能简单化。开发者不应该需要过多的配置来让它们的 ***edgeDevice*** 在 ***Shifu*** 中可用。下面是一些要求：

#### 1. 自动发现 ***edgeDevice***:
***shifud*** 可以发现如ONVIF或者类似协议的 ***edgeDevice***, 不需要用户/开发者的过多介入。

#### 2. 最简化的发现不支持自动发现的 ***edgeDevice***:
开发者只需要提供必须的信息来使 ***shifud*** 去发现一个特定的设备。

## 设计目标和非目标
### 设计目标
#### 自主
***shifud*** 可以在 ***Shifu*** 框架安装后自己运行。

#### 轻量
***shifud*** 会最小化每一个 ***edgeNode*** 合在整个集群中的内存消耗。

#### 扩展性
***shifud*** 可以接入大部分的IoT协议。

### 设计非目标
[TODO]

## 设计总览
  

### 组件

#### 软件组件

##### 1. ***deviceDiscoverer***
***deviceDiscoverer*** 是一个持续扫描 ***edgeNode*** 设备事件的进程，包括但不限于网络连通性，USB事件。

##### 2. ***deviceVerifier***
***deviceVerifier*** 是一个与 ***edgeDevice*** 交互的进程，会尝试获取并校验设备的信息来和Kubernetes集群中添加的 ***edgeDevice*** 来进行比对。

##### 3. ***deviceUpdater***
***deviceUpdater*** 会根据 ***edgeDevice*** 的校验状态通过apiServer更新  ***edgeDevice***  的资源。

### ***shifud*** 输入 & 输出
***shifud*** 的输入输出总览可以被总结为下图：
[![shifud input and output overview](/img/shifud-input-output.svg)](/img/shifud-input-output.svg)    

***shifud*** 来自Kubernetes ***edgeDevice*** 资源的输入应该是一个 ***edgeDevice*** 列表：
```
apiVersion: v1
kind: edgeDevice
metadata:
  name: franka-emika-1
spec:
- sku: "Franka Emika"
  connection: Ethernet
  status: offline
  address: 10.0.0.1:80
  protocol: HTTP
  disconnectTimeoutInSeconds:600 # optional
  group:["room1", "robot"] # optional
  driverSpec: # optional when no driver is required
  - instructionMap:  # optional
      move_to:
      - api: absolute_move # API of the driver
......
```

#### 结构图
[![shifud design overview](/img/shifud-design-overview.svg)](/img/shifud-design-overview.svg)    


#### ***shifud*** 的执行流(集群):
1. 当请求到设备列表时，***deviceDiscoverer*** 会开始通过互联网协议扫描本地，应支持以下协议：
   ```
   ONVIF
   SNMP
   MQTT
   OPC UA
   PROFINET
   ```

#### ***shifud***的执行流(***edgeNode***):
1. 当请求到设备列表时， ***deviceDiscoverer*** 开始用不同协议扫描本地，应支持以下协议:
   ```
   udev
   MODBUS
   ```
2. 发现过程根据使用协议：
    1. 对于TCP/IP的 ***edgeDevice*** ，可以直接使用ping或者TCP connect。
    2. 对于udev/USB类型的设备，***deviceDiscoverer*** 会利用Linux本身的 [udev](https://man7.org/linux/man-pages/man7/udev.7.html) 工具。
3. 设备被发现后，***deviceVerifier*** 会开始通过连接协议校验设备与设备列表的信息。
    ```
    sample udevadm output:
    E: DEVNAME=/dev/video3
    E: SUBSYSTEM=video4linux
    E: ID_SERIAL=Sonix_Technology_Co.__Ltd._USB_2.0_Camera_SN0001
    ```
4. 校验结束后，***deviceUpdater*** 会通过Kubernetes的apiServer更新 ***edgeDevice*** 的资源。
    ```
   apiVersion: v1
   kind: edgeDevice
   metadata:
     name: franka-emika-1
   spec:
   - sku: "Franka Emika"
     connection: Ethernet
     status: online
     address: 10.0.0.1:80
     protocol: HTTP
     disconnectTimeoutInSeconds:600 # optional
     group:["room1", "robot"] # optional
     driverSpec: # optional when no driver is required
     - instructionMap:  # optional
         move_to:
         - api: absolute_move # API of the driver
   ......
    ```
