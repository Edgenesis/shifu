# TelemetryService Refactor

## What is TelemetrySerice

***deviceShifu*** Telemetry is part of the ***deviceShifu*** [Compute](docs/design/design-deviceShifu.md#compute) plugin. It periodically pulls device instructions defined in the telemetries section in ***deviceShifu*** ConfigMap.

The result for each telemetry should update `EdgeDevice`'s status accordingly.

## Why need to refactor TelemetryService?

- Currently, telemetryServices' connection settings are read from deviceShifu, which will send all connection settings to telemetryService, so telemetryService strongly depends on deviceShifu.
- telemetryService's connection settings may include some private information, such as secrets read from deviceShifu, so that deviceShifu will have more access to read this information from the kubernetes API, which will make users feel scared.
- The installation yaml file of Shifu needs to be separate from the yaml file of telemetryService.
- telemetryService pushes to HTTP endpoints differently than other protocols, which makes it difficult to use.

### Design Goals

- The yaml for deploying telemetryService will be compatible with older versions.
- Modify the code in deviceShifu about telemetryService, such as `pushToHTTPTelemetryCollectionService` to telemetryService.
- Let telemetryService get the connection settings from the Kubernetes API. deviceShifu only needs to send the telemetryService's name and raw data to telemetryService.
- Developers developing a new endpoint only need to update the telemetryService.

### Non-Goal

- Adding a new protocol to a telemetry service
- Update the YAML file for deploying telemetryService

## Detail Design

We can use Kubernetes dynamicClient to observe the changes in CRD and update the connection settings in time use a goruntine. First, we need to initialize and get all telemetry services from Kubernetes. Then use a sync.Map to restore the TelemetryService Connection Settings 
```go
var dynamicClient 
func init () {
config, err := rest.InClusterConfig()
clientset, err := kubernetes.NewForConfig(config)
dynamicClient = dynamic.NewForConfigOrDie(config)
tss := dynamicClient.Resource().List()
// ... store telemetryServices into tsMap
}

watcher, err := dynamicClient.Resource(&schema.GroupVersionResource{
    Group:    "shifu.edgenesis.io",
    Version:  "v1alpha1",
    Resource: "TelemetryService",
}).Watch(context.Background(), metav1.ListOptions{})
for event := range watcher.ResultChan() {
    switch event.Type {
        case watch.Added:
            fallthrough
        case watch.Modified:
            // ... handle event
            injectSectet(event)
            tsMap.Store()
        case watch.Deleted:
            // ... handle event
            tsMap.Delete()
    }
}
```


Send data from deviceShifu to telemetryService, like
```go
type TelemetryRequest struct {
	RawData              []byte `json:"rawData,omitempty"`
    TelemetryServiceName string `json:"tsName,omitempty"`
}
```


telemetryService HTTP Server like
```go
mux := http.NewServeMux()
mux.HandleFunc("/", TelemetryRequestHandler)
err := Start(stop, mux, serverListenPort)
```
```go
func TelemetryRequestHandler() {
    connectionSetting,exists := tsMap.Load(TelemetryRequest.TelemetryServiceName)
    if !exists {
        return
    }
    if connectionSetting.SQLSetting != nil {}
    if connectionSetting.MQTTSetting != nil {}
    if connectionSetting.MinIOSetting != nil {}
}
```

pushSetting in deviceShifu's Configmap
```yaml
...
  pushSettings:
    telemetryCollectionService: push-endpoint-1
...
```
