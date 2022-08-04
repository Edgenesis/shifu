package deviceshifuOPCUA

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/edgenesis/shifu/deviceshifu/pkg/deviceshifubase"
	"log"
	"net/http"
	"path"
	"time"

	"edgenesis.io/shifu/k8s/crd/api/v1alpha1"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

type DeviceShifu struct {
	base              *deviceshifubase.DeviceShifuBase
	opcuaInstructions *OPCUAInstructions
	opcuaClient       *opcua.Client
}

type DeviceShifuOPCUAHandlerMetaData struct {
	edgeDeviceSpec v1alpha1.EdgeDeviceSpec
	instruction    string
	properties     *OPCUAInstructionProperty
}

type deviceCommandHandler interface {
	commandHandleFunc(w http.ResponseWriter, r *http.Request) http.HandlerFunc
}

const (
	DEVICE_CONFIGMAP_CERTIFICATE_PATH string = "/etc/edgedevice/certificate"
	EDGEDEVICE_STATUS_FAIL            bool   = false
)

// This function creates a new Device Shifu based on the configuration
func New(deviceShifuMetadata *deviceshifubase.DeviceShifuMetaData) (*DeviceShifu, error) {
	if deviceShifuMetadata.Namespace == "" {
		return nil, fmt.Errorf("DeviceShifu's namespace can't be empty\n")
	}

	base, mux, err := deviceshifubase.New(deviceShifuMetadata)
	if err != nil {
		return nil, err
	}

	ocupaInstructions, err := NewOPCUAInstructions(deviceShifuMetadata.ConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf("Error parsing ConfigMap at %v\n", deviceShifuMetadata.ConfigFilePath)
	}
	var opcuaClient *opcua.Client

	if deviceShifuMetadata.KubeConfigPath != deviceshifubase.DEVICE_KUBECONFIG_DO_NOT_LOAD_STR {
		// switch for different Shifu Protocols
		switch protocol := *base.EdgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolOPCUA:
			for instruction, properties := range ocupaInstructions.Instructions {
				deviceShifuOPCUAHandlerMetaData := &DeviceShifuOPCUAHandlerMetaData{
					base.EdgeDevice.Spec,
					instruction,
					properties.OPCUAInstructionProperties,
				}

				ctx := context.Background()
				endpoints, err := opcua.GetEndpoints(ctx, *base.EdgeDevice.Spec.Address)
				if err != nil {
					log.Fatal("Cannot Get EndPoint Description")
					return nil, err
				}

				ep := opcua.SelectEndpoint(endpoints, ua.SecurityPolicyURINone, ua.MessageSecurityModeNone)
				if ep == nil {
					log.Fatal("Failed to find suitable endpoint")
				}

				var options = make([]opcua.Option, 0)
				// TODO  implement different messageSecurityModes
				options = append(options,
					opcua.SecurityPolicy(ua.SecurityPolicyURINone),
					opcua.SecurityMode(ua.MessageSecurityModeNone),
				)

				var setting = *base.EdgeDevice.Spec.ProtocolSettings.OPCUASetting
				switch ua.UserTokenTypeFromString(*setting.AuthenticationMode) {
				case ua.UserTokenTypeIssuedToken:
					options = append(options, opcua.AuthIssuedToken([]byte(*setting.IssuedToken)))
				case ua.UserTokenTypeCertificate:
					var privateKeyFileName = path.Join(DEVICE_CONFIGMAP_CERTIFICATE_PATH, *setting.PrivateKeyFileName)
					var certificateFileName = path.Join(DEVICE_CONFIGMAP_CERTIFICATE_PATH, *setting.CertificateFileName)
					cert, err := tls.LoadX509KeyPair(certificateFileName, privateKeyFileName)
					if err != nil {
						log.Fatalf("X509 Certificate Or PrivateKey load Default")
					}
					options = append(options,
						opcua.CertificateFile(certificateFileName),
						opcua.PrivateKeyFile(privateKeyFileName),
						opcua.AuthCertificate(cert.Certificate[0]),
					)
				case ua.UserTokenTypeUserName:
					options = append(options, opcua.AuthUsername(*setting.Username, *setting.Password))
				case ua.UserTokenTypeAnonymous:
					fallthrough
				default:
					if *setting.AuthenticationMode != "Anonymous" {
						log.Println("Could not parse your input, you are in Anonymous Mode default")
					}

					options = append(options, opcua.AuthAnonymous())
				}

				options = append(options, opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeFromString(*setting.AuthenticationMode)))
				opcuaClient = opcua.NewClient(*base.EdgeDevice.Spec.Address, options...)
				if err := opcuaClient.Connect(ctx); err != nil {
					log.Fatalf("Unable to connect to OPC UA server, error: %v", err)
				}

				var handler DeviceCommandHandlerOPCUA
				if base.EdgeDevice.Spec.ProtocolSettings.OPCUASetting.ConnectionTimeoutInMilliseconds == nil {
					timeout := deviceshifubase.DEVICE_DEFAULT_CONNECTION_TIMEOUT_MS
					handler = DeviceCommandHandlerOPCUA{opcuaClient, &timeout, deviceShifuOPCUAHandlerMetaData}
				} else {
					timeout := base.EdgeDevice.Spec.ProtocolSettings.OPCUASetting.ConnectionTimeoutInMilliseconds
					handler = DeviceCommandHandlerOPCUA{opcuaClient, timeout, deviceShifuOPCUAHandlerMetaData}
				}

				mux.HandleFunc("/"+instruction, handler.commandHandleFunc())
			}
		}
	}

	ds := &DeviceShifu{
		base:              base,
		opcuaInstructions: ocupaInstructions,
		opcuaClient:       opcuaClient,
	}

	ds.base.UpdateEdgeDeviceResourcePhase(v1alpha1.EdgeDevicePending)
	return ds, nil
}

