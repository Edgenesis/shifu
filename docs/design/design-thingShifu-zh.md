# ***thingShifu*** 设计文档
- [***thingShifu*** 设计文档](#thingshifu-设计文档)
  - [设计目标和非目标](#设计目标和非目标)
    - [设计目标](#设计目标)
      - [数字孪生](#数字孪生)
      - [轻松使用和部署](#轻松使用和部署)
      - [易扩展](#易扩展)
    - [非目标](#非目标)
      - [百分百准确表达](#百分百准确表达)
      - [自动修复硬件故障](#自动修复硬件故障)
  - [关于thingshifu](#关于thingshifu)
  - [关于deviceshifu](#关于deviceshifu)
  - [什么是thing](#什么是thing)
  - [组成部分](#组成部分)
    - [***thingshifu***核心](#thingshifu核心)
      - [1. ***bootstrapper***](#1-bootstrapper)
      - [2. ***core preparer***](#2-core-preparer)
      - [3. ***inbound instruction processor***](#3-inbound-instruction-processor)
      - [4. ***thing telemetry collector***](#4-thing-telemetry-collector)
      - [5. ***thingShifu client message sender***](#5-thingshifu-client-message-sender)
  - [运行模式](#运行模式)
    - [Swarm Mode](#swarm-mode)
  - [运行时的状态](#运行时的状态)
    - [1. ***creation state***](#1-creation-state)
    - [2. ***preparation state***](#2-preparation-state)
    - [3. ***running state***](#3-running-state)
    - [4. ***termination state***](#4-termination-state)
  - [标准结构](#标准结构)
  - [层级结构](#层级结构)
    - [YAML 配置示例](#yaml-配置示例)
        - [***thing*** 配置](#thing-配置)
  - [#Sample YAML of the Factory thing](#sample-yaml-of-the-factory-thing)
  - [#Sample YAML of the Streamline Engine 1 thing](#sample-yaml-of-the-streamline-engine-1-thing)
  - [#Sample YAML of the AR 1-1 thing](#sample-yaml-of-the-ar-1-1-thing)
        - [***thingShifu*** 配置](#thingshifu-配置)
  - [#Sample YAML of the Factory thingShifu](#sample-yaml-of-the-factory-thingshifu)
  - [#if shifu_mode is not specified, standalone will be used](#if-shifu_mode-is-not-specified-standalone-will-be-used)
  - [#Sample YAML of the AR 1-1 thingShifu](#sample-yaml-of-the-ar-1-1-thingshifu)
  - [#Sample YAML of the temperature sensor thingShifu in swarm mode](#sample-yaml-of-the-temperature-sensor-thingshifu-in-swarm-mode)
    - [命令的层级](#命令的层级)
    - [监测数据的层级](#监测数据的层级)
  - [分组](#分组)
  - [Sample YAML of the Group A](#sample-yaml-of-the-group-a)
  - [Race condition](#race-condition)
    - [命令的优先级](#命令的优先级)
  - [未来的计划](#未来的计划)
    - [消息队列](#消息队列)
    - [数据库](#数据库)

## 设计目标和非目标
### 设计目标
#### 数字孪生
***thingShifu*** 是一个 ***thing*** 的数字孪生。人们制造的任何东西都是一个 ***thing***.


#### 轻松使用和部署
通过编写简单的配置，用户可以利用 ***thingShifu*** 来用控制 ***thing*** 而不必担心任何硬件适配问题。 作为 ***Shifu*** 架构的一部分，很多个 ***thingShifu*** 可以很容易地被分组，通过用户的简单命令来实现更复杂的目标。

#### 易扩展
***thingShifu*** 没有限制，任何人都可以简单的添加新功能。

### 非目标
#### 百分百准确表达
***thingShifu*** 是一个 ***thing*** 的数字化表达，但是因为人类并不能100%了解现实世界，我们的 ***thingShifu*** 也不能100%表达一个 ***thing***.

#### 自动修复硬件故障
作为一个数字表达并不能修复现实世界中的电路设计或芯片故障等硬件问题。

## 关于thingshifu
***thingShifu*** 是一个 ***thing*** 的数字孪生，是 ***thing*** 的数字表达。它是 ***Shifu*** 框架中距离终端用户最近的一个组件，可以使开发和运维人员通过简单的API来控制 ***thing*** 并让运维人员很容易的知晓设备当前的状态。

## 关于deviceshifu
***deviceShifu*** 是 ***thingShifu*** 的针对机器设备的一个全功能子集。

## 什么是thing
一个 ***thing*** 可以是任意人为制造的物件，它可以小到一个电路板，一枚芯片，一个摄像头，一个手机，一个机器人，一辆车。大到一栋楼，一座工厂，一条街，一个城市。

***thingShifu*** 的终极目标是成为任何人类制造的设备的“师傅”（老师或者教练）来让他们更加聪明的来服务于人类。在当前阶段，我们关注于让 ***thingShifu*** 来表达一个或多个IoT设备

## 组成部分

### ***thingshifu***核心

#### 1. ***bootstrapper***
***thingShifu*** 创建的入口，它读取创建请求，从Kubernetes apiServer加载配置，建立与 ***thing*** 的连接。

输入: 来自shifuController 的创建请求

输出: thingShifu core 准备完成

#### 2. ***core preparer***
读取并处理配置文件，使 ***thingShifu*** 可以接收命令并收集监测数据。

输入: ***thing*** 的配置: 由开发者判断启动 ***thing*** 有用的信息, 比如:
   1. 可用命令，比如 ```[device ip]/deviceMovement <x-y coordinations>```
   2. 需要收集的监测数据（名称和可选的收集区间），比如 ```device_on, device_off, device_health```
   3. 可读性高的命令到 ***thing*** 原生命令的映射关系


输出（缓存在内存）：
   1. 可用指令
   2. 需要收集的监测数据

#### 3. ***inbound instruction processor***
将开发者的命令翻译

输入: 命令（开发者实时输入）

输出: 可以被 ***thing*** 直接接收的命令

#### 4. ***thing telemetry collector***
不断从 ***thing*** 来收集检测数据（状态/健康度/更新）
- 最简单的方法: 向 ***thing*** 发送一个命令，得到回复
- 其他方法: 从其他的频道订阅更新，如REST API等
- 至少要有一个 “ping” 方法来确认连接到了 ***thing***

#### 5. ***thingShifu client message sender***
- ***thingShifu*** 预计有一个客户端（UI)来使操作人员很简单的来和 ***thingShifu*** 交互
- 连接到 ***thingShifu*** 核心，以代理命令/指标更新来服务
- 一个单独的线程服务于发送消息到客户端（UI）

## 运行模式
**standalone mode**: ***thingShifu*** 管理着一个 ***thing***。例如一个温度计。

**swarm mode**: ***thingShifu*** 管理着多个同类型的 ***thing***，一如一组温度计。

默认情况下，***thingShifu*** 处于**standalone mode**.


### Swarm Mode
下面是一个典型的 **Swarm Mode**样例：

[![thingShifu factory example swarm mode](/img/thingShifu/shifu-thingShifu-example-factory-swarm.svg)](/img/thingShifu/shifu-thingShifu-example-factory-swarm.svg)


## 运行时的状态
### 1. ***creation state***
***shifuController*** 负责创建 ***thingShifu***, 触发 ***bootstrapper***.

当 ***bootstrapper*** 汇报准备完成之后，开始下一步

### 2. ***preparation state***
当 ***bootstrapper*** 汇报准备完成，会触发 ***core preparer***。它会读取配置文件并将所有配置载入到内存中。

接着会进行下一步

### 3. ***running state***
***thingShifu*** 核心开始启动，它会进行下列操作：
 - 周期性去ping ***thing***
 - 周期性收集监测数据
 - 保持指令接受端口开放

### 4. ***termination state***
***shifuController*** 会停止 ***thingShifu*** ，然后释放内存并汇报“terminated”消息。

## 标准结构

[![thingShifu basic structure](/img/thingShifu/shifu-thingShifu-basic-structure.svg)](/img/thingShifu/shifu-thingShifu-basic-structure.svg)

## 层级结构
一个运行的 ***thing*** 之下可能有多层隶属于它的更低层级的 ***thing***。所以 ***thingShifu*** 可以对 ***thing*** 划分层级。 每一个 ***thingShifu*** 会基于到达时间和优先级来执行命令。

假设一个工厂有三条流水线，每个流水线安装了两种机器人：组装机器人（AR）和移动机器人（TR）。

这个架构中，所有的设备都是 ***thing***。用户可以将 ***things*** 的架构定义成如下：

[![thingShifu factory example](/img/thingShifu/shifu-thingShifu-example-factory-structure.svg)](/img/thingShifu/shifu-thingShifu-example-factory-structure.svg)

用户可以将 ***thingShifu*** 的架构定义成如下：

[![thingShifu factory example with thingShifu](/img/thingShifu/shifu-thingShifu-example-factory-thingShifu.svg)](/img/thingShifu/shifu-thingShifu-example-factory-thingShifue.svg)

根据上述的结构定义，层级为：***Factory thingShifu*** 是最上层，它有着两个更低层的 ***thingShifu***：***Streamline Engine 1 thingShifu*** 和 ***Streamline Engine 2 thingShifu***. 每个流水线引擎 ***thingShifu*** 下有着四个 ***thingShifu***， 代表着四个机器人。

注： 结构配置可以由用户任意定义，所以有可能是完全不同的结构

### YAML 配置示例

***thingShifu*** 架构预计会有两种配置：
- 每个 ***thingShifu*** 的 YAML 配置
- 每个 ***thing*** 的YAML 配置

用户定义的层级配置会是如下的格式：
##### ***thing*** 配置
```` 
#Sample YAML of the Factory thing
---
thing: "Factory"
thing_sku: "Factory General SKU"
thing_id: 1 # generated by thingShifu
thing_type: non-end-device
...
#Sample YAML of the Streamline Engine 1 thing
---
thing: "Streamline Engine 1"
thing_sku: "Streamline General SKU"
thing_id: 11 # generated by thingShifu
thing_type: non-end-device
...
#Sample YAML of the AR 1-1 thing
---
thing: "AR 1-1"
thing_sku: "Assemble Robot SKU"
thing_id: 11 # generated by thingShifu
thing_type: end-device
thing_address: edgesample06
thing_port: 8000
...
````
##### ***thingShifu*** 配置

````
#Sample YAML of the Factory thingShifu
---
thing: "Factory"
shifu_mode:standalone
shifu_id: 10001 # generated by thingShifu
instruction: "start", "halt", "stop"
child_things: ["Streamline Engine 1", "Streamline Engine 2"]
telemetry: []
...
#Sample YAML of the Streamline Engine 1 thingShifu
#if shifu_mode is not specified, standalone will be used
---
thing: "Streamline Engine 1"
shifu_id: 12001 # generated by thingShifu
instruction: "start", "halt", "stop"
child_things: ["AR 1-1", "AR 1-2", "TR 1-1", "TR 1-2"]
telemetry: ["engine_state", "engine_location"]
...
#Sample YAML of the AR 1-1 thingShifu
---
thing: "AR 1-1"
shifu_id: 15001 # generated by thingShifu
instruction: "start", "halt", "stop", "moveRobotArm", "rotateRobotArm", "gripPart"
child_things: ["robot_state", "robot_location", "robot_lastmove"]
...
#Sample YAML of the temperature sensor thingShifu in swarm mode
---
thing: "temperature sensor"
mode:swarm
shifu_id: 90001 # generated by thingShifu
instruction: "start", "halt", "stop", "reset"
child_things: []
things_in_swarm: ["TS 1", "TS 2", "TS 3", "TS 4", "TS 5", "TS 6"]
...
````

### 命令的层级
***thingShifu*** 允许用户通过一条命令轻松执行在所有层级的通用命令。

比如 **start**, **halt** 和 **stop** 命令（需要用户实现具体执行方法）。

根据上述结构，开发者只需发送这类命令到 ***工厂 thingShifu***，之后所有 ***工厂 thingShifu*** 下的设备就都会收到这个命令，并根据自己的逻辑去执行。

比如，如果向 ***工厂 thingShifu*** 发送一个 **start** 命令，所有 ***工厂 thingShifu*** 层级下的 ***thing*** 会执行它们自身的 **start** 逻辑：工厂本身会进入到“启动”状态，流水线引擎1和2会开始移动，流水线下的组装机器人会置到“就绪”状态，移动机器人会移动到它们的起始位置等待装载配件。

### 监测数据的层级
用户可以定义下层 ***thingShifu*** 会上报到上层的监测数据。

比如，如果希望 ***Factory thingShifu*** 来汇报每个机器人的当前状态，开发者需要配置机器人的thingShifu 汇报给流水线的thingShifu. 指标的汇报层级如下：

***AR/TR thingShifu*** -> ***Streamline thingShifu*** -> ***Factory thingShifu***.

## 分组
分组可以被想象成一个“平行层”，但不局限于此。在现实中，用户来自由定义多个 ***thingShifu*** 来使命令可以同时达到它们，并以组来收集指标。

分组和层级的区别是：组没有一个“最高”或者“顶级”的 ***thingShifu***。所有同组的 ***thingShifu*** 将会视为同层。

分组和Swarm mode的区别是：分组是一组 ***thingShifu*** 且可以解散。Swarm mode只是一个管理着多个 ***thing*** 的 ***thingShifu*** 且不能被分成多个 ***thingShifu***.

分组的例子：
[![thingShifu factory example grouping](/img/thingShifu/shifu-thingShifu-example-factory-grouping.svg)](/img/thingShifu/shifu-thingShifu-example-factory-grouping.svg)

在这个结构中，**Group A** 包含着TR 1-1, TR 1-2, AR 2-2 和 TR 2-1。所以当发送 **halt** 命令到 **Group A** 中时，这4个 ***thingShifu*** 会执行这个命令

***分组*** 的配置如下：
```` 
Sample YAML of the Group A
---
group: "Group A"
id: 1
thingShifu_in_group:["TR 1-1", "TR 1-2", "AR 2-2", "TR 2-1"]
...
````

用户不需要提供给分组的命令，因为一个分组内部所有***thingShifu*** 都接受的通用命令会被提取出来。


## Race condition
***thingShifu*** 会在执行命令的时候对所有资源加上写入锁来避免其他命令的改动。

当多个命令同时到达时，***thingShifu*** 会先尝试利用优先级来执行。如果所有命令都是同一个等级的话，***thingShifu*** 会随机挑选一个然后加锁执行。

### 命令的优先级
用户可以定义命令为至高优先级（等级 -1），重要等级为 -1 的命令可以中断所有正在执行的命令并强制执行此命令。

直接从用户到 ***thingShifu*** 的优先级永远是最高（等级 0 ）。

从用户到分组的命令永远有第二高的优先级（等级 1）。

在一个多层级的结构里，从更高层级传达给下面层级的命令，***thingShifu*** 会将该命令优先级视为2以及更低（由用户定义）。

## 未来的计划
### 消息队列
当命令和监测数据量很大时，用户可以考虑在 ***thingShifu*** 的层级添加一个消息队列。

### 数据库
***thingShifu*** 并没有预见很大的数据量所以它将所有存储放到内存中，如需要的话，可以添加一个独立的数据库。
