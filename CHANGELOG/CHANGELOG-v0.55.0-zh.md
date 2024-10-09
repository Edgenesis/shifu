

# 自 [v0.54.0](https://github.com/Edgenesis/shifu/releases/tag/v0.54.0) 以来的变更

## 新功能 🎉

* 更新 Golang 至 1.23.1，修复 makefile，提升 kubebuilder 版本，由 @tomqin93 在 [PR #992](https://github.com/Edgenesis/shifu/pull/992) 完成

## Bug 修复

* [问题 #994] 修复 SQL Server E2E 测试不一致问题，由 @tomqin93 在 [PR #995](https://github.com/Edgenesis/shifu/pull/995) 完成

## 改进

* [问题 #984] 修复 lint 问题，更新 golangci-lint 和 action 版本，由 @tomqin93 在 [PR #993](https://github.com/Edgenesis/shifu/pull/993) 完成

* 暂时移除 PLC4X arm64 构建，直到 Go 1.23.3 发布，由 @tomqin93 在 [PR #999](https://github.com/Edgenesis/shifu/pull/999) 完成

## Dependabot 自动更新 🤖

* 由 @dependabot 将 golang.org/x/net 从 0.29.0 升级到 0.30.0，在 [PR #998](https://github.com/Edgenesis/shifu/pull/998) 完成

* 由 @dependabot 将 github.com/taosdata/driver-go/v3 从 3.5.6 升级到 3.5.8，在 [PR #989](https://github.com/Edgenesis/shifu/pull/989) 完成

**完整变更日志**: https://github.com/Edgenesis/shifu/compare/v0.54.0...v0.55.0