type DeviceCommandHandlerOPCUA struct {
	client                          *opcua.Client
	timeout                         *int64
	deviceShifuOPCUAHandlerMetaData *DeviceShifuOPCUAHandlerMetaData
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
	fmt.Printf("deviceShifu %s's http server started\n", ds.base.Name)
	return ds.base.Server.ListenAndServe()
}

func (ds *DeviceShifu) getOPCUANodeIDFromInstructionName(instructionName string) (string, error) {
	if instructionProperties, exists := ds.opcuaInstructions.Instructions[instructionName]; exists {
		return instructionProperties.OPCUAInstructionProperties.OPCUANodeID, nil
	}

	return "", fmt.Errorf("Instruction %v not found in list of deviceShifu instructions", instructionName)
}

func (ds *DeviceShifu) requestOPCUANodeID(nodeID string) error {
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
	ctx, cancel := context.WithTimeout(ctx, time.Duration(deviceshifubase.DEVICE_DEFAULT_REQUEST_TIMEOUT_MS)*time.Millisecond)
	defer cancel()

	resp, err := ds.opcuaClient.ReadWithContext(ctx, req)
	if err != nil {
		log.Printf("Failed to read message from Server, error: %v " + err.Error())
		return err
	}

	if resp.Results[0].Status != ua.StatusOK {
		log.Printf("OPC UA response status is not OK, status: %v", resp.Results[0].Status)
		return err
	}

	log.Printf(fmt.Sprint(resp.Results[0].Value.Value()))

	return nil
}

func (ds *DeviceShifu) collectOPCUATelemetry() (bool, error) {
	if ds.base.EdgeDevice.Spec.Protocol != nil {
		switch protocol := *ds.base.EdgeDevice.Spec.Protocol; protocol {
		case v1alpha1.ProtocolOPCUA:
			telemetries := ds.base.DeviceShifuConfig.Telemetries.DeviceShifuTelemetries
			for telemetry, telemetryProperties := range telemetries {
				if ds.base.EdgeDevice.Spec.Address == nil {
					return false, fmt.Errorf("Device %v does not have an address", ds.base.Name)
				}

				if telemetryProperties.DeviceShifuTelemetryProperties.DeviceInstructionName == nil {
					return false, fmt.Errorf("Device %v telemetry %v does not have an instruction name", ds.base.Name, telemetry)
				}

				instruction := *telemetryProperties.DeviceShifuTelemetryProperties.DeviceInstructionName
				nodeID, err := ds.getOPCUANodeIDFromInstructionName(instruction)
				if err != nil {
					log.Printf(err.Error())
					return false, err
				}

				if err = ds.requestOPCUANodeID(nodeID); err != nil {
					log.Printf("error checking telemetry: %v, error: %v", telemetry, err.Error())
					return false, err
				}

			}
		default:
			log.Printf("EdgeDevice protocol %v not supported in deviceShifu\n", protocol)
			return false, nil
		}
	}

	return true, nil

}

func (ds *DeviceShifu) Start(stopCh <-chan struct{}) error {
	return ds.base.Start(stopCh, ds.collectOPCUATelemetry)
}

func (ds *DeviceShifu) Stop() error {
	return ds.base.Stop()
}
