# shifuController design

***shifuController***'s main responsibility is to manage the lifecycle of ***deviceShifu***.
***shifuController*** react to ***edgeDevice*** and ***edgeNode*** events sent by ***apiServer*** by creating/deleting the corresponding ***deviceShifu*** instances.

## Design goals and non-goals

### Design goals

#### low resource consumption

Since ***shifuController*** can run either on the cloud or on the edge, ***shifuController*** needs to consume as less memory as possible.
Preferably the memory print for ***shifuController*** is under 100MB. 

#### highly available

Being the control plane of ***shifu***, ***shifuController*** needs to be highly available. This is achieved by using Kubernetes deployment.

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
1. Schedule: Determine where to place the pending ***deviceShifu***. 
   1. If the ***edgeDevice*** is connected to a specific ***edgeNode***, then the ***deviceShifu*** will be scheduled on that ***edgeNode***.
   2. If the ***edgeDevice*** is connected to the cluster network, then the ***deviceShifu*** will be scheduled based on below priorities (subject to change):
      1. Location proximity: If location info is available, then the ***deviceShifu*** will be placed on the ***edgeNode*** closest to the ***edgeDevice***.
      2. Resource availability: The ***deviceShifu*** will be placed on the ***edgeNode*** with highest free memory.
2. Add: Add the ***edgeDevice*** to ***edgeMap***.
3. Compose: Compile all computed info and compose the corresponding ***deviceShifu*** deployment object.
4. Create: Create the ***deviceShifu*** deployment by submitting the request to ***apiServer***.

#### 2. delete ***edgeDevice***

Whenever ***shifuController*** receives an ***edgeDevice*** disconnect event, ***shifuController*** will:
1. Remove: Remove the ***edgeDevice*** from ***edgeMap***.
2. Delete: Delete the ***deviceShifu*** deployment by submitting the request to ***apiServer***.

### on ***edgeNode*** events

#### 1. create ***edgeNode***

Whenever ***shifuController*** receives an ***edgeNode*** create event, ***shifuController*** will:
1. Add: Add the ***edgeNode*** to ***edgeMap***.

#### 2. delete ***edgeNode***

Whenever ***shifuController*** receives an ***edgeNode*** delete event, ***shifuController*** will:
1. Delete: Delete the ***edgeNode*** from ***edgeMap***.

### During normal operation

1. Collect: ***shifuController*** will periodically collect resource information on ***edgeNodes***. This information will be used by 
   1. ***edgeMap*** for visualization purposes.
   2. reschedule purposes, see below.
2. Reschedule: ***shifuController*** will reschedule ***deviceShifu*** on the following events:
   1. ***shifuController*** finds a better place for a given ***deviceShifu*** according to the scheduling algorithm stated above.
   2. ***edgeNode*** becomes unavailable. Whenever an ***edgeNode*** becomes unavailable, ***shifuController*** will try to reschedule all the ***deviceShifu*** instances on that ***edgeNode*** to other ***edgeNodes*** in a best-effort manner.


[edgeNode1] --> [camera1] --> [camera2]
[camera1] --> [edgeNode1]
[camera2] --> [edgeNode1]
