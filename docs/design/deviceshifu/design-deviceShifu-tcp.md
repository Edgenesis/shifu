# deviceShifu TCP overall design

deviceShifu-TCP allows Shifu connect to any server which protocol based tcp and forward it out, like nginx.

## Goal

- deviceShifu-TCP can forward all data from target to support all protocol base tcp.

## General Design

Create a deviceShifu-TCP forward the all data from target server, user can connect to deviceShifu-TCP to connect the device.

## Detailed Design

### Protocol Specification

deviceShifu-TCP just forward all data from target server out, user can connect to deviceShifu-TCP like to connect the real device by tcp.

deviceShifu-TCP not support RESTful API

edgedevice.yaml
```yaml
spec:
  sku: tcp Device
  connection: Ethernet
  address: 192.168.0.1:1080
  protocol: TCP
```

### Testing Plan

- forward TCP protocol: using mock tcp server and try to connect to the server with deviceShifu-TCP
- forward HTTP protocol: using mock tcp server and try to connect to the server with deviceShifu-TCP
- forward RTSP protocol: use Hikvision / Dahua Camera, try to connect to the server with deviceShifu-TCP
