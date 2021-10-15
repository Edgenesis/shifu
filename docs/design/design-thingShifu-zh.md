# ***thingShifu*** Design Document
- [thingShifu Design Document](#thingshifu-design-document)
  - [What is thingShifu?](#what-is-thingshifu)
  - [What is "thing"?](#what-is-thing)
  - [thingShifu Components](#thingshifu-components)
    - [thingShifu Core](#thingshifu-core)
      - [1. bootstrapper](#1-bootstrapper)
      - [2. core preparer](#2-core-preparer)
      - [3. inbound instruction processor](#3-inbound-instruction-processor)
      - [4. thing telemetry collector](#4-thing-telemetry-collector)
      - [5. thingShifu client message sender](#5-thingshifu-client-message-sender)
  - [Operation Modes of thingShifu](#operation-modes-of-thingshifu)
    - [Swarm Mode](#swarm-mode)
  - [States of thingShifu](#states-of-thingshifu)
    - [1. creation state](#1-creation-state)
    - [2. preparation state](#2-preparation-state)
    - [3. running state](#3-running-state)
    - [4. termination state](#4-termination-state)
  - [Structure](#structure)
  - [Hierarchy of thing and thingShifu](#hierarchy-of-thing-and-thingshifu)
    - [Sample YAML configuration](#sample-yaml-configuration)
    - [Hierarchy of Instructions](#hierarchy-of-instructions)
    - [Hierarchy of Telemetries](#hierarchy-of-telemetries)
  - [Grouping of thingShifu](#grouping-of-thingshifu)
  - [Sample YAML of the Group A](#sample-yaml-of-the-group-a)
  - [Limitations and External Components Can Be Added](#limitations-and-external-components-can-be-added)
    - [Message Queue](#message-queue)
    - [Database](#database)

## 设计目标和非目标
### 设计目标
#### 数字孪生
***thingShifu*** 是一个 ***thing*** 的数字孪生。 用数字代表现实世界中任何人为制造的东西

#### 容易使用和部署
通过简单的编写配置，用户可以利用 ***thingShifu*** 来很容易的用软件方式，不需要担心任何硬件问题的控制 ***thing***。 作为 ***Shifu*** 架构的一部分，很多个 ***thingShifu*** 可以很容易的被分组，通过用户的简单命令来实现更复杂的目标

#### 容易扩展
***thingShifu*** 没有限制，任何人都可以简单的添加新功能

### 设计非目标
#### 100% 准且表达
***thingShifu*** 是一个 ***thing*** 的数字化表达，但是因为人类并不能100%理解显示，我们的 ***thingShifu*** 也不能100%表达一个 ***thing***

#### 自动修复硬件问题
如果一个现实世界中的 ***thing*** 出现了问题如电路设计或芯片故障， ***thingShifu*** 作为一个数字表达并不能帮助什么

## 什么是 ***thingShifu***?
***thingShifu*** 是一个 ***thing*** 的数字孪生。他是 ***Shifu*** 框架中作为 ***thing*** 的完全表达距离终端用户最近的一个组件。它可以使开发和运维人员通过简单的API来控制 ***thing*** 并让运维人员很容易的知晓他的当前状态

## 什么是 "thing"?
一个 ***thing*** 可以是任意人为制造的物件，它可以小到一个电路板，一枚芯片，一个摄像头，一个手机，一个机器人，一辆车。达到一栋楼，一座工厂，一条街，一个城市

***thingShifu*** 的终极目标是成为任何人类制造的设备的“师傅”（老师或者教练）来是他们更加聪明的来服务于人类。在当前阶段，我们关注于让 ***thingShifu*** 来表达一个或多个IoT设备

## ***thingShifu*** 组件
  
### ***thingShifu*** 核心

#### 1. ***bootstrapper***
***thingShifu*** 创建的入口， 读取创建请求，从Kubernetes apiServer加载配置，建立与 ***thing*** 的连接

输入: shifuController 的创造请求

输出: 准备启动 thingShifu core

#### 2. ***核心准备器***
读取并处理配置文件，使 ***thingShifu*** 可以接收命令并收集指标。将可用的API通过客户消息发送到客户端

输入: ***thing*** 的配置: 由开发者判断启动 ***thing*** 有用的信息, 比如:
   1. 可用命令，比如 ```[device ip]/deviceMovement <x-y coordinations>```
   2. 需要收集的指标（指标名称和可选的收集区间），比如 ```device_on, device_off, device_health```
   3. 可读性高的命令到 ***thing*** 接受命令的映射关系


输出（缓存在内存）：
   1. 可用指令
   2. 收集的指标

#### 3. ***命令输入处理器***
将开发者的命令翻译

输入: 命令（开发者实时输入）

输出: 可以被 ***thing*** 直接接收的命令

#### 4. ***thing 指标收集器***
不断从 ***thing*** 来收集指标（状态/健康度/更新）
- 最简单的方法: 向 ***thing*** 发送一个命令，得到回复
- 其他方法: 从其他的频道订阅更新，如REST API等
- 至少要有一个 “ping” 方法来确认连接到了 ***thing***

#### 5. ***thingShifu 客户消息发送端***
- ***thingShifu*** 预计有一个客户端（UI)来使操作人员很简单的来和 ***thingShifu*** 交互
- 连接到 ***thingShifu*** 核心，以代理命令/指标更新来服务
- 一个单独的线程服务于发送消息到客户端（UI）

## ***thingShifu*** 的工作方式
**独立模式**: ***thingShifu*** 管理着一个 ***thing***。例如一个温度计

**群模式**: ***thingShifu*** 管理着多个同类型的 ***thing***，一如一组温度计

默认情况下，***thingShifu*** 处于**独立模式**

 
### 群模式
一个典型的 **群模式**会是如下：

[![thingShifu factory example swarm mode](/img/thingShifu/shifu-thingShifu-example-factory-swarm.svg)](/img/thingShifu/shifu-thingShifu-example-factory-swarm.svg)


## ***thingShifu*** 的状态
### 1. ***创建状态***
***shifuController*** 负责创建 ***thingShifu***, 触发 ***bootstrapper***.

当 ***bootstrapper***，去下一步

### 2. ***准备状态***
当 ***bootstrapper*** 汇报准备完成，会触发 ***core preparer***。然后他会读取配置文件并将所有载入到内存中

准备完成后，去下一步

### 3. ***运行状态***
当准备完成时，***thingShifu*** 核心会启动：
 - 周期性去ping ***thing***
 - 周期性的手机指标
 - 保持指令接受端口常开

### 4. ***终结状态***
***shifuController*** 会使 ***thingShifu*** 终结，然后他会释放内存，汇报“终结”消息

## 结构

[![thingShifu basic structure](/img/thingShifu/shifu-thingShifu-basic-structure.svg)](/img/thingShifu/shifu-thingShifu-basic-structure.svg)

## ***thing*** 和 ***thingShifu*** 的层级
一个运行的 ***thing*** 很可能使有多层底层的 ***thing***。所以 ***thingShifu*** 允许层级的 ***things*** 的命令和指标。 每一个 ***thingShifu*** 会基于命令的到达时间和重要度来运行

比如一个工厂有三条流水线，每个流水线安装了两种机器人：组装机器人（AR）和移动机器人（TR）

这个架构中，所有的设备都是 ***thing***。用户可以将 ***things*** 的架构定义成如下：

[![thingShifu factory example](/img/thingShifu/shifu-thingShifu-example-factory-structure.svg)](/img/thingShifu/shifu-thingShifu-example-factory-structure.svg)

用户可以将 ***thingShifu*** 的架构定义成如下：

[![thingShifu factory example with thingShifu](/img/thingShifu/shifu-thingShifu-example-factory-thingShifu.svg)](/img/thingShifu/shifu-thingShifu-example-factory-thingShifue.svg)

根据上述的结构定义，层级为：***工厂 thingShifu*** 是最上层，它有着两个更低层的 ***thingShifu***：***流水线引擎 1 thingShifu*** 和 ***流水线引擎 2 thingShifu***. 每个流水线引擎 ***thingShifu*** 下有着四个 ***thingShifu***， 代表着四个机器人

注： 结构配置是由用户任意想象，所以有可能是完全不同的结构

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
***thingShifu*** 允许用户通过一条命令很简单的执行在所有层级的通用命令

比如 **start**, **halt** 和 **stop** 命令（用户来实现）

根据上述结构，开发者只需发送这类命令到 ***工厂 thingShifu***，之后所有 ***工厂 thingShifu*** 下的设备就都会收到这个命令，并根据自己的逻辑去执行

比如，如果向 ***工厂 thingShifu*** 发送一个 **start** 命令，所有 ***工厂 thingShifu*** 层级下的 ***thing*** 会执行它们自身的 **start** 逻辑：工厂本身会进入到“启动”状态，流水线引擎1和2会开始移动，流水线下的组装机器人会置到“就绪”状态，移动机器人会移动到它们的起始位置等待装载配件

### 指标的层级
用户来定义下层 ***thingShifu*** 会上报到上层的指标

比如，如果希望 ***工厂 thingShifu*** 来汇报每个机器人的当前状态，开发者需要配置 ***机器人 thingShifu*** 汇报给 ***流水线 thingShifu***。指标的汇报层级会是如下：

***AR/TR thingShifu*** -> ***流水线 thingShifu*** -> ***工厂 thingShifu***.

## ***thingShifu*** 的分组
分组可以被想象成一个“平行层”，但不局限于此。在现实中，用户来自由定义多个 ***thingShifu*** 来使命令可以同时达到它们，并以组来收集指标。

分组和层级的区别是：组没有一个“最高”或者“顶级”的 ***thingShifu***。所有同组的 ***thingShifu*** 将会视为同层。

分组和群模式的区别是：分组时一组 ***thingShifu*** 且不能被解散。群模式只是一个管理着多个 ***thing*** 的 ***thingShifu*** 且不能被分成多个 ***thingShifu*** 

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

用户不需要提供给分组的命令，因为 ***thingShifu*** 的通用命令会被提取出来


## 竞争条件
***thingShifu*** 会在执行命令的时候对所有资源加上写入锁来避免其他命令的改动

当多个命令同时到达时，***thingShifu*** 会先尝试利用重要等级来执行。如果所有命令都是同一个等级的话，***thingShifu*** 会随机挑选一个然后加锁执行

### 命令的重要等级
用户可以定义命令至高重要等级（等级 -1），重要等级为 -1 的命令可以中断所有正在执行的命令

直接从用户到 ***thingShifu*** 的命令等级永远是最高（等级 0 ）

从用户到分组的命令永远会是第二重要等级（等级 1）

从更高层级传达下来的命令，***thingShifu*** 会将该命令视为更高等级（等级2，用户定义）

## 可以添加的限制和外部组件
### 消息队列
当命令和指标量很大时，用户可以考虑在 ***thingShifu*** 的层级添加一个消息队列

### 数据库
***thingShifu*** 并没有预见很大的数据量所以它将所有存储放到内存中，如需要的话，可以添加一个独立的数据库
