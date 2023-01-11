package deviceshifutcp

import (
	"fmt"
	"io"
	"net"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
)

type DeviceShifu struct {
	base          *deviceshifubase.DeviceShifuBase
	TcpConnection *ConnectMetaData
}

type ConnectMetaData struct {
	ForwardAddress string
	Ln             net.Listener
}

func New(deviceShifuMetadata *deviceshifubase.DeviceShifuMetaData) (*DeviceShifu, error) {
	base, _, err := deviceshifubase.New(deviceShifuMetadata)
	if err != nil {
		return nil, err
	}
	var cm *ConnectMetaData
	if deviceShifuMetadata.KubeConfigPath != deviceshifubase.DeviceKubeconfigDoNotLoadStr {
		switch protocol := *base.EdgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolTCP:
			connectionType := base.EdgeDevice.Spec.ProtocolSettings.TCPSetting.NetworkType
			if connectionType == nil || *connectionType != "tcp" {

				return nil, fmt.Errorf("Sorry!, Shifu currently only support TCP Socket")
			}
			ListenAddress := ":" + *base.EdgeDevice.Spec.ProtocolSettings.TCPSetting.ListenPort
			Listener, err := net.Listen("tcp", ListenAddress)
			if err != nil {
				return nil, fmt.Errorf("Listen error")
			}
			cm = &ConnectMetaData{
				ForwardAddress: *base.EdgeDevice.Spec.Address,
				Ln:             Listener,
			}

		}
	}
	ds := DeviceShifu{base: base, TcpConnection: cm}

	return &ds, nil
}
func (m *ConnectMetaData) handleTCPConnection(conn net.Conn) {
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
func (m *ConnectMetaData) Start(Stopch <-chan struct{}) error {
	for {
		conn, err := m.Ln.Accept()
		if err != nil {
			return err
		}
		// create a new goroutine
		go m.handleTCPConnection(conn)
	}
}
