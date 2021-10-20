<div align="center">

![Edgenesis logo](img/logo.png)

</div>

[![Build Status](https://dev.azure.com/Edgenesis/shifu/_apis/build/status/Edgenesis.shifu?branchName=main)](https://dev.azure.com/Edgenesis/shifu/_build/latest?definitionId=1&branchName=main)

<div align="right">

## [中文文档](README-zh.md)

</div>

- [Shifu](#shifu)
  - [What is Shifu?](#what-is-shifu)
  - [Why use Shifu?](#why-use-shifu)
  - [How to use Shifu?](#how-to-use-shifu)
  - [Quick Start Guide with Demo](#quick-start-guide-with-demo)
- [Our Roadmap](#our-roadmap)
  - [Current state of Shifu OS](#current-state-of-shifu-os)
  - [Protocols](#protocols)
    - [Supported](#supported)
  - [Features](#features)
    - [Supported](#supported-1)
    - [Not yet supported](#not-yet-supported)
  - [Milestone](#milestone)
- [Shifu's vision](#shifus-vision)
  - [Make developers and operators happy again](#make-developers-and-operators-happy-again)
  - [Software Defined World (SDW)](#software-defined-world-sdw)
- [Community](#community)
  - [Contact](#contact)

# Shifu

## What is Shifu?

Shifu is a framework designed to abstract out the complexity of interacting with IoT devices. Shifu aims to achieve TRUE plug'n'play IoT device management.

## Why use Shifu?

Shifu let you manage and control your IoT devices extremely easily. Whenever you connect your device, Shifu will recognize it and spawn an augmented digital twin called ***deviceShifu*** for it. deviceShifu provides you with a high-level abstraction to interact with your device. By implementing the interface of the deviceShifu, your IoT device can achieve everything its designed for, and much more! For example, your device's state can be rolled back with a single line of command. (If physically permitted, of course.) Shifu is able to abstract ***deviceShifu*** horizontally (grouping, batch execute) and vertically (layers, allow high level command to be executed. e.g.: `factory start`). Furthermore, Shifu will secure your entire IoT system from ground up. A simulation feature which allows developer to simulate a scenario before actually running will be available later.

## How to use Shifu?

Currently, Shifu runs on [Kubernetes](k8s.io). We will provide more deployment methods including standalone deployment in the future.

## Quick Start Guide with Demo
We prepared a [demo](docs/guide/quick-start-demo.md) for developers to intuitively show how our `shifu` is able to create and manage digital twins of any physical devices in real world. 

# Our Roadmap

## Current state of Shifu OS
We will continuously add in features as we develop Shifu OS
## Protocols
### Supported
- HTTP
- Driver implementation w/ command line execution
- ... More on the way
## Features
### Supported
- Telemetry collection
- Command proxy to device
- Integration with Kubernetes with CRD
- Basic Shifu Controller
### Not yet supported
- Declarative API
- Advanced Shifu Controller
- shifud
- Abstraction:
  - Horizontal
  - Vertical
- Simulation
- Security features:
  - Firewall
  - uTLS

## Milestone
<style type="text/css">
.tg  {border-collapse:collapse;border-spacing:0;}
.tg td{border-color:black;border-style:solid;border-width:1px;font-family:Arial, sans-serif;font-size:14px;
  overflow:hidden;padding:10px 5px;word-break:normal;}
.tg th{border-color:black;border-style:solid;border-width:1px;font-family:Arial, sans-serif;font-size:14px;
  font-weight:normal;overflow:hidden;padding:10px 5px;word-break:normal;}
.tg .tg-0pky{border-color:inherit;text-align:left;vertical-align:top}
</style>
<table class="tg">
<thead>
  <tr>
    <th class="tg-0pky">By</th>
    <th class="tg-0pky">Protocol</th>
    <th class="tg-0pky">Features</th>
  </tr>
</thead>
<tbody>
  <tr>
    <td class="tg-0pky">Q4 2021</td>
    <td class="tg-0pky">HTTP<br>Driver w/ command line<br></td>
    <td class="tg-0pky">Telemetry<br>Command proxy<br>CRD integration<br>Basic Controller</td>
  </tr>
  <tr>
    <td class="tg-0pky">Q1 2022</td>
    <td class="tg-0pky">At least:<br>MQTT<br>Modbus<br>ONVIF<br>国标GB28181<br>USB</td>
    <td class="tg-0pky">Declarative API<br>Advanced Controller<br>shifud</td>
  </tr>
  <tr>
    <td class="tg-0pky">Q2 2022</td>
    <td class="tg-0pky">At least:<br>OPC UA<br>Serial<br>Zigbee<br>LoRa<br>PROFINET</td>
    <td class="tg-0pky">Abstraction</td>
  </tr>
  <tr>
    <td class="tg-0pky">Q3 2022</td>
    <td class="tg-0pky">TBD</td>
    <td class="tg-0pky">Security Features</td>
  </tr>
  <tr>
    <td class="tg-0pky">Q3 2023</td>
    <td class="tg-0pky">TBD</td>
    <td class="tg-0pky">Simulation</td>
  </tr>
</tbody>
</table>

# Shifu's vision

## Make developers and operators happy again

Developers and operators should 100% focus on using their creativity to make huge impacts, not fixing infrastructures and re-inventing the wheels day in and day out. As a developer and an operator, Shifu knows your pain!!! Shifu wants to take out all the underlying mess, and make us developers and operators happy again!

## Software Defined World (SDW)

If every IoT device has a Shifu with it, we can leverage software to manage our surroundings, and make the whole world software defined. In a SDW, everything is intelligent, and the world around you will automatically change itself to serve your needs. After all, all the technology we built are designed to serve human beings. 

# Community 

## Contact

Feel free to open a GitHub issue or contact us in the following ways:
- Send an email to info@edgenesis.com
