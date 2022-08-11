package deviceshifuSocket

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/edgenesis/shifu/deviceshifu/pkg/deviceshifubase"
	"github.com/edgenesis/shifu/deviceshifu/pkg/utils"

	"edgenesis.io/shifu/k8s/crd/api/v1alpha1"
)

type DeviceShifu struct {
	base             *deviceshifubase.DeviceShifuBase
	socketConnection *net.Conn
}

type DeviceShifuSocketHandlerMetaData struct {
	edgeDeviceSpec v1alpha1.EdgeDeviceSpec
	instruction    string
	properties     *deviceshifubase.DeviceShifuInstruction
	connection     *net.Conn
}

type deviceCommandHandler interface {
	commandHandleFunc(w http.ResponseWriter, r *http.Request) http.HandlerFunc
}

func New(deviceShifuMetadata *deviceshifubase.DeviceShifuMetaData) (*DeviceShifu, error) {
	base, mux, err := deviceshifubase.New(deviceShifuMetadata)
	if err != nil {
		return nil, err
	}
	var socketConnection net.Conn

	if deviceShifuMetadata.KubeConfigPath != deviceshifubase.DEVICE_KUBECONFIG_DO_NOT_LOAD_STR {
		switch protocol := *base.EdgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolSocket:
			// Open the connection:
			connectionType := base.EdgeDevice.Spec.ProtocolSettings.SocketSetting.NetworkType
			encoding := base.EdgeDevice.Spec.ProtocolSettings.SocketSetting.Encoding
			if connectionType == nil || *connectionType != "tcp" {
				return nil, fmt.Errorf("Sorry!, Shifu currently only support TCP Socket")
			}

			if encoding == nil {
				log.Println("Socket encoding not specified, default to UTF-8")
				return nil, fmt.Errorf("Encoding error")
			}

			socketConnection, err := net.Dial(*connectionType, *base.EdgeDevice.Spec.Address)
			if err != nil {
				return nil, fmt.Errorf("Cannot connect to %v", *base.EdgeDevice.Spec.Address)
			}

			log.Printf("Connected to '%v'\n", *base.EdgeDevice.Spec.Address)
			for instruction, properties := range base.DeviceShifuConfig.Instructions.Instructions {
				deviceShifuSocketHandlerMetaData := &DeviceShifuSocketHandlerMetaData{
					base.EdgeDevice.Spec,
					instruction,
					properties,
					&socketConnection,
				}

				mux.HandleFunc("/"+instruction, deviceCommandHandlerSocket(deviceShifuSocketHandlerMetaData))
			}
		}
	}

	ds := &DeviceShifu{base: base, socketConnection: &socketConnection}
	ds.base.UpdateEdgeDeviceResourcePhase(v1alpha1.EdgeDevicePending)
	return ds, nil
}

func deviceCommandHandlerSocket(deviceShifuSocketHandlerMetaData *DeviceShifuSocketHandlerMetaData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := utils.ParseAllParams(r.URL.String())

		headerContentType := r.Header.Get("Content-Type")
		if headerContentType != "application/json" {
			http.Error(w, "content-type is not application/json", http.StatusBadRequest)
			log.Println("content-type is not application/json")
			return
		}

		var socketRequest DeviceShifuSocketRequestBody
		err := json.NewDecoder(r.Body).Decode(&socketRequest)
		if err != nil {
			log.Printf("error decode: %v", socketRequest)
			http.Error(w, "error decode JSON "+err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("After decode socket request: '%v', timeout:'%v'", socketRequest.Command, socketRequest.Timeout)
		connection := deviceShifuSocketHandlerMetaData.connection
		command := socketRequest.Command
		timeout := socketRequest.Timeout
		if timeout != 0 {
			(*connection).SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
		}

		log.Printf("Sending %v", []byte(command+"\n"))
		(*connection).Write([]byte(command + "\n"))

		var message []byte
		if end, exists := params["end"]; exists {
			message, err = bufio.NewReader(*connection).ReadBytes(end[0])
		} else {
			message, err = bufio.NewReader(*connection).ReadBytes('\n')
		}

		if err != nil {
			log.Printf("Failed to ReadString from Socket, current message: %v", message)
			http.Error(w, "Failed to read message from socket, error: "+err.Error(), http.StatusBadRequest)
		}

		returnMessage := DeviceShifuSocketReturnBody{
			Message: string(message),
			Status:  http.StatusOK,
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(returnMessage)
	}
}

func (ds *DeviceShifu) startHttpServer(stopCh <-chan struct{}) error {
	fmt.Printf("deviceShifu %s's http server started\n", ds.base.Name)
	return ds.base.Server.ListenAndServe()
}

// TODO: update configs
// TODO: update status based on telemetry

func (ds *DeviceShifu) collectSocketTelemetry() (bool, error) {
	if ds.base.EdgeDevice.Spec.Address == nil {
		return false, fmt.Errorf("Device %v does not have an address", ds.base.Name)
	}

	if ds.base.EdgeDevice.Spec.Protocol != nil {
		switch protocol := *ds.base.EdgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolSocket:
			conn, err := net.Dial("tcp", *ds.base.EdgeDevice.Spec.Address)
			if err != nil {
				log.Printf("error checking telemetry: error: %v", err.Error())
				return false, err
			}

			defer conn.Close()
			return true, nil
		default:
			log.Printf("EdgeDevice protocol %v not supported in deviceShifu\n", protocol)
			return false, nil
		}
	}
	return true, nil
}

func (ds *DeviceShifu) Start(stopCh <-chan struct{}) error {
	return ds.base.Start(stopCh, ds.collectSocketTelemetry)
}

func (ds *DeviceShifu) Stop() error {
	return ds.base.Stop()
}
