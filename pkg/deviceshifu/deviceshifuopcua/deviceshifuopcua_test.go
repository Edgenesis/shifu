package deviceshifuopcua

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/edgenesis/shifu/pkg/deviceshifu/unitest"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/edgenesis/shifu/pkg/logger"
	"github.com/gopcua/opcua/ua"
	"github.com/stretchr/testify/assert"

	"k8s.io/apimachinery/pkg/util/wait"
)

func TestMain(m *testing.M) {
	err := GenerateConfigMapFromSnippet(MockDeviceCmStr, MockDeviceConfigFolder)
	if err != nil {
		logger.Errorf("error when generateConfigmapFromSnippet,err: %v", err)
		os.Exit(-1)
	}
	m.Run()
	err = os.RemoveAll(MockDeviceConfigPath)
	if err != nil {
		logger.Fatal(err)
	}
}

func TestNew(t *testing.T) {
	deviceShifuMetadata := &deviceshifubase.DeviceShifuMetaData{
		Name:           "TeststartHTTPServer",
		ConfigFilePath: "etc/edgedevice/config",
		KubeConfigPath: deviceshifubase.DeviceKubeconfigDoNotLoadStr,
		Namespace:      "TeststartHTTPServerNamespace",
	}

	_, err := New(deviceShifuMetadata)
	if err != nil {
		t.Errorf("Failed creating new deviceshifu")
	}
}

func TestDeviceShifuEmptyNamespace(t *testing.T) {
	deviceShifuMetadata := &deviceshifubase.DeviceShifuMetaData{
		Name:           "TestDeviceShifuEmptyNamespace",
		ConfigFilePath: "etc/edgedevice/config",
		KubeConfigPath: deviceshifubase.DeviceKubeconfigDoNotLoadStr,
	}

	_, err := New(deviceShifuMetadata)
	if err != nil {
		logger.Errorf("%v", err)
	} else {
		t.Errorf("DeviceShifu Test with empty namespace failed")
	}
}

func TestStart(t *testing.T) {
	deviceShifuMetadata := &deviceshifubase.DeviceShifuMetaData{
		Name:           "TestStartOPCUA",
		ConfigFilePath: "etc/edgedevice/config",
		KubeConfigPath: deviceshifubase.DeviceKubeconfigDoNotLoadStr,
		Namespace:      "TestStartNamespace",
	}

	mockds, err := New(deviceShifuMetadata)
	if err != nil {
		t.Errorf("Failed creating new deviceshifu")
	}

	if err := mockds.Start(wait.NeverStop); err != nil {
		t.Errorf("DeviceShifu.Start failed due to: %v", err.Error())
	}

	if err := mockds.Stop(); err != nil {
		t.Errorf("unable to stop mock deviceShifu, error: %+v", err)
	}
}

func TestDeviceHealthHandler(t *testing.T) {
	deviceShifuMetadata := &deviceshifubase.DeviceShifuMetaData{
		Name:           "TeststartHTTPServerOPCUA",
		ConfigFilePath: "etc/edgedevice/config",
		KubeConfigPath: deviceshifubase.DeviceKubeconfigDoNotLoadStr,
		Namespace:      "TeststartHTTPServerNamespace",
	}

	mockds, err := New(deviceShifuMetadata)
	if err != nil {
		t.Errorf("Failed creating new deviceshifu")
	}

	if err := mockds.Start(wait.NeverStop); err != nil {
		t.Errorf("DeviceShifu.Start failed due to: %v", err.Error())
	}

	resp, err := unitest.RetryAndGetHTTP("http://localhost:8080/health", 3)
	if err != nil {
		t.Errorf("HTTP GET returns an error %v", err.Error())
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("unable to read response body, error: %v", err.Error())
	}

	if string(body) != deviceshifubase.DeviceIsHealthyStr {
		t.Errorf("%+v", body)
	}

	if err := mockds.Stop(); err != nil {
		t.Errorf("unable to stop mock deviceShifu, error: %+v", err)
	}
}

