# deviceShifu Telemetry design

***deviceShifu*** Telemetry is part of the ***deviceShifu*** [Compute](docs/design/design-deviceShifu.md#compute) plugin. It periodically pulls device instructions defined in the telemetries section in ***deviceShifu*** ConfigMap.

The result for each telemetry should update `EdgeDevice`'s status accordingly.

## Design goals and non-goals

### Design goals

#### Polymorphism

Telemetry should be protocol-independent, each protocol specific ***deviceShifu*** should implement their own telemetry to achieve desired functionality.

#### Stateless

#### Lightweight

### Design non-goals

## Architecture

Telemetry consistes of telemetry settings, telemetries and telemetry properties.

### Telemetry Settings

Telemetry Settings should consist the following configurable options:

1. `telemetryUpdateIntervalInMilliseconds` (optional, default to `3000`ms)  
  This specifies the update interval for all configured telemetries in milliseconds.

2. `telemetryTimeoutInMilliseconds` (optionalm, default to `3000`ms)  
  This specifies the default timeout for all configured telemetries in milliseconds.

3. `defaultPushToServer` (optional, default to `false`)  
  This specifies the polling mode for telemetries.  
  **false**  
  By default this will be `false` and `deviceShifu` will not act on the result data besides updating the `EdgeDevice`'s status based on the return code/status.  
  **true**  
  If `true` is specified, then each telemetry result will be posted to the endpoint specified `defaultServer` via the default north-bound protocol for the specific ***deviceShifu***.

4. `defaultTelemetryCollectionService` (optional if `defaultPushToServer` is `false`, required for `defaultPushToServer: yes`)  
  This specifies the default posting Kubernetes `Endpoint` for telemetries configured.

### Telemetries

Telemetries are configured method used by ***deviceShifu*** to query its connected `EdgeDevice` for various purposes.

### Telemetry properties

Telemetry properties are configurable property for each telemetry, it should consists the following:

1. `pushToServer` (optional if you would like to use `defaultPushToServer` and its , else it will post the specified telemetry's result to the configured endpoint)  

2. `pushSettings` (optional if you would like to use the global related settings)

   2.1  `telemetryCollectionService` (optional if you would like to use the global `defaultTelemetryCollectionService`)

## Examples

### Example active telemetry configuration for HTTP protocol

With the following configuration in `ConfigMap`:

```yaml
telemetrySettings:
  telemetryUpdateIntervalInMilliseconds: 6000
  telemetryTimeoutInMilliseconds: 3000
  defaultPushToServer: true
  defaultTelemetryCollectionService: push-endpoint-1
telemetries:
  device_health:
    properties:
      instruction: hello
      pushSettings:
        telemetryCollectionService: push-endpoint-2
  device_health2:
    properties:
      pushSettings:
        pushToServer: false
      instruction: hello2
  device_health3:
    properties:
      instruction: hello3
```

The `telemetryCollectionService` endpoint can be defined through Shifu's `TelemetryService` CRD:

```yaml
---telemetry_service.yaml
apiVersion: shifu.edgenesis.io/v1alpha1
kind: TelemetryService
metadata:
  name: push-endpoint-1
  namespace: devices
spec:
  type: HTTP
  address: 1.2.3.4:8081/api
```

this ***deviceShifu*** will have the following telemetries:

1. *device_health*: query `EdgeDevice` using the `hello` instruction every `6` seconds, with a timeout of `3` seconds and push the result to `push-endpoint-2` using the endpoint's specific protocol.

2. *device_health2*: query `EdgeDevice` using the `hello2` instruction every `6` seconds, with a timeout of `3` seconds and do nothing.

3. *device_health3*: query `EdgeDevice` using the `hello3` instruction every `6` seconds, with a timeout of `3` seconds and push the result to `push-endpoint-1` using the endpoint's specific protocol.
