# Shifu Framework TelemetryService design

`TelemetryService` is part of the `Shifu` CRD. It describes a service endpoint which ***deviceShifu*** can push data to using the telemetry configuration, checkout the [telemetry configuration guide](docs/design/deviceshifu/telemetry.md)

## Design goals and non-goals

### Design goals

### Design non-goals

#### 100% Compatiblility

A `TelemetryService` object should be able to describe most existing service endpoints such as `MySQL`, `HTTP` servers, `MQTT` endpoints and etc. But it should not be able to describe and compatible with 100% of the existing service endpoints.

## Architecture

A `TelemetryService` object consists following configuration:

- name
- serviceType
- serviceSettings

### serviceType

`serviceType` is the type of service, can be `HTTP` for now, with more support on the way.

### serviceSettings

`serviceSettings` is settings related to the specific service.

## Example

```yaml
--- #telemetry_service.yaml
apiVersion: shifu.edgenesis.io/v1alpha1
kind: TelemetryService
metadata:
  name: push-endpoint-1
spec:
  type: HTTP
  address: 1.2.3.4:1234/api1
  serviceSettings:
    HTTPSetting:
      username: admin
      password: password
```
