# HTTP to PowerShell driver stub
## Introduction
In order for ***Shifu*** to integrate with your driver. We have implemented a simple HTTP-to-PowerShell stub written in Go for developers to try out.

### Design
This HTTP to PowerShell stub is designed the following way:
- The stub exposes an HTTP interface on the host machine
- The HTTP interface is used to forward requests from external to the `Windows` host machine
- The stub will proxy the result and execution status back to the requestor

### Functionality
#### Proxy HTTP body to PowerShell shell and execute
The main function for this stub is to take whatever passed to it in the HTTP body and issue the command with a specified timeout.


## Building:
### To build the stub, use:

`386`:
```bash
GOOS=windows GOARCH=386 go build -a -o http2powershell.exe cmd/httpstub/powershellstub/powershellstub.go
```

`amd64`:
```bash
GOOS=windows GOARCH=amd64 go build -a -o http2powershell.exe cmd/httpstub/powershellstub/powershellstub.go
```

## Usage:

The executable takes the following environmental variables:
- `EDGEDEVICE_DRIVER_HTTP_PORT` (Optional)
  - The HTTP server port of the driver container, default to `11112`
- `EDGEDEVICE_DRIVER_EXEC_TIMEOUT_SECOND` (Optional)
  - The timeout of an execution, this is achieved by appending `timeout <seconds>` in front of the command

### On `Windows` host:
To run the stub, double click `http2powershell.exe`, by default the stub listens to port `11112` on `0.0.0.0`

### On `Shifu`:
Use the sample deployment files provided in`/examples/simple-powershell-stub`

In `shifu`'s root directory issue:
```bash
kubectl apply -f driver_util/http-to-powershell-stub/examples/simple-powershell-stub
```

### Proxy the command:
Use cURL to post request the `Windows` host:
```bash
root@nginx:/# curl "edgedevice-powershell/issue_cmd?flags_no_parameter=ls,C:"


    Directory: C:\


Mode                 LastWriteTime         Length Name                                                   
----                 -------------         ------ ----                                                   
d-----          6/5/2021   8:10 PM                PerfLogs                                               
d-r---          6/9/2022   2:48 PM                Program Files                                          
d-r---         4/29/2022   8:02 PM                Program Files (x86)                                    
d-r---         4/16/2022   1:46 AM                Users                                                  
d-----          6/9/2022   2:48 PM                Windows                                                
d-----         4/17/2022   5:23 PM                xampp                                                  

root@nginx:/# curl "edgedevice-powershell/issue_cmd?flags_no_parameter=ping,8.8.8.8"

Pinging 8.8.8.8 with 32 bytes of data:
Reply from 8.8.8.8: bytes=32 time=64ms TTL=114
Reply from 8.8.8.8: bytes=32 time=56ms TTL=114
Reply from 8.8.8.8: bytes=32 time=57ms TTL=114
Reply from 8.8.8.8: bytes=32 time=59ms TTL=114

Ping statistics for 8.8.8.8:
    Packets: Sent = 4, Received = 4, Lost = 0 (0% loss),
Approximate round trip times in milli-seconds:
    Minimum = 56ms, Maximum = 64ms, Average = 59ms
```

### For example:

When using CURL to post a request to a given URL, the command looks like the following:

`curl "example.com/issue_cmd?flags_no_parameter=ping,8.8.8.8`

The request will then passes from the HTTP stub into the `PowerShell` of the `Windows` host:

`> powershell.exe ping 8.8.8.8`

Note that the default timeout `EDGEDEVICE_DRIVER_EXEC_TIMEOUT_SECOND` can be overwritten by the `timeout` flag in URL, for example:
Without flag(command timeout, incomplete output):
```bash
root@nginx:/# curl "example.com/issue_cmd?flags_no_parameter=ping,-n,6,8.8.8.8"   

Pinging 8.8.8.8 with 32 bytes of data:
Reply from 8.8.8.8: bytes=32 time=58ms TTL=114
Reply from 8.8.8.8: bytes=32 time=51ms TTL=114
Reply from 8.8.8.8: bytes=32 time=59ms TTL=114
Reply from 8.8.8.8: bytes=32 time=45ms TTL=114
Reply from 8.8.8.8: bytes=32 time=59ms TTL=114
```

With flag(complete output):
```bash
root@nginx:/# curl "example.com/issue_cmd?timeout=10&flags_no_parameter=ping,-n,6,8.8.8.8" 

Pinging 8.8.8.8 with 32 bytes of data:
Reply from 8.8.8.8: bytes=32 time=60ms TTL=114
Reply from 8.8.8.8: bytes=32 time=60ms TTL=114
Reply from 8.8.8.8: bytes=32 time=59ms TTL=114
Reply from 8.8.8.8: bytes=32 time=59ms TTL=114
Reply from 8.8.8.8: bytes=32 time=59ms TTL=114
Reply from 8.8.8.8: bytes=32 time=60ms TTL=114

Ping statistics for 8.8.8.8:
    Packets: Sent = 6, Received = 6, Lost = 0 (0% loss),
Approximate round trip times in milli-seconds:
    Minimum = 59ms, Maximum = 60ms, Average = 59ms
```

We also added a parameter `stub_toleration` to handle latency issue between deviceShifu and the stub. By default it is set to `1` second. You can override this using the following:
```bash
root@nginx:/# curl "example.com/issue_cmd?timeout=10&flags_no_parameter=ping,-n,6,8.8.8.8&stub_toleration=0" 
```
