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
        "datatime":"2022-06-30 07:55:51",
        "eUnit":"℃",
        "eValue":"19.9",
        "eKey":"e3",
        "eName":"大气温度",
        "eNum":"101"
    },
    {
        "datatime":"2022-06-30 07:55:51",
        "eUnit":"%RH",
        "eValue":"88.2",
        "eKey":"e4",
        "eName":"大气湿度",
        "eNum":"102"
    }]
}
```

`parsed data` required by application:

```json
{
    [{
        "code":"20990922009",
        "name":"大气温度",
        "val":"37",
        "unit":"℃",
        exception:"温度过高"
    },
    {
        "code":"20990922009",
        "name":"大气湿度",
        "val":"35",
        "unit":"",
        exception:"湿度过高"
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


Give driver developers a SDK to work with deviceShifu.

`deviceShifu.py` developers should edit this file in the following steps:

1. create a handler function, such as `func_a`.
2. register the handler function by editing `register_handlers`.

```python
class DeviceShifu():
    def __init__(self):
        self.register_handlers()

    def register_handlers(self):
        self.register_handler(self.func_a)
        self.register_handler(self.func_b)

    def start(self):

    # User defined callback handler functions
    # To debug, simply print to stdout and access from kubectl logs (we should have shifuctl too!)
    def func_a(self, raw_data):

    def func_b(self, raw_data):
```

`deviceShifu.py` developers shouldn't edit this file.

```python
def main():
    ds = DeviceShifu()
    ds.start()
```
