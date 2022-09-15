# 【WIP】shifuController design
- [【WIP】shifuController design](#wipshifucontroller-design)
  - [Design goals and non-goals](#design-goals-and-non-goals)
    - [Design goals](#design-goals)
      - [low resource consumption](#low-resource-consumption)
      - [highly available](#highly-available)
      - [stateless](#stateless)
      - [minimum privileges](#minimum-privileges)
    - [Design non-goals](#design-non-goals)
      - [***shifud*** management](#shifud-management)
      - [***edgeDevice*** management](#edgedevice-management)
  - [Design overview](#design-overview)
    - [On ***edgeDevice*** events](#on-edgedevice-events)
      - [1. create ***edgeDevice***](#1-create-edgedevice)
      - [2. delete ***edgeDevice***](#2-delete-edgedevice)
    - [on ***edgeNode*** events](#on-edgenode-events)
      - [1. create ***edgeNode***](#1-create-edgenode)
      - [2. delete ***edgeNode***](#2-delete-edgenode)
    - [During normal operation](#during-normal-operation)
    - [***deviceShifu*** object](#deviceshifu-object)
      - [Basic thermometer sample](#basic-thermometer-sample)
        - [Basic thermometer sample deployment](#basic-thermometer-sample-deployment)
        - [Basic thermometer sample service](#basic-thermometer-sample-service)
    - [***edgeMap*** design](#edgemap-design)
      - [***edgeMap*** data structure](#edgemap-data-structure)
      - [***edgeVertex*** data structure](#edgevertex-data-structure)

***shifuController***'s main responsibility is to manage the lifecycle of ***deviceShifu***.
***shifuController*** reacts to ***edgeDevice*** and ***edgeNode*** events sent by ***apiServer*** by creating/deleting the corresponding ***deviceShifu*** instances.

## Design goals and non-goals

### Design goals

#### low resource consumption

Since ***shifuController*** can run either on the cloud or on the edge, ***shifuController*** needs to consume as less memory as possible.
Preferably the memory print for ***shifuController*** is under 100MB. 

#### highly available

Being the control plane of ***shifu***, ***shifuController*** needs to be highly available. This is achieved by using [Kubernetes deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) and [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/).

#### stateless

***shifuController*** offloads its persistent store to etcd (or any Kubernetes backend persistent store) and become truly stateless.

#### minimum privileges

***shifuController*** should always require minimum privileges to reduce potential interference with other micro-services and maintain a high security standard.

### Design non-goals

#### ***shifud*** management

***shifud*** is a [Kubernetes daemonSet](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/), which is managed by Kubernetes.

#### ***edgeDevice*** management

***shifuController*** only manages ***deviceShifu***, not ***edgeDevice*** nor ***edgeNode***.

## Design overview

***shifuController*** as a [Kubernetes controller](https://kubernetes.io/docs/concepts/architecture/controller/), manages the whole lifecycle of ***deviceShifu*** deployments through [Kuberenetes CRD(Custom Resource Definition)](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/).
***shifuController*** will internally cache the current network topology by maintaining an [adjacency list](https://en.wikipedia.org/wiki/Adjacency_list#:~:text=In%20graph%20theory%20and%20computer,particular%20vertex%20in%20the%20graph.). This adjacency list is called ***edgeMap***. In the future, ***edgeMap*** will be used to render topology visualizations for the developer/operator. 

### On ***edgeDevice*** events
#### 1. create ***edgeDevice***

Whenever ***shifuController*** receives an ***edgeDevice*** connect event, ***shifuController*** will: 

1.1 Schedule: Determine where to place the pending ***deviceShifu***.    
   1.1.1 If the ***edgeDevice*** is connected to a specific ***edgeNode***, then the ***deviceShifu*** will be scheduled on that ***edgeNode***.  
   1.1.2 If the ***edgeDevice*** is connected to the cluster network, then the ***deviceShifu*** will be scheduled based on below priorities (subject to change):  
      a. Location proximity: If location info is available, then the ***deviceShifu*** will be placed on the ***edgeNode*** closest to the ***edgeDevice***.  
      b. Resource availability: The ***deviceShifu*** will be placed on the ***edgeNode*** with highest available memory.

1.2 Compose: Compile all computed info and compose the corresponding ***deviceShifu*** deployment object.

1.3 Create:   
    1.3.1 Create the ***deviceShifu*** deployment by submitting the request to ***apiServer***.
    1.3.2 Expose the newly created ***deviceShifu*** as a Kubernetes Service.

1.4 Add: Add the ***edgeDevice*** to ***edgeMap***.

#### 2. delete ***edgeDevice***

Whenever ***shifuController*** receives an ***edgeDevice*** disconnect event, ***shifuController*** will:  

2.1 Remove: Remove the ***edgeDevice*** from ***edgeMap***.

2.2 Delete:   
   2.2.1 Delete the ***deviceShifu*** deployment by submitting the request to ***apiServer***.
   2.2.2 Delete the ***deviceShifu***'s corresponding Kubernetes service.

### on ***edgeNode*** events

#### 1. create ***edgeNode***

Whenever ***shifuController*** receives an ***edgeNode*** create event, ***shifuController*** will:

1.1 Add: Add the ***edgeNode*** to ***edgeMap***.

#### 2. delete ***edgeNode***

Whenever ***shifuController*** receives an ***edgeNode*** delete event, ***shifuController*** will:

2.1 Delete: Delete the ***edgeNode*** from ***edgeMap***.

### During normal operation

1. Collect: ***shifuController*** will periodically collect resource information on ***edgeNodes***. This information will be used by 
   1. ***edgeMap*** for visualization purposes.
   2. reschedule purposes, see below.
2. Reschedule: ***shifuController*** will reschedule ***deviceShifu*** on following events:
   1. ***shifuController*** finds a better place for a given ***deviceShifu*** according to the scheduling algorithm stated above.
   2. ***edgeNode*** becomes unavailable. Whenever an ***edgeNode*** becomes unavailable, ***shifuController*** will try to reschedule all the ***deviceShifu*** instances on that ***edgeNode*** to other ***edgeNodes*** in a best-effort manner.

### ***deviceShifu*** object

***deviceShifu*** object is a bundle of a Kubernetes deployment and a Kubernetes service for an ***edgeDevice*** managed by ***shifuController***. Users should **NOT** manage ***deviceShifu*** by themselves.

#### Basic thermometer sample

##### Basic thermometer sample deployment
Below is a sample ***deviceShifu*** deployment yaml for a basic thermometer:

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
##### Basic thermometer sample service
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

### ***edgeMap*** design

[Architectural diagram here]

#### ***edgeMap*** data structure

***edgeMap*** is implemented as an adjacency list. Below are the definitions within the adjacency list:

```edgeVertex```: an ***edgeNode*** or an ***edgeDevice***

```edgeLink```： type of connection between two ```edgeVertex```. For example, Ethernet or USB.

#### ***edgeVertex*** data structure

***edgeVertex*** is a node in a linked list. Below are the fields within the the node:

```vertexType```: ***edgeNode*** or ***edgeDevice***

```neighborVertex```: the immediate neighbor of the current vertex

```neighborLinkType```: the connection type between the current ***edgeVertex*** and ```neighborVertex```
