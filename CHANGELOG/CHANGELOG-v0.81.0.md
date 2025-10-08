# Changelog since [v0.80.0](https://github.com/Edgenesis/shifu/releases/tag/v0.80.0)

## New Features 🎉

- Add unit tests for DeviceShifu LwM2M to improve test coverage and reliability

## Bug Fixes 🐛

- Remove macOS demo test from CI pipeline to prevent failures on unsupported platforms

## Enhancements ⚡

- Migrate Kubebuilder from v3 to v4 to leverage new features and maintain compatibility

## Breaking Changes 💥

- **Remove PLC4X DeviceShifu Module** (#1271)
  - Removed PLC4X integration module including `pkg/deviceshifu/deviceshifuplc4x/`, `cmd/deviceshifu/cmdplc4x/`, and related examples
  - Removed `v1alpha1.ProtocolPLC4X` protocol type from EdgeDevice CRD
  - Removed `deviceshifu-http-plc4x` Docker images and build infrastructure
  - Removed PLC4X Go module dependency (`github.com/apache/plc4x/plc4go`)
  - **Migration**: Users can continue using Shifu versions <v0.81.0 for PLC4X support, or contact info@edgenesis.com for protocol support requests. We will continue to support industrial protocols (such as Modbus TCP/RTU, BACnet) in the future through alternative implementations
  - **Rationale**: Low usage, build complexity, and outdated dependencies. Protocol support will be implemented on-demand based on actual user requirements

## Documentation 📚

- Refresh social badges and update Shifu logo for improved project branding

- Add changelog for v0.80.0 to keep release history up to date

## Dependabot Updates 🤖

- Bump github.com/pion/dtls/v3 from 3.0.6 to 3.0.7 by @dependabot[bot] in https://github.com/Edgenesis/shifu/pull/1278

- Bump k8s.io/client-go from 0.34.0 to 0.34.1 by @dependabot[bot] in https://github.com/Edgenesis/shifu/pull/1280

- Bump github.com/taosdata/driver-go/v3 from 3.7.3 to 3.7.6 by @dependabot[bot] in https://github.com/Edgenesis/shifu/pull/1284

- Bump github.com/eclipse/paho.mqtt.golang from 1.5.0 to 1.5.1 by @dependabot[bot] in https://github.com/Edgenesis/shifu/pull/1301

- Bump github.com/onsi/ginkgo/v2 from 2.25.3 to 2.26.0 by @dependabot[bot] in https://github.com/Edgenesis/shifu/pull/1303

**Full Changelog**: https://github.com/Edgenesis/shifu/compare/v0.80.0...v0.81.0