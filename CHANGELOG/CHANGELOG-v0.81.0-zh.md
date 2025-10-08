# 自 [v0.80.0](https://github.com/Edgenesis/shifu/releases/tag/v0.80.0) 以来的变更

## 新功能 🎉

- 新增 DeviceShifu LwM2M 单元测试，提升测试覆盖率和可靠性

## Bug 修复 🐛

- 移除 CI 流水线中的 macOS 演示测试，避免在不支持的平台上导致失败

## 功能增强 ⚡

- 将 Kubebuilder 从 v3 升级到 v4，利用新特性并保持兼容性

## Breaking Change 💥

- **移除 PLC4X DeviceShifu 模块** (#1271)
  - 移除 PLC4X 集成模块，包括 `pkg/deviceshifu/deviceshifuplc4x/`、`cmd/deviceshifu/cmdplc4x/` 及相关示例
  - 从 EdgeDevice CRD 中移除 `v1alpha1.ProtocolPLC4X` 协议类型
  - 移除 `deviceshifu-http-plc4x` Docker 镜像及构建基础设施
  - 移除 PLC4X Go 模块依赖 (`github.com/apache/plc4x/plc4go`)
  - **迁移指南**：用户可以继续使用 Shifu v0.81.0 之前的版本以获得 PLC4X 支持,或联系 info@edgenesis.com 请求协议支持。我们将在未来通过替代实现方式继续支持工业协议(如 Modbus TCP/RTU、BACnet)
  - **移除原因**：使用率低、构建复杂度高、依赖过时。协议支持将根据实际用户需求按需实现

## 文档更新 📚

- 更新社交徽章和 Shifu 标志，提升项目品牌形象

- 新增 v0.80.0 版本变更日志，保持发布历史的完整性

## Dependabot 自动更新 🤖

- Bump github.com/pion/dtls/v3 from 3.0.6 to 3.0.7 by @dependabot[bot] in https://github.com/Edgenesis/shifu/pull/1278

- Bump k8s.io/client-go from 0.34.0 to 0.34.1 by @dependabot[bot] in https://github.com/Edgenesis/shifu/pull/1280

- Bump github.com/taosdata/driver-go/v3 from 3.7.3 to 3.7.6 by @dependabot[bot] in https://github.com/Edgenesis/shifu/pull/1284

- Bump github.com/eclipse/paho.mqtt.golang from 1.5.0 to 1.5.1 by @dependabot[bot] in https://github.com/Edgenesis/shifu/pull/1301

- Bump github.com/onsi/ginkgo/v2 from 2.25.3 to 2.26.0 by @dependabot[bot] in https://github.com/Edgenesis/shifu/pull/1303

**完整变更日志**: https://github.com/Edgenesis/shifu/compare/v0.80.0...v0.81.0