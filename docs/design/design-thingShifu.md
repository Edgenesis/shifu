# ***thingShifu*** Design Document
- [***thingShifu*** Design Document](#thingshifu-design-document)
  - [Design Goals and Non-goals](#design-goals-and-non-goals)
    - [Design Goals](#design-goals)
      - [Digital Twin](#digital-twin)
      - [Easy to Deploy and Use](#easy-to-deploy-and-use)
      - [Easy to Extend](#easy-to-extend)
    - [Design Non-goals](#design-non-goals)
      - [Absolute Accurate Representation](#absolute-accurate-representation)
      - [Automatically Fix Hardware Issues](#automatically-fix-hardware-issues)
  - [What is ***thingShifu***?](#what-is-thingshifu)
  - [What is ***deviceShifu***?](#what-is-deviceshifu)
  - [What is "thing"?](#what-is-thing)
  - [***thingShifu*** Components](#thingshifu-components)
    - [***thingShifu*** Core](#thingshifu-core)
      - [1. ***bootstrapper***](#1-bootstrapper)
      - [2. ***core preparer***](#2-core-preparer)
      - [3. ***inbound instruction processor***](#3-inbound-instruction-processor)
      - [4. ***thing telemetry collector***](#4-thing-telemetry-collector)
      - [5. ***thingShifu client message sender***](#5-thingshifu-client-message-sender)
  - [Operation Modes of thingShifu](#operation-modes-of-thingshifu)
    - [Swarm Mode](#swarm-mode)
  - [States of thingShifu](#states-of-thingshifu)
    - [1. ***creation state***](#1-creation-state)
    - [2. ***preparation state***](#2-preparation-state)
    - [3. ***running state***](#3-running-state)
    - [4. ***termination state***](#4-termination-state)
  - [Structure](#structure)
  - [Hierarchy of ***thing*** and ***thingShifu***](#hierarchy-of-thing-and-thingshifu)
    - [Sample YAML configuration](#sample-yaml-configuration)
  - [#Sample YAML of the Factory thing](#sample-yaml-of-the-factory-thing)
  - [#Sample YAML of the Streamline Engine 1 thing](#sample-yaml-of-the-streamline-engine-1-thing)
  - [#Sample YAML of the AR 1-1 thing](#sample-yaml-of-the-ar-1-1-thing)
  - [#Sample YAML of the Factory thingShifu](#sample-yaml-of-the-factory-thingshifu)
  - [#if shifu_mode is not specified, standalone will be used](#if-shifu_mode-is-not-specified-standalone-will-be-used)
  - [#Sample YAML of the AR 1-1 thingShifu](#sample-yaml-of-the-ar-1-1-thingshifu)
  - [#Sample YAML of the temperature sensor thingShifu in swarm mode](#sample-yaml-of-the-temperature-sensor-thingshifu-in-swarm-mode)
    - [Hierarchy of Instructions](#hierarchy-of-instructions)
    - [Hierarchy of Telemetries](#hierarchy-of-telemetries)
  - [Grouping of ***thingShifu***](#grouping-of-thingshifu)
  - [Sample YAML of the Group A](#sample-yaml-of-the-group-a)
  - [Limitations and External Components Can Be Added](#limitations-and-external-components-can-be-added)
    - [Message Queue](#message-queue)
    - [Database](#database)


## Design Goals and Non-goals
### Design Goals
#### Digital Twin
***thingShifu*** is a digital twin of a ***thing***, a digital representation of evey man-made thing that has an entity in the real world.
#### Easy to Deploy and Use
By simply writting a configuration, the user can utilize ***thingShifu*** to easily control the ***thing*** in the most software way without worrying about hardware issues; being a part of the ***shifu framework***, multiple ***thingShifu*** can be easily grouped to accomplish more complex goals with very simple instruction from the user.
#### Easy to Extend
***thingShifu*** has no limitation. Anyone can add new features to it easily.

### Design Non-goals
#### Absolute Accurate Representation
***thingShifu*** serves as a digital representation of a ***thing***, but becaue human's understanding of the real world is still not 100%, our ***thingShifu*** can not 100% represent a ***thing*** either.
#### Automatically Fix Hardware Issues
If there are some issues about the ***thing*** real world representation like bad circuit design or malfunction chip, ***thingShifu*** as a digital representation software cannot help.

## What is ***thingShifu***?
***thingShifu*** is an augmented digital twin of a ***thing***. It is the component that is closest to the end user in the ***shifu*** framework, serving as a complete representation of the ***thing*** - it will let developers and operators to use simple APIs to control the ***thing***, and will let the operating personnel easily know the status of it.

## What is ***deviceShifu***?
***deviceShifu*** is a subset of ***thingShifu*** and is used for mechanical devices. It contains all features of ***thingShifu***.

## What is "thing"?
A ***thing*** basically can be anything that has an entity created by human in the real world - it can be as small as a circuit board, a microchip, a a camera, a phone, a robot, a car, and can be as big as a building, a manufacturing site, a street, and a city.

The ultimate goal of ***thingShifu*** is to become the "shifu" (a.k.a. teacher or instructor) to all man-made devices to make them smarter on serving the needs of human beings. For the current stage, we are focusing on making ***thingShifu*** to represent one or more IoT devices.

## ***thingShifu*** Components
  
### ***thingShifu*** Core

#### 1. ***bootstrapper***
entry of ***thingShifu*** creation, reading the spawn request, load the configuration from the kubernetes apiServer, and establish a connection to the ***thing***.

input: spawn request from shifuController
output: ready to prepare thingShifu core

#### 2. ***core preparer***
parses and processes the configuration file and makes the thingShifu ready to process inbound instructions and collect telemetries, also publish the available API to client via client message sender.

input: configuration of the ***thing***: the info useful for starting the ***thing***, determined by the user, for example, it can include:
   1. available instructions like ```[device ip]/deviceMovement <x-y coordinations>```
   2. desired telemetries to collect (telemetry name and optional collecting interval), such as ```device_on, device_off, device_health```
   3. mapping of human-readable instruction from the user to the format accepted by the ***thing***


output (cached in memory): 
   1. available instructions
   2. teletemries to collect

#### 3. ***inbound instruction processor***
translates the instruction received from the user

input: instructions (from user at real-time)

output: instructions that can be directly sent to the ***thing***.

#### 4. ***thing telemetry collector***
keeps collecting the telemetry (status/healthness/updates) from the ***thing***
- simplest way: sending an instruction to ***thing***, getting the response
- other ways: subscrpition of updates in separate channel, REST API, ...
- it must have at least a "ping" method to make sure the ***thing*** is connected

#### 5. ***thingShifu client message sender***
- ***thingShifu*** expects to have a client (UI) to let the operating personnel to easily interact with thingShifu
- attached with ***thingShifu*** Core, serving as a proxy of instruction/telemetry updates
- a separate thread serves only to send message to client (UI)

## Operation Modes of thingShifu
**standalone mode**: the ***thingShifu*** is managing one single ***thing***. For example, a temperature sensor.

**swarm mode**: the ***thingShifu*** is managing multiple same-type ***things***. For example, a group of temperature sensors.

By default, the ***thingShifu*** is in **standalone mode**.

 
### Swarm Mode
 A typical **swarm mode** will be like this:

[![thingShifu factory example swarm mode](/img/thingShifu/shifu-thingShifu-example-factory-swarm.svg)](/img/thingShifu/shifu-thingShifu-example-factory-swarm.svg)


## States of thingShifu
### 1. ***creation state***
***shifuController*** is responsible for creating the ***thingShifu***, triggering ***bootstrapper***.

Once bootstrapper reports ready, go to next step

### 2. ***preparation state***
Upon bootstrapper reports ready, ***core preparer*** is triggered. Then it reads the config file and loads everything into memory.

Once the preparation is complete, go to next step.

### 3. ***running state***
Upon the preparation is complete, ***thingShifu core*** will start:
 - pinging the ***thing*** periodically,
 - collecting the telemetries periodically,
 - keep the inbound instruction port always open.

### 4. ***termination state***
***shifuController*** will tell the ***thingShifu*** to terminate, and it will free the memory used, and report a "terminated" message.

## Structure

[![thingShifu basic structure](/img/thingShifu/shifu-thingShifu-basic-structure.svg)](/img/thingShifu/shifu-thingShifu-basic-structure.svg)

## Hierarchy of ***thing*** and ***thingShifu***
A working ***thing*** is highly possible to have multiple layers of lower-level ***things***. Therefore, ***thingShifu*** allows the instruction and telemetry hierarchy of ***things***. Each ***thingShifu*** will run the instruction based on the arrival time and priority of the instruction.

Let's say we have a factory with 3 streamline and each streamline has 2 types of robots installed: assemble robot (AR) and transport robot (TR).

In this structure, all of the entities are ***things***. Let's say the user has specified the structure of ***things*** like this:

[![thingShifu factory example](/img/thingShifu/shifu-thingShifu-example-factory-structure.svg)](/img/thingShifu/shifu-thingShifu-example-factory-structure.svg)

and the user has specified the ***thingShifu*** structure like this:

[![thingShifu factory example with thingShifu](/img/thingShifu/shifu-thingShifu-example-factory-thingShifu.svg)](/img/thingShifu/shifu-thingShifu-example-factory-thingShifue.svg)

Based on the structure defined above, the hierarchy is clear to us: ***Factory thingShifu*** is at the top, and it has 2 lower layer of ***thingShifu***: ***Streamline Engine 1 thingShifu*** and ***Streamline Engine 2 thingShifu***. Each Streamline Engine thingShifu then has 4 lower ***thingShifu***, which denotes the robots.

Note that the structure setup depends on what the user likes to have, so it is also possible to have a totally different structure.

### Sample YAML configuration

***thingShifu*** structure expects to get two types of configurations:
- YAML configuration of each ***thingShifu***
- YAML configuration of each ***thing***

User-specified configuration of the hierarchy will be in the format like this:
#####thing configuration
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
#####thingShifu configuration

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

### Hierarchy of Instructions
***thingShifu*** allows the user to easily execute the common instructions among all layers at once via one single instruction.

For example, we can have the instructions of **start**, **halt** and **stop** (implemented by user).

With the structure above, the user just needs to send such instruction to the ***Factory thingShifu***, and then all the entities that exists in the ***Factory thingShifu*** hierarchy will receive such instruction, and execute the instruction according to its own logic. 

For example, if we send a **start** instruction to ***Factory thingShifu***, all the ***things*** in the hierarchy of the ***factory thing*** will execute their own **start** logic: the factory itself will set to "started" state,  the streamline engines 1 and 2 will start moving, the assemble robots in those two streamlines will set themselves to a ready state, and the transport robots in those two streamlines will move to their starting location waiting to load product parts.

### Hierarchy of Telemetries
What telemetries the lower layer ***thingShifu*** will report to its immediate upper layer are depending on user's configuration. 

For example, if we want the ***Factory thingShifu*** to tell us the current state of each robot, we want to config the robot thingShifu to report to streamline thingShifu, and config the streamline thingShifu to report to factory thingShifu. Thus the telemetry of robot state will go through the report path like this:

***AR/TR thingShifu*** -> ***Streamline thingShifu*** -> ***Factory thingShifu***.


## Grouping of ***thingShifu***
Grouping can be thought as a "horizontal hierarchy" but it is more flexible. In real-time, the user has the freedom to group several thingShifu so an instruction can be sent to all of them at once, and telemetries can be collected as a group.

The difference between grouping and hierarchy is, the grouping does not have a "highest" or "top" ***thingShifu*** and all grouped ***thingShifu*** are treated as in the same layer.

The difference between grouping and swarm mode is, the grouping is a group of ***thingShifu*** and can be disbanded, but swarm mode is just one ***thingShifu*** managing a group of ***things*** and cannot be divided into multiple ***thingShifu***.

An example of grouping is like this:
[![thingShifu factory example grouping](/img/thingShifu/shifu-thingShifu-example-factory-grouping.svg)](/img/thingShifu/shifu-thingShifu-example-factory-grouping.svg)

In this structure, we have a **Group A** containing TR 1-1, TR 1-2, AR 2-2 and TR 2-1. Therefore, if we send a **halt** instruction to Group A, these 4 ***thingShifu*** will execute this instruction.

The configuration of this ***grouping*** is like this:
```` 
Sample YAML of the Group A
---
group: "Group A"
id: 1
thingShifu_in_group:["TR 1-1", "TR 1-2", "AR 2-2", "TR 2-1"]
...
````

User does not need to provide instructions to the group, as the common instructions of the grouped ***thingShifu*** will be extracted.

##Race Condition
***thingShifu*** will have a write lock all resources when executing one instruction so that other instructions can't change.

When multiple instructions come at the same time, the ***thingShifu*** will first try to run the instructions according to the priority; if they are all of the same priority, ***thingShifu*** will randomly select one to get the lock and execute.

###Priorities of Instructions
User can define the highest priority over all other instructions (priority -1), the instruction with priority -1 can break any currently executing instructions.

The instruction from user to the ***thingShifu*** directly always has the highest priority (priority 0).

The instruction on a group from user always has the second highest priority (priority 1). 

For instructions passed by the higher layer in the hierarchy, ***thingShifu*** will treat the instruction come from higher layer as having higher prioirty (priority from 2, defined by user).



## Limitations and External Components Can Be Added
### Message Queue
Under high volume of instructions and telemetries, user should consider adding a message queue to the hierarchy of ***thingShifu***.

### Database
***thingShifu*** does not expect large amount of data to be stored so it keeps everything in memory; but if needed, a standalone database can be added.

<!---
### Sample YAML configuration file

```` 
Sample YAML
---
device: PandaArm
commands:
   MoveTo:
   - driver_instruction: "absolute_move"
   - input1: coordinate_system
     type: string
   - input2: goal_pose
     type: nested
     sub_inputs: [x, y, z, a, b, c, elbow]
     sub_types: [int, int, int, int, int, int, int]
   - input3: max_velocity
     type: int
   - input4: max_acceleration
     type: int
   - input5: movement_type
     type: int
   Grip:
   - driver_instruction: "grip"
   - input1: grip_force
     type: int
   - input2: object_size
     type: int
...
````
is corresponding to
````
{
    "command":"absolute_move",
    "args": {
 	    "coordinate_system":"cartesian",
	    "goal_pose":{
		    "x":123,
	    	"y":456,
			"z":789,
			"a":10,
			"b":20,
			"c":30,
			"elbow": 0
 	    },
        "max_velocity":500,
        "max_acceleration":100,
        "movement_type":"standard"
    }
}


{
    "command":"grip",
    "args": {
		"grip_force": 100,
        "object_size": 30
 	}
}
````
-->