func TestCreateValue(t *testing.T) {
	assert := assert.New(t)
	testCases := []struct {
		name         string
		ref          interface{}
		newValue     interface{}
		exceptOutput interface{}
	}{
		{
			name:         "int64 to int",
			ref:          int(0),
			newValue:     int64(64),
			exceptOutput: int(64),
		},
		{
			name:         "int to int16",
			ref:          int16(0),
			newValue:     int(64),
			exceptOutput: int16(64),
		},
		{
			name:         "int to int32",
			ref:          int32(0),
			newValue:     int(64),
			exceptOutput: int32(64),
		},
		{
			name:         "string to int",
			ref:          int(0),
			newValue:     "64",
			exceptOutput: nil,
		},
		{
			name:         "float64 to int16",
			ref:          int16(0),
			newValue:     float64(64.1),
			exceptOutput: int16(64),
		},
		{
			name:         "int to float32",
			ref:          float32(0),
			newValue:     123,
			exceptOutput: float32(123),
		},
		{
			name:         "nil to nil",
			ref:          nil,
			newValue:     nil,
			exceptOutput: nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			output := convertValueToRef(tC.ref, tC.newValue)
			assert.Equal(tC.exceptOutput, output)
		})
	}
}

// MockClient is a mock implementation of OPCUAClient
type MockClient struct {
	ReadFunc  func(ctx context.Context, req *ua.ReadRequest) (*ua.ReadResponse, error)
	WriteFunc func(ctx context.Context, req *ua.WriteRequest) (*ua.WriteResponse, error)
}

func (m *MockClient) Read(ctx context.Context, req *ua.ReadRequest) (*ua.ReadResponse, error) {
	if m.ReadFunc != nil {
		return m.ReadFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockClient) Write(ctx context.Context, req *ua.WriteRequest) (*ua.WriteResponse, error) {
	if m.WriteFunc != nil {
		return m.WriteFunc(ctx, req)
	}
	return nil, nil
}

func TestCollectOPCUATelemetry(t *testing.T) {
	// Mock Request
	instruction := "test_instruction"
	nodeID := "ns=2;s=TestNode"

	// Mock Configuration
	telemetries := &deviceshifubase.DeviceShifuTelemetries{
		DeviceShifuTelemetries: map[string]*deviceshifubase.DeviceShifuTelemetry{
			"test_telemetry": {
				DeviceShifuTelemetryProperties: deviceshifubase.DeviceShifuTelemetryProperties{
					DeviceInstructionName: &instruction,
				},
			},
		},
	}

	opcuaInstructions := &OPCUAInstructions{
		Instructions: map[string]*OPCUAInstruction{
			instruction: {
				OPCUAInstructionProperty: &OPCUAInstructionProperty{
					OPCUANodeID: nodeID,
				},
			},
		},
	}

	protocol := v1alpha1.ProtocolOPCUA
	address := "opc.tcp://localhost"
	mockEdgeDeviceSpec := v1alpha1.EdgeDeviceSpec{
		Protocol: &protocol,
		Address:  &address,
	}

	ds := &DeviceShifu{
		base: &deviceshifubase.DeviceShifuBase{
			Name: "test-device",
			EdgeDevice: &v1alpha1.EdgeDevice{
				Spec: mockEdgeDeviceSpec,
			},
			DeviceShifuConfig: &deviceshifubase.DeviceShifuConfig{
				Telemetries: telemetries,
			},
		},
		opcuaInstructions: opcuaInstructions,
	}

	// Case 1: Success
	mockClientSuccess := &MockClient{
		ReadFunc: func(ctx context.Context, req *ua.ReadRequest) (*ua.ReadResponse, error) {
			return &ua.ReadResponse{
				Results: []*ua.DataValue{
					{Status: ua.StatusOK, Value: ua.MustVariant("good")},
				},
			}, nil
		},
	}
	ds.opcuaClient = mockClientSuccess

	ok, err := ds.collectOPCUATelemetry()
	assert.Nil(t, err)
	assert.True(t, ok)

	// Case 2: Read Error
	mockClientError := &MockClient{
		ReadFunc: func(ctx context.Context, req *ua.ReadRequest) (*ua.ReadResponse, error) {
			return nil, fmt.Errorf("read error")
		},
	}
	ds.opcuaClient = mockClientError

	ok, err = ds.collectOPCUATelemetry()
	assert.NotNil(t, err)
	assert.False(t, ok)
	assert.Contains(t, err.Error(), "read error")

	// Case 3: Bad Status
	mockClientBadStatus := &MockClient{
		ReadFunc: func(ctx context.Context, req *ua.ReadRequest) (*ua.ReadResponse, error) {
			return &ua.ReadResponse{
				Results: []*ua.DataValue{
					{Status: ua.StatusBad},
				},
			}, nil
		},
	}
	ds.opcuaClient = mockClientBadStatus

	ok, err = ds.collectOPCUATelemetry()
	assert.NotNil(t, err)
	assert.False(t, ok)
	assert.Contains(t, err.Error(), "StatusBad")
}
