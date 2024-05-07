package gatewaylwm2m

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/edgenesis/shifu/pkg/gateway/gatewaylwm2m/lwm2m"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/edgenesis/shifu/pkg/logger"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/rest"
	"knative.dev/pkg/configmap"
)

const (
	ConfigmapFolderPath      = "/etc/edgedevice/config"
	ConfigmapInstructionsStr = "instructions"
	ObjectIdStr              = "ObjectId"
	DataTypeStr              = "DataType"
)

type Gateway struct {
	client         *lwm2m.Client
	KRestfulClient *rest.RESTClient
	edgedevice     *v1alpha1.EdgeDevice
}

func New() (*Gateway, error) {
	edgedevice, krclient, err := deviceshifubase.NewEdgeDevice(&deviceshifubase.EdgeDeviceConfig{
		NameSpace:  "devices",
		DeviceName: "edgedevice-lwm2m",
	})
	if err != nil {
		return nil, err
	}

	if edgedevice.Spec.GatewaySettings == nil {
		return nil, fmt.Errorf("GatewaySettings not found in EdgeDevice spec")
	}

	lwm2mSettings := edgedevice.Spec.GatewaySettings.LwM2MSettings

	client, err := lwm2m.NewClient(lwm2m.Config{
		EndpointUrl:  *edgedevice.Spec.GatewaySettings.Address,
		EndpointName: lwm2mSettings.EndpointName,
	})
	if err != nil {
		return nil, err
	}

	var gateway = &Gateway{
		edgedevice:     edgedevice,
		client:         client,
		KRestfulClient: krclient,
	}

	return gateway, nil
}

type Config struct {
	ServiceName     string `yaml:"serviceName"`
	NameSpace       string `yaml:"namespace"`
	InstructionName string `yaml:"instructionName"`
	ResourceId      string `yaml:"resourceId"`
	ObjectId        string `yaml:"objectId"`
	DataType        string `yaml:"type"`
}

func (g *Gateway) LoadCfg() error {
	cfg, err := configmap.Load(ConfigmapFolderPath)
	if err != nil {
		return err
	}

	var instructions deviceshifubase.DeviceShifuInstructions
	if instructions, ok := cfg[ConfigmapInstructionsStr]; ok {
		err := yaml.Unmarshal([]byte(instructions), &instructions)
		if err != nil {
			logger.Fatalf("Error parsing %v from ConfigMap, error: %v", ConfigmapInstructionsStr, err)
			return err
		}
	}

	for instructionName, instruction := range instructions.Instructions {
		objectId, exists := instruction.DeviceShifuGatewayProperties[ObjectIdStr]
		if !exists {
			continue
		}

		var gwInstruction ShifuInstruction
		gwInstruction.ObjectId = objectId
		gwInstruction.DataType, exists = instruction.DeviceShifuGatewayProperties[DataTypeStr]
		if !exists {
			gwInstruction.DataType = "string"
		}

		obj := lwm2m.NewObject(instructionName, &gwInstruction)
		g.client.AddObject(*obj)
	}

	return nil
}

func (g *Gateway) Start() error {
	err := g.client.Register()
	if err != nil {
		logger.Errorf("Error registering client: %v", err)
		return err
	}

	select {}
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
	}

	return string(rawData), nil
}

func (si *ShifuInstruction) Write(data interface{}) error {
	dataStr := fmt.Sprintf("%v", data)
	resp, err := http.Post(si.Endpoint, "plain/text", strings.NewReader(dataStr))
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
