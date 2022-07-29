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

### Architecture

Telemetry consistes of telemetry settings, telemetries and telemetry properties.

#### Telemetry Settings

Telemetry Settings should consist the following configurable options:

1. `telemetryUpdateIntervalInMilliseconds` (optional, default to `3000`ms)  
  This specifies the update interval for all configured telemetries in milliseconds.

2. `telemetryTimeoutInMilliseconds` (optionalm, default to `3000`ms)  
  This specifies the default timeout for all configured telemetries in milliseconds.

3. `telemetryMode` (optional, default to `passive`)  
  This specifies the polling mode for telemetries.  
  `passive`  
  By default this will be `passive` and `deviceShifu` will not act on the result data besides updating the `EdgeDevice`'s status based on the return code/status.  
  `active`  
  If `active` is specified, then each telemetry result will be posted to the endpoint specified `telemetryDefaultActiveEndpoint` via the default north-bound protocol for the specific ***deviceShifu***.

4. `telemetryDefaultActiveEndpoint` (optional if `telemetryMode` is `passive`, required for `telemetryMode: active`)  
  This specifies the default posting endpoint for telemetries configured.

#### Telemetries

Telemetries are configured method used by ***deviceShifu*** to query its connected `EdgeDevice` for various purposes.

#### Telemetry properties

Telemetry properties are configurable property for each telemetry, it should consists the following:

1. `telemetryActiveEndpoint` (optional if you would like to use `telemetryDefaultActiveEndpoint`, else it will post the specified telemetry's result to the configured endpoint)  

## Examples

### Example active telemetry configuration for HTTP protocol

With the following configuration:

```yaml
telemetrySettings:
  telemetryUpdateIntervalInMilliseconds: 6000
  telemetryTimeoutInMilliseconds: 3000
  telemetryDefaultActiveEndpoint: 8.8.8.8
  telemetryMode: active
telemetries:
  device_health:
    properties:
      instruction: hello
      telemetryActiveEndpoint: 1.2.3.4:8081/api3
  device_health2:
    properties:
      instruction: hello2
```

the ***deviceShifu*** will have two telemetries:

1. query `EdgeDevice` using the `hello` instruction every `6` seconds, with a timeout of `3` seconds and push the result using the north-bound protocol of the ***deviceShifu*** to `1.2.3.4:8081/api3`

2. query `EdgeDevice` using the `hello2` instruction every `6` seconds, with a timeout of `3` seconds and push the result using the north-bound protocol of the ***deviceShifu*** to `8.8.8.8`
