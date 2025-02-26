<div align="right">

[中文](README-zh.md) | [English](README.md) | 日本語

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
    KubernetesネイティブのIoTゲートウェイ
    <br />
</h1>
ShifuはKubernetesネイティブで、プロトコルやベンダーに依存しない、プロダクショングレードのオープンなIoTゲートウェイです。
</div>
</div>
<br/><br/>

<div align="center">
    <a href="https://discord.gg/2tbdBrkGHv"><img src="https://img.shields.io/badge/-Discord-424549?style=social&logo=discord" height="25"></a>
    &nbsp;
    <a href="https://twitter.com/ShifuFramework"><img src="https://img.shields.io/badge/-Twitter-red?style=social&logo=twitter" height="25"></a>
    &nbsp;
    <a href="https://www.linkedin.com/company/76257633/admin/"><img src="https://img.shields.io/badge/-Linkedin-red?style=social&logo=linkedin" height="25"></a>
     &nbsp;
    <a href="https://github.com/Edgenesis/shifu"><img src="https://img.shields.io/github/stars/Edgenesis/shifu?style=social" height="25"></a>
</div>

## Shifuの価値: アプリケーションを開発し、インフラを構築しない

<div align="center">
<img width="900px" src="./img/iot-stack-with-shifu.svg"></img>
</div>

## CNCFライブデモ

[![Cloud Native Live](https://img.youtube.com/vi/qMrdM1QcLMk/maxresdefault.jpg)](https://www.youtube.com/watch?v=qMrdM1QcLMk)

## 特徴

**Kubernetesネイティブ** — アプリケーション開発とデバイス管理を同時に行い、追加の運用インフラを構築する必要がありません。

**オープンプラットフォーム** — ベンダーロックインなしで、Shifuをパブリッククラウド、プライベートクラウド、ハイブリッドクラウドのいずれでも簡単に展開できます。ShifuはKubernetesをIoTエッジコンピューティングシーンに導入し、IoTアプリケーションのスケーラビリティと高可用性を実現します。

**プロトコルに依存しない** — HTTP、MQTT、RTSP、Siemens S7、TCPソケット、OPC UAなど、Shifuのマイクロサービスアーキテクチャにより、新しいプロトコルの迅速な統合が可能です。

## 用語集

**shifu** — IoTデバイスをKubernetesクラスターに統合するためのCRD（カスタムリソース定義）。

**DeviceShifu** — Kubernetesのポッドであり、Shifuの最小単位です。主にデバイスのドライバが含まれており、クラスタ内でIoTデバイスを表現します。これを「デジタルツイン」と呼ぶこともできます。

<div align="center">
<img width="900px" src="./img/shifu-architecture.png"></img>
</div>

## 5行コードで私有プロトコルのカメラに接続する方法

<div align="center">
<img width="900px" src="./img/five-lines-to-connect-to-a-camera.gif"></img>
</div>

## コミュニティ

Shifuコミュニティに参加し、あなたの考えやアイデアをシェアしてください。あなたの意見は非常に貴重です。
皆さんの参加を心待ちにしています。

[![Discordで参加](https://img.shields.io/badge/Discord-join-brightgreen)](https://discord.gg/CkRwsJ7raw)
[![Twitterでフォロー](https://img.shields.io/badge/Twitter-follow-blue)](https://twitter.com/ShifuFramework)
[![GitHub Discussionsで議論](https://img.shields.io/badge/GitHub%20Discussions-post-orange)](https://github.com/Edgenesis/shifu/discussions)

## はじめに

詳細な情報は[Shifuドキュメント](https://shifu.dev/)をチェックしてください：
- 🔧[Shifuのインストール](https://shifu.dev/docs/guides/install/install-shifu-dev)
- 🔌[デバイスの接続](https://shifu.dev/docs/guides/cases/)
- 👨‍💻[アプリケーション開発](https://shifu.dev/docs/guides/application/)
- 🎮[KillerCodaデモを試す](https://killercoda.com/shifu/shifu-demo)

## 貢献

[Issueを投稿](https://github.com/Edgenesis/shifu/issues/new/choose)したり、[PRを提出](https://github.com/Edgenesis/shifu/pulls)してください！

貢献してくださった皆様に感謝します。

## Shifuは[CNCFランドスケーププロジェクト](https://landscape.cncf.io/)に正式に登録されています。

<div align="center">
<img width="900px" src="./img/cncf-logo.png"></img>
</div>

## GitHubスター数の推移

[![Stargazers over time](https://starchart.cc/Edgenesis/shifu.svg)](https://starchart.cc/Edgenesis/shifu)

## ライセンス

このプロジェクトはApache 2.0ライセンスの下で配布されています。
