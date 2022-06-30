- [***Shifu*** high-level design](#shifu-high-level-design)
  - [Design principles](#design-principles)
    - [Human centric](#human-centric)
      - [1. Easy to deploy](#1-easy-to-deploy)
      - [2. Plug'n'Play](#2-plugnplay)
      - [3. Zero maintenance](#3-zero-maintenance)
      - [4. Easy to use SDKs](#4-easy-to-use-sdks)
  - [Design goals and non-goals](#design-goals-and-non-goals)
    - [Design goals](#design-goals)
      - [Simple](#simple)
      - [Highly available](#highly-available)
        - [1. Self healing](#1-self-healing)
        - [2. Stable](#2-stable)
      - [Cross-Platform](#cross-platform)
      - [Lightweight](#lightweight)
      - [Extensible and flexible](#extensible-and-flexible)
    - [Design non-goals](#design-non-goals)
      - [Platform other than Kubernetes](#platform-other-than-kubernetes)
      - [100% up-time](#100-up-time)
  - [Design overview](#design-overview)
    - [Components](#components)
      - [Physical components](#physical-components)
        - [1. ***edgeDevice***](#1-edgedevice)
        - [2. ***edgeNode***](#2-edgenode)
      - [Software components (Control Plane)](#software-components-control-plane)
        - [1. ***shifud***](#1-shifud)
        - [2. ***shifuController***](#2-shifucontroller)
      - [Software Components (Data Plane)](#software-components-data-plane)
        - [1. ***deviceShifu***](#1-deviceshifu)
    - [Architecture diagrams](#architecture-diagrams)
      - [The lifecycle of ***deviceShifu***](#the-lifecycle-of-deviceshifu)
        - [1. Device connect (user workload not shown in below figure)](#1-device-connect-user-workload-not-shown-in-below-figure)
        - [2. Device operating | TODO: formalize deviceShifu interface](#2-device-operating--todo-formalize-deviceshifu-interface)
        - [3. Device disconnect (user workload not shown in below figure)](#3-device-disconnect-user-workload-not-shown-in-below-figure)

# ***Shifu*** high-level design

## Design principles

### Human centric

***Shifu***'s foremost job is to make developers and operators happy. Thus, all ***Shifu***'s designs are human centric. Thus, there are a few design requirements to achieve the human centric goal:

#### 1. Easy to deploy

***Shifu*** can be always be summoned(deployed) by one single line of command.

#### 2. Plug'n'Play

***Shifu*** should be able to recognize and start providing basic functionalities to the newly added IoT device automagically. Once the developer finishes implementing ***Shifu***'s interface, all designed features of the device should be immediately available. The developer can further implement ***Shifu***'s interface to create customized features and open up endless possibilities.

#### 3. Zero maintenance

***Shifu*** aims to achieve zero maintenance. After all, ***Shifu*** should be able to take care of himself and make operators' lives easier!

#### 4. Easy to use SDKs

***Shifu*** will always provide developers with super easy to use SDKs, simple because *Shifu* wants to make developers happy!

## Design goals and non-goals

### Design goals

#### Simple

It is extremely important to have a simple architecture and logic to preserve a high standard of readability and maintainability.
An easy-to-read and easy-to-change code base will enable developers to improve ***Shifu*** or even create their very own versions of ***Shifu***.

#### Highly available

Being ***Shifu***, robustness is a must. A highly resilient ***Shifu*** is very important to operator's mental health. Who wants to deal with crash loops everyday?

##### 1. Self healing

***Shifu*** should always be able to heal itself upon unexpected events and drive itself towards goal state.

##### 2. Stable

***Shifu*** aims to achieve high availability and eliminate your operation costs.

#### Cross-Platform

***Shifu*** should be able to run on all major platforms, including but not limited to x86/64, ARM64, etc.

#### Lightweight

Being an IoT framework running on the edge, ***Shifu*** has to be lightweight. Hey Shifu, time to lose some weights!

#### Extensible and flexible

With so many heterogenous IoT devices around the world, ***Shifu*** needs to be extensible and flexible to accommodate different scenarios.

### Design non-goals

#### Platform other than Kubernetes

Our first priority is to ensure ***shifu*** runs on Kubernetes smoothly. We might provide standalone version in the future.

#### 100% up-time

***Shifu***'s design has built-in fault-tolerance. ***Shifu*** strives to achieve >99.9999% up-time, but doesn't aim for 100% up-time.

## Design overview

The current version of ***Shifu*** resembles a [Kubernetes Operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/). By leveraging the operator pattern, Shifu automatically gains all the benefits of Kubernetes offers.

### Components

#### Physical components

##### 1. ***edgeDevice*** 

***edgeDevice*** is a physical IoT device managed by ***Shifu***.

##### 2. ***edgeNode***

***edgeNode*** is a [Kubernetes node](https://kubernetes.io/docs/concepts/architecture/nodes/) that can connect to multiple ***edgeDevices***. By default, all worker nodes in the Kubernetes cluster are [tainted](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) with ***edgeNode***. For example User is able to configure nodes to not be an ***edgeNode***, therefore isolating their application Pods and ***deviceShifu*** Pods.

#### Software components (Control Plane)

##### 1. ***shifud***

***shifud*** is a daemonset runs on every ***edgeNode*** monitoring hardware changes, i.e., ***edgeDevice*** connect/disconnect events. ***shifud*** is managed by ***shifuController***.

##### 2. ***shifuController***

***shifuController*** is a [Kubernetes controller](https://kubernetes.io/docs/concepts/architecture/controller/) listens to hardware events sent by ***shifud*** and make corresponding actions to manage the lifecycle of ***deviceShifu***.

#### Software Components (Data Plane)
##### 1. ***deviceShifu***

***deviceShifu*** is an structural [digital twin](https://en.wikipedia.org/wiki/Digital_twin) of the ***edgeDevice***. We call it **structural** because it's not only a virtual presentation of the ***edgeDevice*** but also it's capable of driving the corresponding ***edgeDevice*** towards its goal state. For example, if you want your robot to move a box but it's currently busy on something else, ***deviceShifu*** will cache your instructions and tell your robot to move the box whenever it's available. For the basic part, ***deviceShifu*** provides some general functionalities such as ***edgeDevice*** health monitoring, state caching, etc. By implementing the interface of ***deviceShifu***, your ***edgeDevice*** can achieve everything its designed for, and much more!
***deviceShifu*** has two operation modes: 
1. ***standalone mode***: ***standalone mode*** is designed to manage a single complex ***edgeDevice*** like robotic arm to provide high quality 1-to-1 management for the ***edgeDevice***.
2. ***swarm mode***: ***swarm mode*** is designed to manage massive simple ***edgeDevices*** of the same kind like temperature sensors to provide efficient 1-to-N management for the ***edgeDevices***.

### Architecture diagrams

#### The lifecycle of ***deviceShifu***

##### 1. Device connect (user workload not shown in below figure)

1.1 connect: an ***edgeDevice*** physically connects to an ***edgeNode***.
1.2 device connect: ***shifud*** detects the newly connected device, and sends the event to ***shifuController***.
1.3 create: ***shifuController*** creates a ***deviceShifu*** in ***standalone/swarm mode*** for the ***edgeDevice***.
1.4 manage: ***deviceShifu*** starts to manage the newly connected ***edgeDevice***.

Upon an ***edgeDevice*** connects to the ***edgeNode***, ***Shifu*** will create a ***deviceShifu***, an augmented digital twin of the ***edgeDevice*** to manage it.

[![shifu-device connect](/img/shifu-device-connect.svg)](/img/shifu-device-connect.svg)

##### 2. Device operating | TODO: formalize deviceShifu interface

During normal operation, ***shifud*** and ***shifuController*** don't do much. User workloads interacts ***deviceShifu*** directly. For example, you can call ***deviceShifu***'s API to retrieve device metadata, device health, etc. Since it's two-way communication, once you implement your ***edgeDevice***'s specific methods in ***deviceShifu***'s API, can also call ***deviceShifu***'s API to manage your ***edgeDevice***. For example, you can setup your video camera's video stream through a few lines of code.

[![shifu-device operating](/img/shifu-device-operating.svg)](/img/shifu-device-operating.svg)

##### 3. Device disconnect (user workload not shown in below figure)

3.1 disconnect: an ***edgeDevice*** physically disconnects to an ***edgeNode***.
3.2 device disconnect: ***shifud*** detects the newly disconnected device, and sends the event to ***shifuController***.
3.3 delete: ***shifuController*** deletes the ***deviceShifu*** for the ***edgeDevice***. The delete process might take longer due to cleanup.

[![shifu-device disconnect](/img/shifu-device-disconnect.svg)](/img/shifu-device-disconnect.svg)
