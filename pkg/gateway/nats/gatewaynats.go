package nats

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/edgenesis/shifu/pkg/gateway/nats/client"
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
	PublisherIntervalMsStr   = "PublisherIntervalMs"
	TopicStr                 = "Topic"
)

type Mode string

const (
	ModePublisher  Mode = "publisher"
	ModeSubscriber Mode = "subscriber"
)

type InstructionConfig struct {
	Topic               string `yaml:"topic"`
	Mode                Mode   `yaml:"mode"`
	PublisherIntervalMs int    `yaml:"publisherIntervalMs"`
}

type Gateway struct {
	ctx        context.Context
	k8sClient  *rest.RESTClient
	edgeDevice *v1alpha1.EdgeDevice

	natsClient   *client.Client
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
		natsClient:   natsClient,
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

		topic, exists := instruction.DeviceShifuGatewayProperties[TopicStr]
		if !exists {
			continue
		}
		ic.Topic = topic

		// Skip if instruction does not set the Mode
		mode, exists := instruction.DeviceShifuGatewayProperties[ModeStr]
		if exists {
			ic.Mode = Mode(mode)
		}

		publisherIntervalMs, exists := instruction.DeviceShifuGatewayProperties[PublisherIntervalMsStr]
		if exists {
			publisherIntervalMsInt, err := strconv.Atoi(publisherIntervalMs)
			if err != nil {
				logger.Errorf("Instruction %v has an invalid PublisherIntervalMs: %v", instructionName, err)
				continue
			}
			ic.PublisherIntervalMs = publisherIntervalMsInt
		}

		g.instructions[instructionName] = ic
	}

	logger.Infof("Loaded %v instructions", len(g.instructions))

	return nil
}

func (g *Gateway) Start() error {
	for instructionName, instruction := range g.instructions {
		switch instruction.Mode {
		case ModePublisher:
			logger.Infof("Instruction %v is a NATS instruction", instructionName)
			go g.RegisterPublisher(instructionName, instruction.Topic)
		case ModeSubscriber:
			logger.Infof("Instruction %v is a NATS instruction", instructionName)
			err := g.natsClient.Subscribe(instruction.Topic, func(msg *nats.Msg) {
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

func (g *Gateway) RegisterPublisher(instructionName string, topic string) {
	interval := g.instructions[instructionName].PublisherIntervalMs
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

			if err := g.natsClient.Publish(topic, body); err != nil {
				logger.Errorf("Error publishing message for instruction %v: %v", instructionName, err)
			}
		case <-g.ctx.Done():
			return
		}
	}
}
