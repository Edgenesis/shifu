# 【暂未实现】shifuController 设计
- [【暂未实现】shifuController 设计](#暂未实现shifucontroller-设计)
  - [设计目标和非目标](#设计目标和非目标)
    - [设计目标](#设计目标)
      - [低资源消耗](#低资源消耗)
      - [高可用](#高可用)
      - [无状态](#无状态)
      - [最小权限](#最小权限)
    - [设计非目标](#设计非目标)
      - [***shifud*** 管理](#shifud-管理)
      - [***edgeDevice*** 管理](#edgedevice-管理)
  - [设计大纲](#设计大纲)
    - [响应***edgeDevice*** 事件](#响应edgedevice-事件)
      - [1. 创建 ***edgeDevice***](#1-创建-edgedevice)
      - [2. 删除 ***edgeDevice***](#2-删除-edgedevice)
    - [响应***edgeNode*** 事件](#响应edgenode-事件)
      - [1. 创建 ***edgeNode***](#1-创建-edgenode)
      - [2. 删除 ***edgeNode***](#2-删除-edgenode)
    - [正常运行时](#正常运行时)
    - [***deviceShifu*** 对象](#deviceshifu-对象)
      - [简单案例：温度计](#简单案例温度计)
        - [温度计部署示例](#温度计部署示例)
        - [温度计服务示例](#温度计服务示例)
    - [***edgeMap*** 设计](#edgemap-设计)
      - [***edgeMap*** 数据结构](#edgemap-数据结构)
      - [***edgeVertex*** 数据结构](#edgevertex-数据结构)

***shifuController*** 的主要责任是管理 ***deviceShifu*** 的生命周期。

***shifuController*** 会通过创建/删除相应 ***deviceShifu*** 实例来响应通过 ***apiServer*** 发送的 ***edgeDevice*** 和 ***edgeNode*** 事件。

## 设计目标和非目标

### 设计目标

#### 低资源消耗

***shifuController*** 可以跑在云端和边缘，所以 ***shifuController*** 需要尽可能地降低内存消耗 —— 因此***shifuController*** 的内存占用低于100MB。

#### 高可用
作为 ***shifu*** 的控制层， ***shifuController*** 需要拥有高可用性。这是通过[Kubernetes deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) 和 [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) 来实现的。

#### 无状态
***shifuController*** 会将持久存储的数据放到etcd（或任意Kubernetes后台持久化存储）中来达到真正无状态。

#### 最小权限

***shifuController*** 应该总是尽可能地降低权限需求来减少对其他微服务的潜在影响并保持高安全标准。

### 设计非目标

#### ***shifud*** 管理

***shifud*** 是一个[Kubernetes daemonSet](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/)， 通过Kubernetes管理。

#### ***edgeDevice*** 管理

***shifuController*** 只管理 ***deviceShifu***， 而不管理 ***edgeDevice*** 和 ***edgeNode*** 。

## 设计大纲

***shifuController*** 作为一个 [Kubernetes controller](https://kubernetes.io/docs/concepts/architecture/controller/)， 通过[Kuberenetes CRD(Custom Resource Definition)](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/)管理整个 ***deviceShifu*** 部署的生命周期。

***shifuController*** 内部会通过[adjacency list](https://en.wikipedia.org/wiki/Adjacency_list#:~:text=In%20graph%20theory%20and%20computer,particular%20vertex%20in%20the%20graph.)缓存当前网络拓扑。 这个adjacency list叫做 ***edgeMap***， 未来 ***edgeMap*** 将会被用来渲染拓扑图给开发者/运维人员。

### 响应***edgeDevice*** 事件
#### 1. 创建 ***edgeDevice***

在任何情况下，当 ***shifuController*** 收到 ***edgeDevice*** 连接事件时，***shifuController*** 会：

1.1 Schedule: 决定把准备中的 ***deviceShifu*** 放到哪里。   
    1.1.1 如果 ***edgeDevice*** 是连接到某个 ***edgeNode***， ***deviceShifu*** 就会被调度到那个 ***edgeNode*** 上。  
    1.1.2 如果 ***edgeDevice*** 是连接到集群的网络上， ***deviceShifu*** 就会根据如下优先级被调度（非最终设定）：  
      a. 位置: 如果位置信息可用，***deviceShifu*** 就会被放在最靠近 ***edgeDevice*** 的 ***edgeNode*** 上。  
      b. 可用资源: ***deviceShifu*** 会被放在拥有最高可用内存的 ***edgeNode*** 上。
   
1.2 Compose: 整合所有该 ***deviceShifu*** 部署的的计算和编排信息。

1.3 Create:  
   1.3.1 通过向 ***apiServer*** 提交请求来添加 ***deviceShifu*** 部署。  
   1.3.2 将新创建的 ***deviceShifu*** 通过Kubernetes service暴露出来。  
   1.3.3 Add: 把 ***edgeDevice*** 添加到 ***edgeMap*** 。

#### 2. 删除 ***edgeDevice***

当 ***shifuController*** 收到 ***edgeDevice*** 连接断开事件时，***shifuController*** 会：

2.1 Remove: 将 ***edgeDevice*** 从 ***edgeMap*** 中移除。 

2.2 Delete:   
   2.2.1 通过向 ***apiServer*** 提交请求来移除 ***deviceShifu*** 部署。  
   2.2.2 删除 ***deviceShifu*** 的相关 Kubernetes service 。

### 响应***edgeNode*** 事件

#### 1. 创建 ***edgeNode***

在任何情况下，当 ***shifuController*** 收到 ***edgeNode*** 连接事件时，***shifuController*** 会：  

1.1 Add: 把 ***edgeNode*** 添加到 ***edgeMap*** 。

#### 2. 删除 ***edgeNode***

当 ***shifuController*** 收到 ***edgeNode*** 删除事件时，***shifuController*** 会：  

2.1 Remove: 将 ***edgeNode*** 从 ***edgeMap*** 中移除。

### 正常运行时

1. Collect: ***shifuController*** 会间断性收集 ***edgeNodes*** 的资源信息， 这个信息会被用来：
   1. ***edgeMap*** 做可视化用途。
   2. Reschedule，见下方。
2. Reschedule: ***shifuController*** 会在以下事件中重规划 ***deviceShifu***：
   1. ***shifuController*** 通过上述重规划算法给某个 ***deviceShifu*** 找到了一个更适合的位置。
   2. ***edgeNode*** 为不可用状态。 当一个 ***edgeNode*** 不可用时， ***shifuController*** 会尝试用最优方式重调度所有该 ***edgeNode*** 上的 ***deviceShifu*** 实例到其他 ***edgeNode*** 上。

### ***deviceShifu*** 对象

***deviceShifu*** 对象是一个通过 ***shifuController*** 管理的 ***edgeDevice*** 的 Kubernetes deployment 和 Kubernetes service 绑定。 用户**不应该**手动管理 ***deviceShifu***.

#### 简单案例：温度计

##### 温度计部署示例
下面是为一个简单温度计的 ***deviceShifu*** 部署 yaml 文件示例：

```
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: edgedevice-thermometer-deployment
  name: edgedevice-thermometer-deployment
  namespace: deviceshifu
spec:
  replicas: 1
  selector:
    matchLabels:
      app: edgedevice-thermometer-deployment
  template:
    metadata:
      labels:
        app: edgedevice-thermometer-deployment
    spec:
      containers:
      - image: edgehub/deviceshifu-http-http:v0.0.1
        name: deviceshifu-http
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: edgedevice-config
          mountPath: "/etc/edgedevice/config"
          readOnly: true
        env:
        - name: EDGEDEVICE_NAME
          value: "edgedevice-thermometer"
        - name: EDGEDEVICE_NAMESPACE
          value: "devices"
      volumes:
      - name: edgedevice-config
        configMap:
          name: thermometer-configmap-0.0.1
      serviceAccountName: edgedevice-mockdevice-sa
```
##### 温度计服务示例
```
apiVersion: v1
kind: Service
metadata:
  labels:
    app: edgedevice-thermometer-deployment
  name: edgedevice-thermometer
  namespace: deviceshifu
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 8080
  selector:
    app: edgedevice-thermometer-deployment
  type: LoadBalancer
```

### ***edgeMap*** 设计

[设计图]

#### ***edgeMap*** 数据结构

***edgeMap*** 是通过 adjacency list 来实现。 下面是 adjacency list 中的定义：
```edgeVertex```: 一个 ***edgeNode*** 或者 ***edgeDevice***.

```edgeLink```： 两个 ```edgeVertex``` 中的连接，比如 Ethernet 或 USB.

#### ***edgeVertex*** 数据结构

***edgeVertex*** 是linked list的一个节点。 下面是节点的内容：

```vertexType```: ***edgeNode*** 或 ***edgeDevice***

```neighborVertex```: 当前vertex的下一个邻居。

```neighborLinkType```: 当前 ***edgeVertex*** 和 ```neighborVertex``` 的连接。
