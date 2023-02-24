# DeviceShifu Hololens Design
Devceishifu Hololens act as a module that serves core functionality required by Hololens apps. 

###Flowchart:
```mermaid
    flowchart LR
        subgraph holo[hololens]
            hlapa[hololens-application];
        end

        subgraph shifuholo[deviceShifu hololens];
            cdshttp[customized deviceShifuHttp]
            nginx[nginx]
            dss[node dss]
            webrtc[webrtc server]

            cdshttp-->nginx
            nginx-->dss
            nginx-->webrtc
        end

        holo-->shifuholo
        shifuholo--webm-->TelemetryService
```
## General Design
As a module, deviceShifu Hololens would consist 3 parts. Customized deviceshifuHttp would serve as the entry point of the entire module. Customized deviceShifuHTTP will forward the request it received from the client to nginx and nginx will route the request to node dss for signalling or to webRTC server for getting video or audio snippet. The entire module should be in a single deployment.

## Detailed
### DeviceShifuHTTP
We don't need to do particular changes to deviceShifuHTTP, we only need to add certain instructions into the configmap in order to allow deviceShifuHTTP to send requests to webRTC server and Node DSS.
```yaml
instructions: |
    instructions:
      offer:
      audio_recognition:
      image_recognition:
```

### Customized DeviceShifu
We need 2 customized deviceShifu to process the image and audio snippet passed by deviceShifuHTTP. 

```python
def processImage(rawData)
    image = rawData[audio]
    // process image
    return processed_image_json

def processAudio(rawData)
    audio = rawData[audio]
    // process audio
    return processed_audio_json
```

### Node DSS and Nginx
For node dss, we can use the existing node dss server as our signal server, we don't need to do any particular change, we only need to add it to a container and put it into the same pod. 
For nginx, we only need to configure it to route our requests to designated Node DSS or webRTC server and put it as a container to the same pod.


### WebRTC Server
For webRTC server, we already supported capture stream from we need to add more functionalities. 

#### Image and Audio capture:
We need to open up an api for video and audio capture so we can allow user to use them for AI analysis.

```golang
http.HandleFunc("/snapshot", snapshot)
// We can reference: https://github.com/pion/example-webrtc-applications/blob/master/snapshot/main.go
// for capture image snapshot
// For audio we can just use different codec.
```

#### Stream playback
WebRTC server currently can forward the stream from hololens end to web end, however it also need to stream the web's stream to hololens' end. It is not too hard to implement, we only need to do the reverse of the hololens->web end.

```golang
// current connectWeb.Start will return the peerConnection to web.
webpc := connectWeb.Start(videoStreamData, audioStreamData, stop)
// Create Track that we send web video back to hololens on
videoTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8}, "video", "video")
if err != nil {
	return err
}

audioTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "audio")
if err != nil {
	return err
}

rtpAudioSender, audioTrackErr := webpc.AddTrack(audioTrack)
if audioTrackErr != nil {
	anic(audioTrackErr)
}
// Add this newly created track to the PeerConnection
rtpVideoSender, err := webpc.AddTrack(videoTrack)
if err != nil {
	panic(err)
}
// Write track back
webpc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
    // Write track back
    // Refer to how we write the track to web on:
    // https://github.com/Edgenesis/WebRTC/blob/main/client/connectWeb/connectWeb.go#L178-L225
})
```

### Deployment.yaml
We should include all the above module into a single `deployment.yaml` file. Along with necessary configs like `nginx` and newly built customized deviceShifuHTTP image.