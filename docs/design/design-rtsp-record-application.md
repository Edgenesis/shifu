# RTSP recording application 

User can use this application to watch RTSP using ffmpeg or other third-party package to store the stream into PV in other host which supporting GPU that can accelerate video encoding.

## Goal

- application running in the kubernetes allow user to connect RTSP server, and record the RTSP stream into file like mp4 / flv and so on, user can export file out of kubernetes.

## General Design

The application is a HTTP Server running in the kubernetes, which can store the RTSP stream into a file like mp4 / flv and so on, user can export file out of kubernetes with the instruction manual.



## Detailed Design

### Management RTSP Servers

application manage a map to record the info of each device which record and can save this map into a file in PV for Backups, When The Application is up, it will recover from backups first.

user can register his device using `/register`, so that application need using a Kubernetes to get this namespace's secret.

rbac.yaml
```yaml
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: application
  namespace: shifu-app
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: [ "get", "list", "watch", "update", "create","delete"]
```

### Store the Stream

We need to store the video in pieces, each clip is 1 hour by default, the user can set this option, user can customize the path where the files saved

this application could using the third-party package like ffmpeg to implement it or using `exec.Command`
ffmpeg command example:
```bash
ffmpeg -i RTSP://[RTSP_URL] -c copy -map 0 -segment_time 300 -f segment video/output%03d.mp4
```

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

### Serving requests

`/register` regist the RTSP Stream which need to record
```json
{
    "deviceName":"",
    "secretName":"",
    "serverAddress":"",
    "recoding":true, // default recoding when register
}
```

`/unregister` unregister is remove the Target RTSP Server from application and stop to recording but the data exists will not delete.
```json
{
    "deviceName":"",
}
```

`/update` turn on or down the recording from the device(RTSP Server)
```json
{
    "deviceName":"",
    "record":false,
}
```


### Testing Plan

We can use Hikvision / Dahua Camera, create a deviceShifu-TCP and application to connect it and record the stream into file. Meanwhile we can create a mock-RTSP server or build a mock-RTSP device docker image to test.
