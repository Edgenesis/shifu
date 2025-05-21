package natsio

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/edgenesis/shifu/pkg/gateway/natsio/client"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/edgenesis/shifu/pkg/logger"

	"github.com/nats-io/nats.go"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/rest"
	"knative.dev/pkg/configmap"
)

const (
	ConfigmapFolderPath      = "/etc/edgedevice/config"
	ConfigmapInstructionsStr = "instructions"
	deviceShifuHost          = "http://localhost:8080"
	ModeStr                  = "Mode"
	PubGetIntervalMsStr      = "PubGetIntervalMs"
)

type InstructionConfig struct {
	Mode             string `yaml:"mode"`
	PubGetIntervalMs int    `yaml:"pubGetIntervalMs"`
}

type Gateway struct {
	ctx        context.Context
	k8sClient  *rest.RESTClient
	edgeDevice *v1alpha1.EdgeDevice

	natsioClient *client.Client
	instructions map[string]InstructionConfig
}

func New() (*Gateway, error) {
	edgedevice, krclient, err := deviceshifubase.NewEdgeDevice(&deviceshifubase.EdgeDeviceConfig{
		NameSpace:  os.Getenv("EDGEDEVICE_NAMESPACE"),
		DeviceName: os.Getenv("EDGEDEVICE_NAME"),
	})
	if err != nil {
		return nil, err
	}

	natsClient, err := client.New(*edgedevice.Spec.GatewaySettings.Address)
	if err != nil {
		return nil, err
	}

	var gateway = &Gateway{
		k8sClient:    krclient,
		edgeDevice:   edgedevice,
		natsioClient: natsClient,
		instructions: make(map[string]InstructionConfig),
		ctx:          context.Background(),
	}

	err = gateway.LoadConfiguration()
	if err != nil {
		return nil, err
	}
	return gateway, nil
}

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

	for instructionName, instruction := range instructions.Instructions {
		// Skip if instruction is nil
		if instruction == nil {
			logger.Errorf("Instruction %v is nil", instructionName)
			continue
		}

		ic := InstructionConfig{}

		// Skip if instruction does not set the Mode
		mode, exists := instruction.DeviceShifuGatewayProperties[ModeStr]
		if exists {
			ic.Mode = mode
		}

		pubGetIntervalMs, exists := instruction.DeviceShifuGatewayProperties[PubGetIntervalMsStr]
		if exists {
			pubGetIntervalMsInt, err := strconv.Atoi(pubGetIntervalMs)
			if err != nil {
				logger.Errorf("Instruction %v has an invalid PubGetIntervalMs: %v", instructionName, err)
				continue
			}
			ic.PubGetIntervalMs = pubGetIntervalMsInt
		}

		g.instructions[instructionName] = ic
	}

	logger.Infof("Loaded %v instructions", len(g.instructions))

	return nil
}

func (g *Gateway) Start() error {
	for instructionName, instruction := range g.instructions {
		switch instruction.Mode {
		case "publisher":
			logger.Infof("Instruction %v is a NATSIO instruction", instructionName)
			go g.RegisterPublisher(instructionName)
		case "subscriber":
			logger.Infof("Instruction %v is a NATSIO instruction", instructionName)
			err := g.natsioClient.Subscribe(instructionName, func(msg *nats.Msg) {
				logger.Infof("Received message: %v", string(msg.Data))
				resp, err := http.Post(deviceShifuHost+instructionName, "plain/text", bytes.NewBuffer(msg.Data))
				if err != nil {
					logger.Errorf("Error sending message to deviceShifu: %v", err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != 200 {
					logger.Errorf("Error sending message to deviceShifu: %v", resp.StatusCode)
				}
			})
			if err != nil {
				return err
			}
		default:
			logger.Errorf("Instruction %v has an unknown mode: %v, SKIPPING", instructionName, instruction.Mode)
		}
	}

	return nil
}

func (g *Gateway) ShutDown() {
	return
}

func (g *Gateway) RegisterPublisher(instructionName string) {
	interval := g.instructions[instructionName].PubGetIntervalMs
	if interval == 0 {
		interval = 1000
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			resp, err := http.Get(deviceShifuHost + "/" + instructionName)
			if err != nil {
				logger.Errorf("Error registering publisher for instruction %v: %v", instructionName, err)
				continue
			}
			if resp.StatusCode != http.StatusOK {
				logger.Errorf("Error registering publisher for instruction %v: %v", instructionName, resp.StatusCode)
				resp.Body.Close()
				continue
			}

			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				logger.Errorf("Error reading response body for instruction %v: %v", instructionName, err)
				continue
			}

			if err := g.natsioClient.Publish(instructionName, body); err != nil {
				logger.Errorf("Error publishing message for instruction %v: %v", instructionName, err)
			}
		case <-g.ctx.Done():
			return
		}
	}
}
