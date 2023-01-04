# deviceShifu TCP overall design

deviceShifu-TCP allows Shifu connect to any server which protocol based tcp and forward it out, like nginx.

## Goal

deviceShifu-TCP can forward all data from target to support all protocol base tcp.

## General Design

Create a deviceShifu-TCP forward the all data from target server, user can connect to deviceShifu-TCP to connect the device.

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
edgedevices_type.go

Add a new protocoltype ProtocolTCP
### Testing Plan

- forward TCP protocol: using mock tcp server and try to connect to the server with deviceShifu-TCP
- forward HTTP protocol: using mock tcp server and try to connect to the server with deviceShifu-TCP
- forward RTSP protocol: use Hikvision / Dahua Camera, try to connect to the server with deviceShifu-TCP

### Demo

```go
type MetaData struct {
	ForwardAddress string
	ln             net.Listener
}

func New(deviceShifuMetadata *deviceshifubase.DeviceShifuMetaData) (*MetaData, error) {
	base, _, err := deviceshifubase.New(deviceShifuMetadata)
	if err != nil {
		return nil, err
	}

	if deviceShifuMetadata.KubeConfigPath != deviceshifubase.DeviceKubeconfigDoNotLoadStr {
		switch protocol := *base.EdgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolTCP:
			connectionType := base.EdgeDevice.Spec.ProtocolSettings.TCPSetting.NetworkType
			if connectionType == nil || *connectionType != "tcp" {

				return nil, fmt.Errorf("Sorry!, Shifu currently only support TCP Protocal")
			}
			listener, err := net.Listen("tcp", ":40080")//default port
			if err != nil {
				return nil, fmt.Errorf("Listen error")
			}
			return &MetaData{
				ForwardAddress: *base.EdgeDevice.Spec.Address,
				ln:             listener,
			}, nil

		}
	}

	return nil, nil
}
func (m *MetaData) handleTCPConnection(conn net.Conn) {
	defer conn.Close()
	// Forward the TCP connection to the destination
	forwardConn, err := net.Dial("tcp", m.ForwardAddress)
	if err != nil {
		// Handle error
		return
	}
	defer forwardConn.Close()
	// Copy data between bidirectional connections.
	go io.Copy(forwardConn, conn)
	io.Copy(conn, forwardConn)
}

func (m *MetaData) Start() error {
	for {
		conn, err := m.ln.Accept()
		if err != nil {
			return err
		}

		// create a new goroutine
		go m.handleTCPConnection(conn)
	}
}
```

