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

***Shifu***是帮助IoT开发者对接物理设备的云原生框架。以虚拟化的的方式对设备管理和监控，赋予其自动化运行的能力。解决了协议与驱动的接入烦恼，使设备即插即用。

## 为什么用 ***Shifu***

- Shifu是Kubernetes原生框架，达到航天级稳定性（99.9999%）。
- 设备协议和驱动以配置文件的方式编译，高效且能复用。
- 模块化部署设备，开发者按需加载即可。

## 开始使用 ***Shifu***

### 安装

***Shifu*** 提供了`shifu_install.yml`文件。在已有Kubernetes集群的情况下，使用`kubectl apply`命令即可安装至集群：

```sh
cd shifu
kubectl apply -f pkg/k8s/crd/install/shifu_install.yml
```

### 演示

如果您不熟悉Kubernetes，我们准备了 [***Shifu*** Demo](https://shifu.run/zh-Hans/disclaimer/)。您可以直观的体验 ***Shifu*** 如何通过数字孪生来连接和管理实体设备。

### 使用文档

请在 <https://shifu.run/zh-Hans/docs/> 查看 ***Shifu*** 的使用文档。

## 深入理解 ***Shifu***

查看 [`docs/`](./docs/) 下的 Markdown文件 来了解 ***Shifu*** 的 [设计细节](./docs/design/) 和 [开发指南](./docs/guide/)。

## ***Shifu*** 愿景

### 让开发者和运维人员再次开心

开发者和维护人员应100%聚焦在发明创造上，而不是修补基础设施以及重复造轮子。身为开发者和运维人员本身，***Shifu*** 的作者们深刻理解您的痛点！所以我们发自内心地想帮您解决掉底层的问题，让开发者和运维人员再次开心！

### 软件定义世界

如果每一个IoT设备都有一个 ***deviceShifu***，我们就可以借助软件来管理我们周围的世界。在一个软件定义的世界 (Software Define World) 中，所有东西都是智能的。您周围的一切会自动改变，进而更好地服务您；因为归根到底，科技以人为本。

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
