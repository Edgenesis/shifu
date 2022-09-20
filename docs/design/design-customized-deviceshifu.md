# Customized deviceShifu design

Customizing deviceShifu will enable deviceShifu to handle customized logic such as parsing data from devices.

## Data flow

```mermaid
sequenceDiagram
    participant id as IoT device
    participant cd as customized deviceShifu
    id->>cd: raw data
    cd->>cd: parse raw data via handler function
    cd->>application: parsed data
```
Anything between `raw data` and `app-friendly data` are components in `deviceShifu`:
```mermaid
  flowchart LR
    IoT-device --> |raw data|deviceShifu-data-adapter-->deviceShifu-data-handler-->deviceShifu-data-sender-->|app-friendly data|application

```

Examples:

`raw data` provided by IoT device:

```
root@nginx:/# curl http://deviceshifu-test.deviceshifu.svc.cluster.local/18000851;echo
```

```json
{
    "statusCode": "200",
    "message":"success",
    "entity":[{
        "deviceId":"20990922009",
        "datatime":"2022-06-30 07:55:51",
        "eUnit":"℃",
        "eValue":"37",
        "eKey":"e3",
        "eName":"atmosphere temperature",
        "eNum":"101"
    },
    {
        "deviceId":"20990922009",
        "datatime":"2022-06-30 07:55:51",
        "eUnit":"%RH",
        "eValue":"88",
        "eKey":"e4",
        "eName":"atmosphere humidity",
        "eNum":"102"
    }]
}
```

`parsed data` required by application:

```json
{
    [{
        "code":"20990922009",
        "name":"atmosphere temperature",
        "val":"37",
        "unit":"℃",
        "exception":"temperature is too high"
    },
    {
        "code":"20990922009",
        "name":"atmosphere humidity",
        "val":"88",
        "unit":"",
        "exception":"humidity is too high"
    }]
}
```

## Developer workflow

```mermaid
flowchart LR
    subgraph devgd[develop general deviceShifu]
        wdyf[Write deviceShifu yaml files]-->dd[deploy general deviceShifu]
        dd-->tdc[test device connection]
    end

    subgraph devcd[develop customized deviceShifu]
        idhf[implement deviceShifu handler functions]
        idhf-->tdhf[test deviceShifu handler functions]
        tdhf-->bcde[build customized deviceShifu executable]
        bcde-->bcdi[build customized deviceShifu image]
        bcdi-->dcd[deploy customized deviceShifu]
        dcd-->tcd[test customized deviceShifu]
    end

    devgd-->devcd
```

## Implementation

By default, `deviceShifu` will provide raw data from physical device to the applications.

To make `deviceShifu` able to "translate" the raw data to a more application-friendly format as well as filter out unneeded data, we can provide customized handlers to `deviceShifu`. For example, using Python, we add customized handlers like this:

1. in `deviceshifu/python_customized_handlers/customized_handlers.py`, add a new function with the function name being instruction name.
2. make sure the function takes a string (raw_data) and returns a string (processed_data which app-friendly).


By default, `deviceShifu` will use the default command handler to process the data of each instruction, and calling the endpoint of the instruction will give a response of raw data from device. 

If user registers the custom handler, `deviceShifu` will switch to use the custom handler for that instruction. As a result, a call to the endpoint of that instruction will respond the processed data instead of raw data.

As for error handling, we expect the `deviceShifu` to log out an error on a handler failure. 

**customized_handlers.py**\
We add a function to process data from instruction ```humidity```:
```python

# function name is the instruction name, i.e., in configmap, there will be an instruction named "humidity"
def humidity(raw_data):
    new_data = []
    # raw_loaded = json.load(raw_data)
    # translate the raw data to new data

    entities = raw_data["entity"]
    for i in range(len(entities)):
        new_data_entry = {"code": entities[i]["deviceId"],
                          "name": entities[i]["eName"],
                          "val": entities[i]["eValue"],
                          "unit": entities[i]["eUnit"],
                          "exception": check_regular_measurement_exception(entities[i]["eName"], entities[i]["eValue"])}
        new_data.append(new_data_entry)
    return new_data

```
and we can have helper and constants here like this:
```python
    TEMPERATURE_MEASUREMENT = "atmosphere temperature"
    HUMIDITY_MEASUREMENT = "atmosphere humidity"


    def check_regular_measurement_exception(measurement_name, measurement_value):
        exception_message = ""
        if measurement_name == TEMPERATURE_MEASUREMENT:
            if int(measurement_value) > 35:
                exception_message = "temperature is too high"
        elif measurement_name == HUMIDITY_MEASUREMENT:
            if int(measurement_value) > 60:
                exception_message = "humidity is too high"

        return exception_message
```

## Internal structure of data
`deviceShifu`'s handling of data contains 4 components: data-adapter, data-handler, data-provider, data-cache:

1. **data-adapter** is responsible for receiving data from physical device. `deviceShifu` loads the driver and enables the data flow from physical device to data-adapter.
2. **data-handler** has custom-implemented handlers that process the data in action, handler function is invoked every time new data comes.
3. **data-provider** is the portal for actively sending data to applications or for applications to ask for data.
4. **data-cache** is cutsomizable to store data used very frequently.

