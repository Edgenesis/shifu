package deviceshifuopcua

import (
	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/edgenesis/shifu/pkg/deviceshifu/unitest"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

func TestMain(m *testing.M) {
	err := GenerateConfigMapFromSnippet(MockDeviceCmStr, MockDeviceConfigFolder)
	if err != nil {
		klog.Errorf("error when generateConfigmapFromSnippet,err: %v", err)
		os.Exit(-1)
	}
	m.Run()
	err = os.RemoveAll(MockDeviceConfigPath)
	if err != nil {
		klog.Fatal(err)
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
		klog.Errorf("%v", err)
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

func TestCollectOPCUATelemetry(t *testing.T) {
	testCases := []struct {
		Name        string
		inputDevice *DeviceShifu
		expected    bool
		expErrStr   string
	}{
		{
			"case 1 Protocol is nil",
			&DeviceShifu{
				base: &deviceshifubase.DeviceShifuBase{
					Name: "test",
					EdgeDevice: &v1alpha1.EdgeDevice{
						Spec: v1alpha1.EdgeDeviceSpec{
							Address: unitest.ToPointer("localhost"),
						},
					},
				},
			},
			true,
			"",
		},
		{
			"case 2 Address is nil",
			&DeviceShifu{
				base: &deviceshifubase.DeviceShifuBase{
					Name: "test",
					EdgeDevice: &v1alpha1.EdgeDevice{
						Spec: v1alpha1.EdgeDeviceSpec{
							Protocol: unitest.ToPointer(v1alpha1.ProtocolOPCUA),
						},
					},
					DeviceShifuConfig: &deviceshifubase.DeviceShifuConfig{
						Telemetries: &deviceshifubase.DeviceShifuTelemetries{
							DeviceShifuTelemetries: map[string]*deviceshifubase.DeviceShifuTelemetry{
								"": &deviceshifubase.DeviceShifuTelemetry{},
							},
						},
					},
				},
			},
			false,
			"Device test does not have an address",
		},
		{
			"case 3 DeviceInstructionName is nil",
			&DeviceShifu{
				base: &deviceshifubase.DeviceShifuBase{
					Name: "test",
					EdgeDevice: &v1alpha1.EdgeDevice{
						Spec: v1alpha1.EdgeDeviceSpec{
							Address:  unitest.ToPointer("localhost"),
							Protocol: unitest.ToPointer(v1alpha1.ProtocolOPCUA),
						},
					},
					DeviceShifuConfig: &deviceshifubase.DeviceShifuConfig{
						Telemetries: &deviceshifubase.DeviceShifuTelemetries{
							DeviceShifuTelemetries: map[string]*deviceshifubase.DeviceShifuTelemetry{
								"test_DeviceShifuTelemetries": {
									deviceshifubase.DeviceShifuTelemetryProperties{
										DeviceInstructionName: unitest.ToPointer(""),
									},
								},
							},
						},
					},
				},
				opcuaInstructions: &OPCUAInstructions{
					map[string]*OPCUAInstruction{
						"test_Instructions": {
							&OPCUAInstructionProperty{
								"OPCUANodeID",
							},
						},
					},
				},
			},
			false,
			"Instruction  not found in list of deviceshifu instructions",
		},
		//TODO : new opcuaClient
	}

	for _, c := range testCases {
		t.Run(c.Name, func(t *testing.T) {
			got, err := c.inputDevice.collectOPCUATelemetry()
			assert.Equal(t, c.expected, got)
			if got {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, c.expErrStr, err.Error())
			}
		})
	}
}
