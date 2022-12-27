# deviceShifu RTSP overall design

deviceShifu-RTSP is deviceShifu-TCP with a application in kubernetes, which allows shifu connect to rtsp server and forward the rtsp stream. User can watch rtsp video using media player like VLC and using ffmpeg or other third-party package to store the stream into PV in other host which has RTX 4090.

## Goal

Create a deviceShifu-RTSP include deviceShifu-TCP and a application running in the kubernetes allow user to connect RTSP server, and record the rtsp stream into file like mp4 / flv and so on, user can export file out of kubernetes.

## General Design

Create a deviceShifu-TCP forward the all data from target server, user can connect to deviceShifu-TCP to connect the device.

The application is a container running in the kubernetes, which can store the rtsp stream into a file like mp4 / flv and so on, user can export file out of kubernetes with the instruction manual.

By Default, the application connect to the deviceShifu-TCP to achieve deviceShifu-RTSP, so that user can record the RTSP Stream into file, or watch the RTSP Stream in VLC from deviceShifu.
## Detailed Design

### Protocol Specification

deviceShifu-TCP just forward all data from target server out, user can connect to deviceShifu-TCP like to connect the real device by tcp.

deviceShifu-TCP not support Restful API

edgedevice.yaml
```yaml
spec:
  sku: tcp Device
  connection: Ethernet
  address: 192.168.0.1:1080
  protocol: TCP
```

### Recording Application
```mermaid
flowchart LR
  dv[device]
  subgraph k8s
    ds[deviceshifu-TCP]
    rrs[RTSP Recording Server]
    pv[Persistent Volumes]
  end
  dv --> ds --> rrs --> pv
```

We need to store the video in pieces, each clip is 1 hour by default, the user can set this option, user can customize the path where the files saved

this application can implemenet with a shell script, or a Binary files and so on
ffmpeg example:
```bash
ffmpeg -i rtsp://[RTSP_URL] -c copy -map 0 -segment_time 300 -f segment video/output%03d.mp4
```

The application provides two interfaces, such as `/record/on` and `/record/off`, to control the record state and can stop the record gracefully.

Create PVC and PV need Separate with other deploy to avoid user delete pv by mistake. 
```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: <devicename>-pv
  namespace: deviceshifu
  labels:
    pv: <devicename>-pv
spec:
  capacity:
    storage: 10Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: slow
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: <devicename>-pvc
spec:
  resources:
    requests:
      storage: 10Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
  selector:
    matchLabels:
      pv: <devicename>-pv
```
recording server deployment.yaml
```yaml
# deployment.yaml
...
- env:
  RTSPServerAddress: ""
  UserName: ""
  PasswordSecret: "" # this just a case, secret using the secret env mode
  FileSavePath: ""
...
volumeMounts:
- name: storage-pv
  mountPath: /data
...
volumns:
- name: storge-pv
  persistentVolumeClaim:
    claimName:  pvc-name
    optional: true
```

### Testing Plan

We can use Hikvision / Dahua Camera, create a deviceShifu-TCP and application to connect it and record the stream into file. Meanwhile we can create a mock-rtsp server or build a mock-rtsp device docker image to test.
