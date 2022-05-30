package deviceshifuOPCUA

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	v1alpha1 "edgenesis.io/shifu/k8s/crd/api/v1alpha1"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
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

type DeviceShifuOPCUAHandlerMetaData struct {
	edgeDeviceSpec v1alpha1.EdgeDeviceSpec
	instruction    string
	properties     *DeviceShifuInstructionProperty
}

type deviceCommandHandler interface {
	commandHandleFunc(w http.ResponseWriter, r *http.Request) http.HandlerFunc
}

const (
	DEVICE_IS_HEALTHY_STR                string = "Device is healthy"
	DEVICE_CONFIGMAP_FOLDER_PATH         string = "/etc/edgedevice/config"
	DEVICE_KUBECONFIG_DO_NOT_LOAD_STR    string = "NULL"
	DEVICE_NAMESPACE_DEFAULT             string = "default"
	DEVICE_DEFAULT_CONNECTION_TIMEOUT_MS int64  = 3000
	DEVICE_DEFAULT_PORT_STR              string = ":8080"
	KUBERNETES_CONFIG_DEFAULT            string = ""
)

// This function creates a new Device Shifu based on the configuration
func New(deviceShifuMetadata *DeviceShifuMetaData) (*DeviceShifu, error) {
	if deviceShifuMetadata.Name == "" {
		return nil, fmt.Errorf("DeviceShifu's name can't be empty\n")
	}

	if deviceShifuMetadata.Namespace == "" {
		return nil, fmt.Errorf("DeviceShifu's namespace can't be empty\n")
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

		// switch for different Shifu Protocols
		switch protocol := *edgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolOPCUA:
			for instruction, properties := range deviceShifuConfig.Instructions {
				deviceShifuOPCUAHandlerMetaData := &DeviceShifuOPCUAHandlerMetaData{
					edgeDevice.Spec,
					instruction,
					properties.DeviceShifuInstructionProperties,
				}

				ctx := context.Background()
				// TODO: handle different security modes
				client := opcua.NewClient(*edgeDevice.Spec.Address,
					opcua.SecurityMode(ua.MessageSecurityModeNone))
				if err := client.Connect(ctx); err != nil {
					log.Fatalf("Unable to connect to OPC UA server, error: %v", err)
				}

				var handler DeviceCommandHandlerOPCUA
				if edgeDevice.Spec.ProtocolSettings.OPCUASetting.ConnectionTimeoutInMilliseconds == nil {
					timeout := DEVICE_DEFAULT_CONNECTION_TIMEOUT_MS
					handler = DeviceCommandHandlerOPCUA{client, &timeout, deviceShifuOPCUAHandlerMetaData}
				} else {
					timeout := edgeDevice.Spec.ProtocolSettings.OPCUASetting.ConnectionTimeoutInMilliseconds
					handler = DeviceCommandHandlerOPCUA{client, timeout, deviceShifuOPCUAHandlerMetaData}
				}

				mux.HandleFunc("/"+instruction, handler.commandHandleFunc())
			}
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

// deviceHealthHandler writes the status as healthy
func deviceHealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, DEVICE_IS_HEALTHY_STR)
}

type DeviceCommandHandlerOPCUA struct {
	client                          *opcua.Client
	timeout                         *int64
	deviceShifuOPCUAHandlerMetaData *DeviceShifuOPCUAHandlerMetaData
}

func instructionNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Error: Device instruction does not exist!")
	http.Error(w, "Error: Device instruction does not exist!", http.StatusNotFound)
}

func (handler DeviceCommandHandlerOPCUA) commandHandleFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nodeID := handler.deviceShifuOPCUAHandlerMetaData.properties.OPCUANodeID
		log.Printf("Requesting NodeID: %v", nodeID)

		id, err := ua.ParseNodeID(nodeID)
		if err != nil {
			log.Fatalf("invalid node id: %v", err)
		}

		req := &ua.ReadRequest{
			MaxAge: 2000,
			NodesToRead: []*ua.ReadValueID{
				{NodeID: id},
			},
			TimestampsToReturn: ua.TimestampsToReturnBoth,
		}

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Duration(*handler.timeout)*time.Millisecond)
		defer cancel()

		resp, err := handler.client.ReadWithContext(ctx, req)
		if err != nil {
			http.Error(w, "Failed to read message from Server, error: "+err.Error(), http.StatusBadRequest)
			log.Printf("Read failed: %s", err)
			return
		}

		if resp.Results[0].Status != ua.StatusOK {
			http.Error(w, "OPC UA response status is not OK "+fmt.Sprint(resp.Results[0].Status), http.StatusBadRequest)
			log.Printf("Status not OK: %v", resp.Results[0].Status)
			return
		}

		log.Printf("%#v", resp.Results[0].Value.Value())

		w.WriteHeader(http.StatusOK)
		// TODO: Should handle different type of return values and return JSON/other data
		// types instead of plain text
		fmt.Fprintf(w, "%v", resp.Results[0].Value.Value())
	}
}

func (ds *DeviceShifu) startHttpServer(stopCh <-chan struct{}) error {
	fmt.Printf("deviceShifu %s's http server started\n", ds.Name)
	return ds.server.ListenAndServe()
}

func (ds *DeviceShifu) telemetryCollection() error {
	// TODO: handle interval for different telemetries
	log.Printf("deviceShifu %s's telemetry collection started\n", ds.Name)

	if ds.edgeDevice.Spec.Protocol != nil {
		switch protocol := *ds.edgeDevice.Spec.Protocol; protocol {
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
