<div align="right">

中文 | [English](README.md)

[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat&logo=github&color=2370ff&labelColor=454545)](http://makeapullrequest.com)
[![Go Report Card](https://goreportcard.com/badge/github.com/Edgenesis/shifu)](https://goreportcard.com/report/github.com/Edgenesis/shifu)
[![codecov](https://codecov.io/gh/Edgenesis/shifu/branch/main/graph/badge.svg?token=OX2UN22O3Z)](https://codecov.io/gh/Edgenesis/shifu)
[![Build Status](https://dev.azure.com/Edgenesis/shifu/_apis/build/status/shifu-build-muiltistage?branchName=main)](https://dev.azure.com/Edgenesis/shifu/_build/latest?definitionId=19&branchName=main)
[![golangci-lint](https://github.com/Edgenesis/shifu/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/Edgenesis/shifu/actions/workflows/golangci-lint.yml)

</div>

# ***Shifu***

<div align="center">

<img width="200px" src="./img/shifu-logo.svg"></img>

</div>

***Shifu***是下一代开源的产业物联网开发框架，让开发工业场景就像开发Web应用一样简单。***Shifu***的高度可扩展性使其能够访问几乎所有的设备和协议。一旦集成了一个设备，***Shifu***就会以微服务的形式在系统中生成一个设备的数字孪生，并以API的形式开放设备数据和功能。这样一来，传统的物联网应用开发就被简化为简单的Web开发，从而大大提高了物联网应用开发的效率、质量和复用性。


|特点<div style="width: 80pt">|  |
|---|---|
|⚡极速|大到使用私有驱动的工程机械，小到使用公有协议的温湿度计，Shifu的超高兼容性能让你轻松应对各种异构设备。|
|🧩模块化|所有接入Shifu的设备及应用都会被封装成一个个拼图式模块，根据场景内的设备不同, 按需加载即可。|
|👨‍💻高效|接入设备后，Shifu会自动把设备的能力抽象成API，让你的应用和硬件设备彻底解耦，让低效的物联网应用开发变得像面向对象编程一样高效。|
|🚀稳定|Shifu已通过在航天场景验证，提供99.9999%的可靠性, 让你远离宕机烦恼。|
|🛡️安全|联合国前云原生安全团队操刀，无论是数据加密还是网络安全，Shifu均可无缝集成。|
|🌐全球化|得益于Kubernetes原生架构, Shifu可以无缝接入强大的云原生软件生态，让全球的开发者帮你解决后顾之忧。|


## 使用

### 安装

在已有Kubernetes集群的情况下，使用`kubectl apply`命令即可将***Shifu***安装至集群：

```sh
cd shifu
kubectl apply -f pkg/k8s/crd/install/shifu_install.yml
```

### 试玩

我们为您准备了 [***Shifu*** Demo](https://shifu.run/zh-Hans/disclaimer/)。

您可以下载Demo体验 ***Shifu*** 如何通过数字孪生来连接和管理实体设备。

### 深入了解

了解更多内容请查看我们的[文档](https://shifu.run/zh-Hans/docs/)。

## 设计理念
#### 📡 易于部署
Shifu必须只用一个命令即可完成部署。
#### 🤖即插即用
Shifu必须自动识别并为一个新的物联网设备提供基本功能。一旦开发者完成了Shifu的模板，设备的所有功能就应该立即可用。
#### 🪄易于扩展
开发者可以进一步实现Shifu的接口/SDK，以创建自定义功能，释放出无限的可能性。
#### 🔧零维护
Shifu的目标是通过采用前沿的云原生最佳实践来实现零维护。毕竟，Shifu需要先照顾好自己，才能让物联网开发人员的工作更轻松!

## 社区

### 贡献

***Shifu*** 欢迎您 [新建 Issue](https://github.com/Edgenesis/shifu/issues/new) 或 [提交 PR](https://github.com/Edgenesis/shifu/pulls)。

### 联系我们

- 电子邮件
    - info@edgenesis.com 
- 微信
    - Donoteattoomuchla 
    - if7369

## GitHub Star 数量

[![Stargazers over time](https://starchart.cc/Edgenesis/shifu.svg)](https://starchart.cc/Edgenesis/shifu)
