# ***Shifu***专有名词Markdown格式规范

## 规范

- 下述列出的特殊专有专有名词在**标题和正文**中使用斜体加粗
- 除了特殊专有名词，其他专有名词**正文**使用`code`标签，**标题**无格式
- **英文大小写**与shifu代码仓库Go语言代码中的定义完全一致
- 普通英语词不加格式；与编程相关的英语词加`code`标签

### 特殊专有名词

- ***Shifu***
    - 介绍：基于`Kubernetes`的高效物联网设备管理开发框架
    - 注意：不存在`Shifu Framework`这种说法，只使用***Shifu***
    - 不指代该框架时，使用shifu即可
- ***Shifu Cloud***
    - 介绍：供物联网开发者快速使用***Shifu***的一站式平台
- 物理组件
    - ***edgeDevice***（***edgeDevice*** 是一个由 ***Shifu*** 管理的 IoT 设备）
    - ***edgeNode***（***edgeNode*** 是一个可以连接到多个 ***edgeDevices*** 的 `Kubernetes node`）
    - ***edgeMap***
- 数据面
    - ***deviceShifu***（***deviceShifu***是一个 ***edgeDevice*** 的结构性数字孪生）
- 控制面
    - ***shifud***（***shifud*** 是一个运行在每一个 ***edgeNode*** 上监控硬件变化的 `DaemonSet`，如 ***edgeDevice*** 连接/断开事件）
    - ***shifuController***（***shifuController*** 是一个接收 ***shifud*** 发送来的硬件事件，并做出相应动作来管理 ***deviceShifu*** 生命周期的 `Kubernetes controller` ）

### 特定专有名词

- 与Kubernetes保持一致
    - [DaemonSet](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/)

## 中英语言规范

- 中文
    - 复数：专有名词复数描述不加`s`
    - 空格：***Shifu*** 等词，左右为`，` `。`等标点时不添加空格，其他情况加空格。
- 英文
    - 复数：专有名词复数描述加`s`，但是`s`不加格式（如***edgeDevice***s）
