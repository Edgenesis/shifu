package gatewaylwm2m

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/edgenesis/shifu/pkg/gateway/gatewaylwm2m/lwm2m"
	"github.com/edgenesis/shifu/pkg/logger"
	"gopkg.in/yaml.v3"
)

const CONFIG_FILE = "/etc/gateway/config/instructions"

type Gateway struct {
	client *lwm2m.Client
}

func New() (*Gateway, error) {
	endpoint := os.Getenv("ENDPOINT")
	serverUrl := os.Getenv("SERVER_URL")
	client, err := lwm2m.NewClient(serverUrl, endpoint)
	if err != nil {
		return nil, err
	}

	var gateway = &Gateway{
		client: client,
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
	file, err := os.Open(CONFIG_FILE)
	if err != nil {
		return err
	}
	defer file.Close()
	var config []Config

	if err := yaml.NewDecoder(file).Decode(&config); err != nil {
		return err
	}

	var objectMap = make(map[string]*lwm2m.Object)
	for _, obj := range config {
		var instruction ShifuInstruction
		rsName := strings.TrimPrefix(obj.ResourceId, "/")
		instruction.ResourceId = rsName
		instruction.ObjectId = obj.ObjectId
		instruction.DataType = obj.DataType
		instruction.Endpoint = fmt.Sprintf("http://%s.%s.svc.cluster.local/%s", obj.ServiceName, obj.NameSpace, obj.InstructionName)
		if _, ok := objectMap[rsName]; !ok {
			objectMap[rsName] = lwm2m.NewObject(rsName, nil)
		}

		objectMap[rsName].AddObject(obj.ObjectId, &instruction)
	}

	for _, obj := range objectMap {
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
	ResourceId string
	ObjectId   string
	Endpoint   string
	DataType   string
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
