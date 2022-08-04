package deviceshifuSocket

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/edgenesis/shifu/deviceshifu/pkg/deviceshifubase"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"edgenesis.io/shifu/k8s/crd/api/v1alpha1"
	"k8s.io/client-go/rest"
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
	if deviceShifuMetadata.Name == "" {
		return nil, fmt.Errorf("DeviceShifu's name can't be empty\n")
	}

	if deviceShifuMetadata.ConfigFilePath == "" {
		deviceShifuMetadata.ConfigFilePath = deviceshifubase.DEVICE_CONFIGMAP_FOLDER_PATH
	}

	deviceShifuConfig, err := deviceshifubase.NewDeviceShifuConfig(deviceShifuMetadata.ConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf("Error parsing ConfigMap at %v\n", deviceShifuMetadata.ConfigFilePath)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", deviceHealthHandler)
	mux.HandleFunc("/", instructionNotFoundHandler)

	edgeDevice := &v1alpha1.EdgeDevice{}
	client := &rest.RESTClient{}
	var socketConnection net.Conn

	if deviceShifuMetadata.KubeConfigPath != deviceshifubase.DEVICE_KUBECONFIG_DO_NOT_LOAD_STR {
		edgeDeviceConfig := &deviceshifubase.EdgeDeviceConfig{
			NameSpace:      deviceShifuMetadata.Namespace,
			DeviceName:     deviceShifuMetadata.Name,
			KubeconfigPath: deviceShifuMetadata.KubeConfigPath,
		}

		edgeDevice, client, err = deviceshifubase.NewEdgeDevice(edgeDeviceConfig)
		if err != nil {
			log.Fatalf("Error retrieving EdgeDevice")
			return nil, err
		}

		if &edgeDevice.Spec == nil {
			log.Fatalf("edgeDeviceConfig.Spec is nil")
			return nil, err
		}

		switch protocol := *edgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolSocket:
			// Open the connection:
			connectionType := edgeDevice.Spec.ProtocolSettings.SocketSetting.NetworkType
			encoding := edgeDevice.Spec.ProtocolSettings.SocketSetting.Encoding
			if connectionType == nil || *connectionType != "tcp" {
				return nil, fmt.Errorf("Sorry!, Shifu currently only support TCP Socket")
			}

			if encoding == nil {
				log.Println("Socket encoding not specified, default to UTF-8")
				return nil, fmt.Errorf("Encoding error")
			}

			socketConnection, err := net.Dial(*connectionType, *edgeDevice.Spec.Address)
			if err != nil {
				return nil, fmt.Errorf("Cannot connect to %v", *edgeDevice.Spec.Address)
			}

			log.Printf("Connected to '%v'\n", *edgeDevice.Spec.Address)
			for instruction, properties := range deviceShifuConfig.Instructions.Instructions {
				deviceShifuSocketHandlerMetaData := &DeviceShifuSocketHandlerMetaData{
					edgeDevice.Spec,
					instruction,
					properties,
					&socketConnection,
				}

				mux.HandleFunc("/"+instruction, deviceCommandHandlerSocket(deviceShifuSocketHandlerMetaData))
			}
		}
	}

	dsbase := &deviceshifubase.DeviceShifuBase{
		Name: deviceShifuMetadata.Name,
		Server: &http.Server{
			Addr:         deviceshifubase.DEVICE_DEFAULT_PORT_STR,
			Handler:      mux,
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 60 * time.Second,
		},
		DeviceShifuConfig: deviceShifuConfig,
		EdgeDevice:        edgeDevice,
		RestClient:        client,
	}

	ds := &DeviceShifu{base: dsbase, socketConnection: &socketConnection}
	ds.base.UpdateEdgeDeviceResourcePhase(v1alpha1.EdgeDevicePending)
	return ds, nil
}

func deviceHealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, deviceshifubase.DEVICE_IS_HEALTHY_STR)
}

func instructionNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Error: Device instruction does not exist!")
	http.Error(w, "Error: Device instruction does not exist!", http.StatusNotFound)
}

func createUriFromRequest(address string, handlerInstruction string, r *http.Request) string {

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

// HTTP header type:
// type Header map[string][]string
func copyHeader(dst, src http.Header) {
	for header, headerValueList := range src {
		for _, value := range headerValueList {
			dst.Add(header, value)
		}
	}
}

func deviceCommandHandlerSocket(deviceShifuSocketHandlerMetaData *DeviceShifuSocketHandlerMetaData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		message, err := bufio.NewReader(*connection).ReadString('\n')
		if err != nil {
			log.Printf("Failed to ReadString from Socket, current message: %v", message)
			http.Error(w, "Failed to read message from socket, error: "+err.Error(), http.StatusBadRequest)
		}

		returnMessage := DeviceShifuSocketReturnBody{
			Message: message,
			Status:  http.StatusOK,
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(returnMessage)
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
