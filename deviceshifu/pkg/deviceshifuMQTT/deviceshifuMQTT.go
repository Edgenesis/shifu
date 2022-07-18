package deviceshifuMQTT

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	v1alpha1 "edgenesis.io/shifu/k8s/crd/api/v1alpha1"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/client-go/rest"
)

type DeviceShifu struct {
	Name              string
	server            *http.Server
	deviceShifuConfig *DeviceShifuConfig
	edgeDevice        *v1alpha1.EdgeDevice
	restClient        *rest.RESTClient
}

type DeviceShifuMetaData struct {
	Name           string
	ConfigFilePath string
	KubeConfigPath string
	Namespace      string
}

type DeviceShifuMQTTHandlerMetaData struct {
	edgeDeviceSpec v1alpha1.EdgeDeviceSpec
	// instruction    string
	// properties     *DeviceShifuInstruction
}

const (
	DEVICE_IS_HEALTHY_STR             string = "Device is healthy"
	DEVICE_CONFIGMAP_FOLDER_PATH      string = "/etc/edgedevice/config"
	DEVICE_KUBECONFIG_DO_NOT_LOAD_STR string = "NULL"
	DEVICE_NAMESPACE_DEFAULT          string = "default"
	DEVICE_DEFAULT_PORT_STR           string = ":8080"
	KUBERNETES_CONFIG_DEFAULT         string = ""
	MQTT_DATA_ENDPOINT                string = "mqtt_data"
)

var (
	MQTT_MESSAGE_STR               string
	MQTT_MESSAGE_RECEIVE_TIMESTAMP time.Time
)

func New(deviceShifuMetadata *DeviceShifuMetaData) (*DeviceShifu, error) {
	if deviceShifuMetadata.Name == "" {
		return nil, fmt.Errorf("DeviceShifu's name can't be empty\n")
	}

	if deviceShifuMetadata.ConfigFilePath == "" {
		deviceShifuMetadata.ConfigFilePath = DEVICE_CONFIGMAP_FOLDER_PATH
	}

	deviceShifuConfig, err := NewDeviceShifuConfig(deviceShifuMetadata.ConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf("Error parsing ConfigMap at %v\n", deviceShifuMetadata.ConfigFilePath)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", deviceHealthHandler)
	mux.HandleFunc("/", instructionNotFoundHandler)

	edgeDevice := &v1alpha1.EdgeDevice{}
	client := &rest.RESTClient{}

	if deviceShifuMetadata.KubeConfigPath != DEVICE_KUBECONFIG_DO_NOT_LOAD_STR {
		edgeDeviceConfig := &EdgeDeviceConfig{
			deviceShifuMetadata.Namespace,
			deviceShifuMetadata.Name,
			deviceShifuMetadata.KubeConfigPath,
		}

		edgeDevice, client, err = NewEdgeDevice(edgeDeviceConfig)
		if err != nil {
			log.Fatalf("Error retrieving EdgeDevice")
			return nil, err
		}

		if &edgeDevice.Spec == nil {
			log.Fatalf("edgeDeviceConfig.Spec is nil")
			return nil, err
		}

		switch protocol := *edgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolMQTT:
			mqttSetting := *edgeDevice.Spec.ProtocolSettings.MQTTSetting
			var mqttServerAddress string
			if mqttSetting.MQTTTopic == nil || *mqttSetting.MQTTTopic == "" {
				return nil, fmt.Errorf("MQTT Topic cannot be empty")
			}

			if mqttSetting.MQTTServerAddress == nil || *mqttSetting.MQTTServerAddress == "" {
				// return nil, fmt.Errorf("MQTT server cannot be empty")
				log.Println("MQTT Server Address is empty, use address instead")
				mqttServerAddress = *edgeDevice.Spec.Address
			} else {
				mqttServerAddress = *mqttSetting.MQTTServerAddress
			}

			opts := mqtt.NewClientOptions()
			opts.AddBroker(fmt.Sprintf("tcp://%s", mqttServerAddress))
			opts.SetClientID(*&edgeDeviceConfig.deviceName)
			opts.SetDefaultPublishHandler(messagePubHandler)
			opts.OnConnect = connectHandler
			opts.OnConnectionLost = connectLostHandler
			client := mqtt.NewClient(opts)
			if token := client.Connect(); token.Wait() && token.Error() != nil {
				panic(token.Error())
			}

			sub(client, *mqttSetting.MQTTTopic)

			deviceShifuMQTTHandlerMetaData := &DeviceShifuMQTTHandlerMetaData{
				edgeDevice.Spec,
			}

			handler := DeviceCommandHandlerMQTT{deviceShifuMQTTHandlerMetaData}
			mux.HandleFunc("/"+MQTT_DATA_ENDPOINT, handler.commandHandleFunc())
		}
	}

	ds := &DeviceShifu{
		Name: deviceShifuMetadata.Name,
		server: &http.Server{
			Addr:         DEVICE_DEFAULT_PORT_STR,
			Handler:      mux,
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 60 * time.Second,
		},
		deviceShifuConfig: deviceShifuConfig,
		edgeDevice:        edgeDevice,
		restClient:        client,
	}

	ds.updateEdgeDeviceResourcePhase(v1alpha1.EdgeDevicePending)
	return ds, nil
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message: %v from topic: %v\n", msg.Payload(), msg.Topic())
	MQTT_MESSAGE_STR = string(msg.Payload())
	MQTT_MESSAGE_RECEIVE_TIMESTAMP = time.Now()
	log.Print("MESSAGE_STR updated")
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connect lost: %v", err)
}

func sub(client mqtt.Client, topic string) {
	// topic := "topic/test"
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	log.Printf("Subscribed to topic: %s", topic)
}

func deviceHealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, DEVICE_IS_HEALTHY_STR)
}

type DeviceCommandHandlerMQTT struct {
	// client                         *rest.RESTClient
	deviceShifuMQTTHandlerMetaData *DeviceShifuMQTTHandlerMetaData
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

func (handler DeviceCommandHandlerMQTT) commandHandleFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// handlerEdgeDeviceSpec := handler.deviceShifuMQTTHandlerMetaData.edgeDeviceSpec
		reqType := r.Method

		if reqType == http.MethodGet {
			returnMessage := DeviceShifuMQTTReturnBody{
				MQTTMessage:   MQTT_MESSAGE_STR,
				MQTTTimestamp: MQTT_MESSAGE_RECEIVE_TIMESTAMP.String(),
			}

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(returnMessage)
		} else {
			http.Error(w, "must be GET method", http.StatusBadRequest)
			log.Println("Request type " + reqType + " is not supported yet!")
			return
		}

	}
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
	fmt.Printf("deviceShifu %s's http server started\n", ds.Name)
	return ds.server.ListenAndServe()
}

// TODO: update configs
// TODO: update status based on telemetry

func (ds *DeviceShifu) collectMQTTTelemetry(telemetry string, settings DeviceShifuTelemetrySettings) (bool, error) {
	if ds.edgeDevice.Spec.Address == nil {
		return false, fmt.Errorf("Device %v does not have an address", ds.Name)
	}

	if settings.DeviceShifuTelemetryUpdateIntervalMiliseconds == nil {
		return false, fmt.Errorf("Device %v telemetry %v does not have an instruction name", ds.Name, telemetry)
	}

	nowTime := time.Now()

	if nowTime.Sub(MQTT_MESSAGE_RECEIVE_TIMESTAMP).Seconds() < float64(*settings.DeviceShifuTelemetryUpdateIntervalMiliseconds) {
		return true, nil
	}

	return false, nil
}

func (ds *DeviceShifu) collectMQTTTelemetries() error {
	telemetryOK := true
	telemetries := ds.deviceShifuConfig.Telemetries.DeviceShifuTelemetries
	telemetriesSettings := ds.deviceShifuConfig.Telemetries.DeviceShifuTelemetrySettings
	for telemetry := range telemetries {
		status, err := ds.collectMQTTTelemetry(telemetry, *telemetriesSettings)
		log.Printf("Status is: %v", status)
		if err != nil {
			log.Printf("Error is: %v", err.Error())
			telemetryOK = false
		}

		if !status && telemetryOK {
			telemetryOK = false
		}
	}

	if telemetryOK {
		ds.updateEdgeDeviceResourcePhase(v1alpha1.EdgeDeviceRunning)
	} else {
		ds.updateEdgeDeviceResourcePhase(v1alpha1.EdgeDeviceFailed)
	}

	return nil
}

func (ds *DeviceShifu) telemetryCollection() error {
	// TODO: handle interval for different telemetries
	log.Printf("deviceShifu %s's telemetry collection started\n", ds.Name)

	if ds.edgeDevice.Spec.Protocol != nil {
		switch protocol := *ds.edgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolHTTP:
			ds.collectMQTTTelemetries()
		default:
			log.Printf("EdgeDevice protocol %v not supported in deviceShifu\n", protocol)
			ds.updateEdgeDeviceResourcePhase(v1alpha1.EdgeDeviceFailed)
		}

		return nil
	}

	return fmt.Errorf("EdgeDevice %v has no telemetry field in configuration\n", ds.Name)
}

func (ds *DeviceShifu) updateEdgeDeviceResourcePhase(edPhase v1alpha1.EdgeDevicePhase) {
	log.Printf("updating device %v status to: %v\n", ds.Name, edPhase)
	currEdgeDevice := &v1alpha1.EdgeDevice{}
	err := ds.restClient.Get().
		Namespace(ds.edgeDevice.Namespace).
		Resource(EDGEDEVICE_RESOURCE_STR).
		Name(ds.Name).
		Do(context.TODO()).
		Into(currEdgeDevice)

	if err != nil {
		log.Printf("Unable to update status, error: %v", err.Error())
		return
	}

	if currEdgeDevice.Status.EdgeDevicePhase == nil {
		edgeDeviceStatus := v1alpha1.EdgeDevicePending
		currEdgeDevice.Status.EdgeDevicePhase = &edgeDeviceStatus
	} else {
		*currEdgeDevice.Status.EdgeDevicePhase = edPhase
	}

	putResult := &v1alpha1.EdgeDevice{}
	err = ds.restClient.Put().
		Namespace(ds.edgeDevice.Namespace).
		Resource(EDGEDEVICE_RESOURCE_STR).
		Name(ds.Name).
		Body(currEdgeDevice).
		Do(context.TODO()).
		Into(putResult)

	if err != nil {
		log.Printf("Unable to update status, error: %v", err)
	}
}

func (ds *DeviceShifu) StartTelemetryCollection() error {
	log.Println("Wait 5 seconds before updating status")
	time.Sleep(5 * time.Second)
	for {
		ds.telemetryCollection()
		time.Sleep(5 * time.Second)
	}
}

func (ds *DeviceShifu) Start(stopCh <-chan struct{}) error {
	fmt.Printf("deviceShifu %s started\n", ds.Name)

	go ds.startHttpServer(stopCh)
	go ds.StartTelemetryCollection()

	return nil
}

func (ds *DeviceShifu) Stop() error {
	if err := ds.server.Shutdown(context.TODO()); err != nil {
		return err
	}

	fmt.Printf("deviceShifu %s's http server stopped\n", ds.Name)
	return nil
}
