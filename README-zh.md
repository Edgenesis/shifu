<div align="right">

中文 | [English](README.md) | [日本語](README-ja.md)

[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat&logo=github&color=2370ff&labelColor=454545)](http://makeapullrequest.com)
[![Go Report Card](https://goreportcard.com/badge/github.com/Edgenesis/shifu)](https://goreportcard.com/report/github.com/Edgenesis/shifu)
[![codecov](https://codecov.io/gh/Edgenesis/shifu/branch/main/graph/badge.svg?token=OX2UN22O3Z)](https://codecov.io/gh/Edgenesis/shifu)
[![Build Status](https://dev.azure.com/Edgenesis/shifu/_apis/build/status/shifu-build-muiltistage?branchName=main)](https://dev.azure.com/Edgenesis/shifu/_build/latest?definitionId=19&branchName=main)
[![golangci-lint](https://github.com/Edgenesis/shifu/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/Edgenesis/shifu/actions/workflows/golangci-lint.yml)

</div>

<div align="center">

<img width="300px" src="./img/shifu-logo.svg"></img>
<div align="center">

<h1 style="border-bottom: none">
<br>
    Kubernetes原生的物联网网关
    <br />
</h1>
Shifu是一个k8s原生的，生产级的，支持多协议设备接入的开放物联网网关。
</div>
</div>
<br/><br/>

<div align="center">
    <a href="https://discord.gg/2tbdBrkGHv"><img src="https://img.shields.io/badge/Discord-5865F2?style=for-the-badge&logo=discord&logoColor=white" height="25"></a>
    &nbsp;
    <a href="https://x.com/ShifuFramework"><img src="https://img.shields.io/twitter/follow/ShifuFramework" height="25"></a>
    &nbsp;
    <a href="https://www.linkedin.com/company/edgenesis/"><img src="https://img.shields.io/badge/LinkedIn-0077B5?style=for-the-badge&logo=linkedin&logoColor=white" height="25"></a>
    &nbsp;
    <a href="https://www.youtube.com/channel/UCsaj5f4IKKKn9OMiTVYCvRA"><img src="https://img.shields.io/youtube/channel/subscribers/UCsaj5f4IKKKn9OMiTVYCvRA" height="25"></a>
    &nbsp;
    <a href="https://github.com/Edgenesis/shifu"><img src="https://img.shields.io/github/stars/Edgenesis/shifu" height="25"></a>
</div>

## ✨招聘✨
我们正在招聘！Shifu大家庭举双手欢迎爱折腾的你！！！

[👉🙋‍♀️**职位点这里**👈🙋‍♂️](https://4g1tj81q9o.jobs.fbmms.cn/page/PSVAGacDW6xEEcT5qbbfRL0FR3)

## Shifu的价值: 让大家开发应用，而不是基础设施
<div align="center">
<img width="900px" src="./img/iot-stack-with-shifu.svg"></img>
</div>

## CNCF现场直播和演示

[![Cloud Native Live](https://img.youtube.com/vi/qMrdM1QcLMk/maxresdefault.jpg)](https://www.youtube.com/watch?v=qMrdM1QcLMk)

## 特点
**Kubernetes原生** — 应用开发的同时进行设备管理，无需再构建额外的运维基础设施

**开放平台**— 避免供应商锁定，你可以轻松地将Shifu部署在公有云、私有云或混合云上。Shifu将Kubernetes带入到物联网边缘计算场景中，助力实现物联网应用程序的可扩展性和高可用性。

**多协议设备接入** — HTTP, MQTT, RTSP, Siemens S7, TCP socket, OPC UA...从公有协议到私有协议，Shifu的微服务架构让我们能够快速整合接入新的协议。

## 定义 
**shifu** - 一个把IoT设备接入Kubernetes集群的CRD。

**DeviceShifu** - 一个Kubernetes pod，同时也是Shifu的最小单元。DeviceShifu的主要组成部分是设备的驱动，代表一个IoT设备，也可以称之为“数字孪生”。


<div align="center">
<img width="900px" src="./img/shifu-architecture.png"></img>
</div>

## 如何用五行代码连接一个使用私有协议的摄像头
<div align="center">

<img width="900px" src="./img/five-lines-to-connect-to-a-camera.gif"></img>

<img width="900px" src="./img/star.gif"></img>
</div>

## 社区

欢迎加入Shifu社区，分享您的思考与想法，

您的意见对我们来说无比宝贵。 我们无比欢迎您的到来！

[![加入 Discord](https://img.shields.io/badge/Discord-Join-5865F2?style=flat&logo=discord&logoColor=white)](https://discord.gg/CkRwsJ7raw)
[![关注 X](https://img.shields.io/twitter/follow/ShifuFramework)](https://x.com/ShifuFramework)
[![Reddit](https://img.shields.io/badge/Reddit-post-orange)](https://www.reddit.com/r/Shifu/)
[![GitHub Discussions](https://img.shields.io/badge/GitHub%20Discussions-post-orange)](https://github.com/Edgenesis/shifu/discussions)
[![YouTube 订阅](https://img.shields.io/youtube/channel/subscribers/UCsaj5f4IKKKn9OMiTVYCvRA)](https://www.youtube.com/channel/UCsaj5f4IKKKn9OMiTVYCvRA)

## 开始上手
欢迎参考🗒️[Shifu技术文档](https://shifu.dev/)获取更详细的信息:
- 🔧[安装Shifu](https://shifu.dev/zh-Hans/docs/guides/install/install-shifu-dev)
- 🔌[设备连接](https://shifu.dev/zh-Hans/docs/guides/cases/)
- 👨‍💻[应用开发](https://shifu.dev/zh-Hans/docs/guides/application/)
- 🎮[KillerCoda Demo在线试玩](https://killercoda.com/shifu/shifu-demo)

## 贡献 
欢迎向我们[提交issue](https://github.com/Edgenesis/shifu/issues/new/choose) 或者[提交PR](https://github.com/Edgenesis/shifu/pulls)!

我们对[贡献者们](https://github.com/Edgenesis/shifu/graphs/contributors)心怀感激🥰.

## Shifu正式加入[CNCF全景图](https://landscape.cncf.io/)

<div align="center">
<img width="900px" src="./img/cncf-logo.png"></img>
</div>

<div align="center">
<img width="900px" src="./img/cncf.png"></img>
</div>

## Github Star数量

[![Stargazers over time](https://starchart.cc/Edgenesis/shifu.svg)](https://starchart.cc/Edgenesis/shifu)

## 许可证
该项目使用Apache2.0许可证。
