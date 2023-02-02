# Telemetry Service TDengine Endpoint Design

## Introduction
Telemetry Service is a standalone service that allow `deviceShifu` send its collected telemetry to designated endpoints. This design aims to add `TDengine` endpoint support to current Telemetry Service.

## Design-Goal
- Let telemetry service support pushing telemetries to TDengine endpoints.

## Design Non-Goal
- Let telemetry service support all timeseriesDB endpoint.

## Design Details
In order to add `TDengine` as an endpoint, we need to add connection related settings telemetry service's CRD, just like  what we did for `MQTT`.

```yaml
SQLConnectionSetting:
    description: SQLConnectionSetting defines SQL specific settings when connecting to SQL endpoint
    properties:
        DBServerAddress:
            type: string
        DBUserName:
            type: string
        DBSecret:
            type: string
        DBName:
            type: string
        DBTable:
            type: string
        DBType:
            type: string
    type: object
```

We also will add the corresponding connection setting as a go struct and will add it to `TelemetryRequest`

```go
type SQLConnectionSetting struct {
	DBServerAddress *string `json:"db_server_address,omitempty"`
	DBUserName      *string `json:"db_username,omitempty"`
	DBSecret        *string `json:"db_secret,omitempty"`
	DBName          *string `json:"db_name,omitempty"`
	DBTable         *string `json:"db_table,omitempty"`
	DBType          *DBType `json:"db_type,omitempty"`
}

type DBType string

type TelemetryRequest struct {
	RawData              []byte                `json:"rawData,omitempty"`
	MQTTSetting          *MQTTSetting          `json:"mqttSetting,omitempty"`
	SQLConnectionSetting *SQLConnectionSetting `json:"sqlConnectionSetting,omitempty"`
}
```

We use the name `SQLConnection` instead of `TDengineConnection` because we want to it to be a generic DB connection setting for scalability concerns. We want to use it to connect to various DBs instead of TDengine alone.

Telemetry Service will push the data to TDengine:
```mermaid
graph LR;
deviceShifu -->|TelemetryRequest| TelemetryService;
TelemetryService -->|RawData| TDengine;

```

After we enabled `TDengine` endpoint ,we need to extract corresponding endpoint settings from `TelemetryRequest`.

```go
func HandleTelemtryRequest(request *TelemetryRequest) err {
	// http part
    // mqtt part
    if (request.SQLConnectionSetting != nil) {
        sqlcs := request.SQLConnectionSetting

    }
}

func sendToTDengine(rawData []byte, sqlcs *v1alpha1.SQLConnectionSetting) err {
	// Send rawData to TDengine
    // sample code:
    taosUri := constructTDengineUri(sqlcs)
    taos, err := sql.Open("taosSql", taosUri)
    if err != nil {
        klog.Error("failed to connect TDengine, err:", err)
        return err
    }
    defer taos.Close()
    ...

}

func constructTDengineUri(sqlcs *v1alpha1.SQLConnectionSetting) (string, error) {

}
```