# Telemetry Secret Loading Design

## Introduction

Telemetry Service is a standalone service, hereafter referred to as TSCenter. But the settings in TelemetryService CRD is loaded by Deviceshifu and sent to TSCenter. If the setting contains password, there will exist the problem of  plaintext password in comunication between Deviceshifu and TSCenter.

There could be multiple ways to deal with it:

1. Add a custom encode/decode algorithm in communication between Deviceshifu and TSCenter.
2. Let the TSCenter to manage all TelemetryService CRD info. Deivceshifu only specifies the name of TelemetryService CRD. But the TSCenter's workload may be too high.
3. Deviceshifu still send all info as before but hide the password, instead, it will let the TSCenter to find the Secret by the name of TelemetryService CRD.

We will use the last way here. There may still exist some improvement space, like:

1. Every time we push data, TSCenter needs to acquire the Secret.

## General Design

1. TSCenter created with k8s client.
2. Deviceshifu load info from TelemetryService CRD. When sending to TSCenter, Deviceshifu need to add the name and namespace of CRD (for TSCenter to locate the Secret) but hide the password.
3. TSCenter load Secret using received CRD name and namespace and write password into setting.
4. Grant TSCenter-sa with Secret read RBAC.

## Design Details

Currently, only SQLSetting used password.

1. Protocol of communication between Deviceshifu and TSCenter. Add `TelemetryName` and `TelemetryNamespace` in `TelemetryRequest`

   ```Go
   type TelemetryRequest struct {
      // add name and namespace
      TelemetryName        string                `json:"telemetryName,omitempty"`
      TelemetryNamespace   string                `json:"telemetryNamespace,omitempty"`
      RawData              []byte                `json:"rawData,omitempty"`
      MQTTSetting          *MQTTSetting          `json:"mqttSetting,omitempty"`
      SQLConnectionSetting *SQLConnectionSetting `json:"sqlConnectionSetting,omitempty"`
   }
   ```

2. Deviceshifubase sending info of CRD, [code](https://github.com/FFFFFaraway/shifu/blob/2de0c5fb8ba08b5b02d5457efbd10a92110fcec3/pkg/deviceshifu/deviceshifubase/pushtelemetry.go#L19)：
   ```Go
   func PushTelemetryCollectionService(tss *v1alpha1.TelemetryService, message *http.Response) error {
     ...
     if tss.Spec.ServiceSettings.SQLSetting != nil {
       // hide password
       *tss.Spec.ServiceSettings.SQLSetting.Secret = ""
       request := &v1alpha1.TelemetryRequest{
         // add name and namespace
         TelemetryName:        tss.Name,
         TelemetryNamespace:   tss.Namespace,
         SQLConnectionSetting: tss.Spec.ServiceSettings.SQLSetting,
       }
       telemetryServicePath := *tss.Spec.TelemetrySeriveEndpoint + v1alpha1.TelemetryServiceURISQL
       err := pushToShifuTelemetryCollectionService(message, request, telemetryServicePath)
       ...
     }
     ...
   }
   ```
   
   Set password to empty and send telemetry name and namspace.
   
   But currently the telemetry name and namespace is not stored, only spec is stored. [code](https://github.com/FFFFFaraway/shifu/blob/59287c205b3aa8270bada40ddc98966123b5ef59/pkg/deviceshifu/deviceshifubase/deviceshifubase.go#L74)
   
   ```Go
   TelemetryCollectionServiceMap map[string]v1alpha1.TelemetryServiceSpec
   ```
   
   So we need to change it to `map[string]v1alpha1.TelemetryService`, and modify the creation in `getTelemetryCollectionServiceMap`.
   
3. TSCenter
   1. Create k8s client at begining.
      1. ```Go
         func init() {
            config, err := rest.InClusterConfig()
            ...
            clientSet, err = kubernetes.NewForConfig(config)
            ...
         }
         ```
   2. When received SQL request, load the Secret using the CRD name with prefix and write it to setting.
      1. ```Go
         func InjectSecret(setting *v1alpha1.SQLConnectionSetting, tsName, tsNamespace string) {
            ...
            pwd, err := utils.GetPasswordFromSecret(TelemetrySecretPrefix+tsName, tsNamespace)
            *setting.Secret = pwd
            ...
         }
         ```

4. Modify the RBAC yaml files. Add `get`, `list`, `watch` to TSCenter-sa.