package lwm2m

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	lwm2mclient "github.com/edgenesis/shifu/pkg/gateway/lwm2m/client"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/edgenesis/shifu/pkg/logger"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/rest"
	"knative.dev/pkg/configmap"
)

const (
	deviceShifuHost          = "http://localhost:8080"
	ConfigmapFolderPath      = "/etc/edgedevice/config"
	ConfigmapInstructionsStr = "instructions"
	ObjectIdStr              = "ObjectId"
	DataTypeStr              = "DataType"
)

type Gateway struct {
	client     *lwm2mclient.Client
	k8sClient  *rest.RESTClient
	edgeDevice *v1alpha1.EdgeDevice

	pingIntervalSec int64
}

func New() (*Gateway, error) {
	edgedevice, krclient, err := deviceshifubase.NewEdgeDevice(&deviceshifubase.EdgeDeviceConfig{
		NameSpace:  os.Getenv("EDGEDEVICE_NAMESPACE"),
		DeviceName: os.Getenv("EDGEDEVICE_NAME"),
	})
	if err != nil {
		return nil, err
	}

	if edgedevice.Spec.GatewaySettings == nil {
		return nil, fmt.Errorf("GatewaySettings not found in EdgeDevice spec")
	}

	client, err := lwm2mclient.NewClient(context.TODO(), lwm2mclient.Config{
		EndpointName:    edgedevice.Spec.GatewaySettings.LwM2MSetting.EndpointName,
		EndpointUrl:     *edgedevice.Spec.GatewaySettings.Address,
		Settings:        *edgedevice.Spec.GatewaySettings.LwM2MSetting,
		DeviceShifuHost: deviceShifuHost,
	})
	if err != nil {
		return nil, err
	}

	var gateway = &Gateway{
		edgeDevice: edgedevice,
		k8sClient:  krclient,
		client:     client,
	}

	if err := gateway.LoadConfiguration(); err != nil {
		return nil, err
	}

	gateway.pingIntervalSec = edgedevice.Spec.GatewaySettings.LwM2MSetting.PingIntervalSec

	return gateway, nil
}

// LoadCfg loads the configuration from the ConfigMap
func (g *Gateway) LoadConfiguration() error {
	// Load the configmap
	cfg, err := configmap.Load(ConfigmapFolderPath)
	if err != nil {
		return err
	}

	var instructions deviceshifubase.DeviceShifuInstructions
	if instructionInCfg, ok := cfg[ConfigmapInstructionsStr]; ok {
		err := yaml.Unmarshal([]byte(instructionInCfg), &instructions)
		if err != nil {
			logger.Errorf("Error parsing %v from ConfigMap, error: %v", ConfigmapInstructionsStr, err)
			return err
		}
	}

	var objMap = make(map[string]*lwm2mclient.Object)
	for instructionName, instruction := range instructions.Instructions {
		// Skip if instruction is nil
		if instruction == nil {
			logger.Errorf("Instruction %v is nil", instructionName)
			continue
		}

		// Skip if instruction does not set the ObjectId
		objectId, exists := instruction.DeviceShifuGatewayProperties[ObjectIdStr]
		if !exists {
			logger.Errorf("Instruction %v does not have an ObjectId", instructionName)
			continue
		}

		var gwInstruction ShifuInstruction
		gwInstruction.ObjectId = objectId
		gwInstruction.Endpoint = g.client.DeviceShifuHost + "/" + instructionName
		gwInstruction.DataType, exists = instruction.DeviceShifuGatewayProperties[DataTypeStr]
		if !exists {
			// Default to string if DataType is not set
			gwInstruction.DataType = "string"
		}

		var resourceId string
		var objPath string
		// parse the object id to get the resource id and the object path
		// example: /3303/0/5700 3303 is the resource id and 0/5700 is the object path
		paths := strings.Split(objectId, "/")
		if len(paths) < 2 {
			logger.Errorf("Invalid object id: %v", objectId)
			continue
		}

		for index, path := range paths {
			if path != "" {
				resourceId = path
				objPath = strings.Join(paths[index+1:], "/")
				break
			}
		}

		// Create the object if it does not exist
		if _, exists := objMap[resourceId]; !exists {
			objMap[resourceId] = lwm2mclient.NewObject(resourceId, nil)
		}

		objMap[resourceId].AddObject(objPath, &gwInstruction)
	}

	// Add the objects to the client
	for _, obj := range objMap {
		g.client.AddObject(*obj)
	}

	return nil
}

// Start starts the gateway
func (g *Gateway) Start() error {
	// Start the client
	if err := g.client.Start(); err != nil {
		return err
	}

	if g.pingIntervalSec > 0 {
		// Ping the client every pingIntervalSec seconds,by default disable
		t := time.NewTicker(time.Second * time.Duration(g.pingIntervalSec))
		for range t.C {
			if err := g.client.Ping(); err != nil {
				logger.Errorf("Error pinging client: %v", err)
				g.ShutDown()
				return err
			}
		}
	}

	return nil
}

func (g *Gateway) ShutDown() {
	g.client.CleanUp()
}

type ShifuInstruction struct {
	ObjectId string
	Endpoint string
	DataType string
}

func (si *ShifuInstruction) Read() (interface{}, error) {
	resp, err := http.Get(si.Endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error reading data: %v", resp.Status)
	}

	rawData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	switch si.DataType {
	case "int":
		return strconv.Atoi(string(rawData))
	case "float":
		return strconv.ParseFloat(string(rawData), 64)
	case "bool":
		return strconv.ParseBool(string(rawData))
	default:
		// Default to string
	}

	return string(rawData), nil
}

func (si *ShifuInstruction) Write(data interface{}) error {
	dataStr := fmt.Sprintf("%v", data)

	req, err := http.NewRequest(http.MethodPut, si.Endpoint, strings.NewReader(dataStr))
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error writing data: %v", resp.Status)
	}

	return nil
}

func (si *ShifuInstruction) Execute() error {
	resp, err := http.Post(si.Endpoint, "plain/text", nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error executing instruction: %v", resp.Status)
	}

	return nil
}
