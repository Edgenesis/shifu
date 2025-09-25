# 自 [v0.79.0](https://github.com/Edgenesis/shifu/releases/tag/v0.79.0) 以来的变更

## 功能增强 ⚡

- 用 Linux 交叉编译替换 macOS 构建器以支持 macOS ARM64，提升构建效率和兼容性

## Bug 修复 🐛

- 修复 lint 问题：将已弃用的 result.Requeue 替换为 RequeueAfter，确保代码库兼容最新标准

## Dependabot 自动更新 🤖

- Bump github.com/onsi/ginkgo/v2 from 2.25.1 to 2.25.3 by @dependabot[bot] in https://github.com/Edgenesis/shifu/pull/1267

- Bump github.com/spf13/cobra from 1.9.1 to 1.10.1 by @dependabot[bot] in https://github.com/Edgenesis/shifu/pull/1266

- Bump k8s.io/api from 0.33.4 to 0.34.1 by @dependabot[bot] in https://github.com/Edgenesis/shifu/pull/1268

**完整变更日志**: https://github.com/Edgenesis/shifu/compare/v0.79.0...v0.80.0