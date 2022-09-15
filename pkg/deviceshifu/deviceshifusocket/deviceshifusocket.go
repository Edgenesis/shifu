package deviceshifusocket

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"k8s.io/klog/v2"
)

// DeviceShifu deviceshifu and socketConnection for Socket
type DeviceShifu struct {
	base             *deviceshifubase.DeviceShifuBase
	socketConnection *net.Conn
}

// HandlerMetaData MetaData for Socket Handler
type HandlerMetaData struct {
	edgeDeviceSpec v1alpha1.EdgeDeviceSpec
	instruction    string
	properties     *deviceshifubase.DeviceShifuInstruction
	connection     *net.Conn
}

// New new socket deviceshifu
func New(deviceShifuMetadata *deviceshifubase.DeviceShifuMetaData) (*DeviceShifu, error) {
	base, mux, err := deviceshifubase.New(deviceShifuMetadata)
	if err != nil {
		return nil, err
	}
	var socketConnection net.Conn

	if deviceShifuMetadata.KubeConfigPath != deviceshifubase.DeviceKubeconfigDoNotLoadStr {
		switch protocol := *base.EdgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolSocket:
			// Open the connection:
			connectionType := base.EdgeDevice.Spec.ProtocolSettings.SocketSetting.NetworkType
			encoding := base.EdgeDevice.Spec.ProtocolSettings.SocketSetting.Encoding
			if connectionType == nil || *connectionType != "tcp" {
				return nil, fmt.Errorf("Sorry!, Shifu currently only support TCP Socket")
			}

			if encoding == nil {
				klog.Errorf("Socket encoding not specified, default to UTF-8")
				return nil, fmt.Errorf("Encoding error")
			}

			socketConnection, err := net.Dial(*connectionType, *base.EdgeDevice.Spec.Address)
			if err != nil {
				return nil, fmt.Errorf("Cannot connect to %v", *base.EdgeDevice.Spec.Address)
			}

			klog.Infof("Connected to '%v'", *base.EdgeDevice.Spec.Address)
			for instruction, properties := range base.DeviceShifuConfig.Instructions.Instructions {
				HandlerMetaData := &HandlerMetaData{
					base.EdgeDevice.Spec,
					instruction,
					properties,
					&socketConnection,
				}

				mux.HandleFunc("/"+instruction, deviceCommandHandlerSocket(HandlerMetaData))
			}
		}
	}
	deviceshifubase.BindDefaultHandler(mux)

	ds := &DeviceShifu{base: base, socketConnection: &socketConnection}
	ds.base.UpdateEdgeDeviceResourcePhase(v1alpha1.EdgeDevicePending)
	return ds, nil
}

func createURIFromRequest(address string, handlerInstruction string, r *http.Request) string {

	queryStr := "?"

	for queryName, queryValues := range r.URL.Query() {
		for _, queryValue := range queryValues {
			queryStr += queryName + "=" + queryValue + "&"
		}
	}

	queryStr = strings.TrimSuffix(queryStr, "&")

	if queryStr == "?" {
		return "http://" + address + "/" + handlerInstruction
	}

	return "http://" + address + "/" + handlerInstruction + queryStr
}

func deviceCommandHandlerSocket(HandlerMetaData *HandlerMetaData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		headerContentType := r.Header.Get("Content-Type")
		if headerContentType != "application/json" {
			http.Error(w, "content-type is not application/json", http.StatusBadRequest)
			klog.Errorf("content-type is not application/json")
			return
		}

		var socketRequest RequestBody
		err := json.NewDecoder(r.Body).Decode(&socketRequest)
		if err != nil {
			klog.Errorf("error decode: %v", socketRequest)
			http.Error(w, "error decode JSON "+err.Error(), http.StatusBadRequest)
			return
		}

		klog.Infof("After decode socket request: '%v', timeout:'%v'", socketRequest.Command, socketRequest.Timeout)
		connection := HandlerMetaData.connection
		command := socketRequest.Command
		timeout := socketRequest.Timeout
		if timeout > 0 {
			err := (*connection).SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
			if err != nil {
				klog.Errorf("cannot send deadline to socket, error: %v", err)
				http.Error(w, "Failed to send deadline to socket, error:  "+err.Error(), http.StatusBadRequest)
				return
			}
		}

		klog.Infof("Sending %v", []byte(command+"\n"))
		_, err = (*connection).Write([]byte(command + "\n"))
		if err != nil {
			klog.Errorf("cannot write command into socket, error: %v", err)
			http.Error(w, "Failed to send message to socket, error: "+err.Error(), http.StatusBadRequest)
			return
		}

		message, err := bufio.NewReader(*connection).ReadString('\n')
		if err != nil {
			klog.Errorf("Failed to ReadString from Socket, current message: %v", message)
			http.Error(w, "Failed to read message from socket, error: "+err.Error(), http.StatusBadRequest)
			return
		}

		returnMessage := ReturnBody{
			Message: message,
			Status:  http.StatusOK,
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(returnMessage)
		if err != nil {
			klog.Errorf("Failed encode message to json, error: %v" + err.Error())
			http.Error(w, "Failed encode message to json, error: "+err.Error(), http.StatusBadRequest)
		}
	}
}

// this function gathers the instruction name and its arguments from user input via HTTP and create the direct call command
// "flags_no_parameter" is a special key where it contains all flags
// e.g.:
// if we have localhost:8081/start?time=10:00:00&flags_no_parameter=-a,-c,--no-dependency&target=machine2
// and our driverExecution is "/usr/local/bin/python /usr/src/driver/python-car-driver.py"
// then we will get this command string:
// /usr/local/bin/python /usr/src/driver/python-car-driver.py --start time=10:00:00 target=machine2 -a -c --no-dependency
// which is exactly what we need to run if we are operating directly on the device
func createHTTPCommandlineRequestString(r *http.Request, driverExecution string, instruction string) string {
	values := r.URL.Query()
	requestStr := ""
	flagsStr := ""
	for parameterName, parameterValues := range values {
		if parameterName == "flags_no_parameter" {
			if len(parameterValues) == 1 {
				flagsStr = " " + strings.Replace(parameterValues[0], ",", " ", -1)
			} else {
				for _, parameterValue := range parameterValues {
					flagsStr += " " + parameterValue
				}
			}
		} else {
			if len(parameterValues) < 1 {
				continue
			}

			requestStr += " " + parameterName + "="
			for _, parameterValue := range parameterValues {
				requestStr += parameterValue
			}
		}
	}
	return driverExecution + " --" + instruction + requestStr + flagsStr
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
				klog.Errorf("error checking telemetry: error: %v", err.Error())
				return false, err
			}

			defer conn.Close()
			return true, nil
		default:
			klog.Warningf("EdgeDevice protocol %v not supported in deviceshifu", protocol)
			return false, nil
		}
	}
	return true, nil
}

// Start start socket telemetry
func (ds *DeviceShifu) Start(stopCh <-chan struct{}) error {
	return ds.base.Start(stopCh, ds.collectSocketTelemetry)
}

// Stop stop http server
func (ds *DeviceShifu) Stop() error {
	return ds.base.Stop()
}
