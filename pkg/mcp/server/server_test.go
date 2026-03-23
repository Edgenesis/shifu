package server

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/edgenesis/shifu/pkg/deviceapi"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func strPtr(s string) *string                                       { return &s }
func protocolPtr(p v1alpha1.Protocol) *v1alpha1.Protocol            { return &p }
func phasePtr(p v1alpha1.EdgeDevicePhase) *v1alpha1.EdgeDevicePhase { return &p }

func testDevices() []v1alpha1.EdgeDevice {
	return []v1alpha1.EdgeDevice{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "edgedevice-thermometer",
				Namespace: "devices",
			},
			Spec: v1alpha1.EdgeDeviceSpec{
				Protocol:       protocolPtr(v1alpha1.ProtocolHTTP),
				Description:    strPtr("Industrial temperature sensor.\nCalibrated for -40C to 200C."),
				ConnectionInfo: strPtr("Base URL: http://deviceshifu-thermometer.deviceshifu.svc.cluster.local"),
			},
			Status: v1alpha1.EdgeDeviceStatus{
				EdgeDevicePhase: phasePtr(v1alpha1.EdgeDeviceRunning),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "edgedevice-robot-arm",
				Namespace: "devices",
			},
			Spec: v1alpha1.EdgeDeviceSpec{
				Protocol:       protocolPtr(v1alpha1.ProtocolMQTT),
				Description:    strPtr("6-axis robot arm (FANUC M-20iD)"),
				ConnectionInfo: strPtr("MQTT broker: mqtt://deviceshifu-robot-arm.deviceshifu.svc.cluster.local:1883"),
			},
			Status: v1alpha1.EdgeDeviceStatus{
				EdgeDevicePhase: phasePtr(v1alpha1.EdgeDeviceRunning),
			},
		},
	}
}

func testK8sObjects() []runtime.Object {
	return []runtime.Object{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deviceshifu-thermometer",
				Namespace: "deviceshifu",
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app": "deviceshifu-thermometer"},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"app": "deviceshifu-thermometer"},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name:  "deviceshifu",
							Image: "edgehub/deviceshifu-http-http:latest",
							Env: []corev1.EnvVar{
								{Name: "EDGEDEVICE_NAME", Value: "edgedevice-thermometer"},
								{Name: "EDGEDEVICE_NAMESPACE", Value: "devices"},
							},
						}},
						Volumes: []corev1.Volume{{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "deviceshifu-thermometer-configmap",
									},
								},
							},
						}},
					},
				},
			},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deviceshifu-thermometer",
				Namespace: "deviceshifu",
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{"app": "deviceshifu-thermometer"},
			},
		},
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deviceshifu-thermometer-configmap",
				Namespace: "deviceshifu",
			},
			Data: map[string]string{
				"instructions": `instructions:
  get_temperature:
    readWrite: R
    safe: true
    description: |
      GET /get_temperature
      Response: {"temperature": 36.5, "unit": "celsius"}
  set_unit:
    readWrite: W
    safe: false
    description: |
      POST /set_unit {"unit": "fahrenheit"}
`,
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deviceshifu-robot-arm",
				Namespace: "deviceshifu",
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app": "deviceshifu-robot-arm"},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"app": "deviceshifu-robot-arm"},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name:  "deviceshifu",
							Image: "edgehub/deviceshifu-mqtt:latest",
							Env: []corev1.EnvVar{
								{Name: "EDGEDEVICE_NAME", Value: "edgedevice-robot-arm"},
							},
						}},
						Volumes: []corev1.Volume{{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "deviceshifu-robot-arm-configmap",
									},
								},
							},
						}},
					},
				},
			},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deviceshifu-robot-arm",
				Namespace: "deviceshifu",
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{"app": "deviceshifu-robot-arm"},
			},
		},
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deviceshifu-robot-arm-configmap",
				Namespace: "deviceshifu",
			},
			Data: map[string]string{
				"instructions": `instructions:
  move_joint:
    readWrite: W
    safe: false
    description: |
      Move a specific joint to a target angle.
      Topic: robot-arm/commands/move_joint
      Message format: {"joint": 1, "angle": 45.0, "speed": 50}
  joint_positions:
    readWrite: R
    safe: true
    description: |
      Real-time joint positions. Subscribe to receive continuous updates.
      Topic: robot-arm/status/joint_positions
  emergency_stop:
    readWrite: W
    safe: false
    description: |
      Immediately halt all motion. Publish any message to trigger.
      Topic: robot-arm/commands/emergency_stop
`,
			},
		},
	}
}

func setupMCPClientSession(t *testing.T) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()

	devices := testDevices()
	k8sObjects := testK8sObjects()
	fakeClient := fake.NewSimpleClientset(k8sObjects...)

	lister := func(ctx context.Context) ([]v1alpha1.EdgeDevice, error) {
		return devices, nil
	}

	resolver := deviceapi.NewResolver(fakeClient, lister)
	apiClient := deviceapi.NewClient(resolver)
	server := New(apiClient)

	t1, t2 := mcp.NewInMemoryTransports()
	_, err := server.Connect(ctx, t1, nil)
	require.NoError(t, err)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.0.1"}, nil)
	session, err := client.Connect(ctx, t2, nil)
	require.NoError(t, err)

	t.Cleanup(func() { session.Close() })
	return session
}

func TestMCPServerTools(t *testing.T) {
	session := setupMCPClientSession(t)
	ctx := context.Background()

	// Verify tools are registered.
	var toolNames []string
	for tool, err := range session.Tools(ctx, nil) {
		require.NoError(t, err)
		toolNames = append(toolNames, tool.Name)
	}
	assert.Contains(t, toolNames, "list_devices")
	assert.Contains(t, toolNames, "get_device_desc")
}

func TestMCPListDevices(t *testing.T) {
	session := setupMCPClientSession(t)
	ctx := context.Background()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "list_devices",
	})
	require.NoError(t, err)
	require.False(t, result.IsError, "list_devices should not return an error")
	require.Len(t, result.Content, 1)

	text := result.Content[0].(*mcp.TextContent).Text
	var devices []deviceapi.DeviceSummary
	err = json.Unmarshal([]byte(text), &devices)
	require.NoError(t, err)
	assert.Len(t, devices, 2)

	deviceMap := make(map[string]deviceapi.DeviceSummary)
	for _, d := range devices {
		deviceMap[d.Name] = d
	}

	thermo, ok := deviceMap["edgedevice-thermometer"]
	require.True(t, ok)
	assert.Equal(t, "HTTP", thermo.Protocol)
	assert.Equal(t, "Running", thermo.Phase)
	assert.Equal(t, "devices", thermo.Namespace)
	assert.Contains(t, thermo.Service, "deviceshifu-thermometer")

	robot, ok := deviceMap["edgedevice-robot-arm"]
	require.True(t, ok)
	assert.Equal(t, "MQTT", robot.Protocol)
	assert.Equal(t, "Running", robot.Phase)
}

func TestMCPGetDeviceDescHTTP(t *testing.T) {
	session := setupMCPClientSession(t)
	ctx := context.Background()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "get_device_desc",
		Arguments: map[string]any{"device_name": "edgedevice-thermometer"},
	})
	require.NoError(t, err)
	require.False(t, result.IsError, "get_device_desc should not return an error")
	require.Len(t, result.Content, 1)

	text := result.Content[0].(*mcp.TextContent).Text
	var desc deviceapi.DeviceDesc
	err = json.Unmarshal([]byte(text), &desc)
	require.NoError(t, err)

	assert.Equal(t, "edgedevice-thermometer", desc.Name)
	assert.Equal(t, "HTTP", desc.Protocol)
	assert.Equal(t, "Running", desc.Phase)
	assert.Contains(t, desc.Description, "Industrial temperature sensor")
	assert.Contains(t, desc.ConnectionInfo, "Base URL")
	assert.Contains(t, desc.Service, "deviceshifu-thermometer")

	// Check interactions.
	interactionMap := make(map[string]deviceapi.Interaction)
	for _, intr := range desc.Interactions {
		interactionMap[intr.Name] = intr
	}

	getTemp, ok := interactionMap["get_temperature"]
	require.True(t, ok)
	assert.Equal(t, "R", getTemp.ReadWrite)
	assert.NotNil(t, getTemp.Safe)
	assert.True(t, *getTemp.Safe)
	assert.Contains(t, getTemp.Description, "GET /get_temperature")

	setUnit, ok := interactionMap["set_unit"]
	require.True(t, ok)
	assert.Equal(t, "W", setUnit.ReadWrite)
	assert.NotNil(t, setUnit.Safe)
	assert.False(t, *setUnit.Safe)
}

func TestMCPGetDeviceDescMQTT(t *testing.T) {
	session := setupMCPClientSession(t)
	ctx := context.Background()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "get_device_desc",
		Arguments: map[string]any{"device_name": "edgedevice-robot-arm"},
	})
	require.NoError(t, err)
	require.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var desc deviceapi.DeviceDesc
	err = json.Unmarshal([]byte(text), &desc)
	require.NoError(t, err)

	assert.Equal(t, "MQTT", desc.Protocol)
	assert.Contains(t, desc.ConnectionInfo, "MQTT broker")

	interactionMap := make(map[string]deviceapi.Interaction)
	for _, intr := range desc.Interactions {
		interactionMap[intr.Name] = intr
	}

	moveJoint, ok := interactionMap["move_joint"]
	require.True(t, ok)
	assert.Equal(t, "W", moveJoint.ReadWrite)
	assert.NotNil(t, moveJoint.Safe)
	assert.False(t, *moveJoint.Safe)
	assert.Contains(t, moveJoint.Description, "robot-arm/commands/move_joint")

	jointPos, ok := interactionMap["joint_positions"]
	require.True(t, ok)
	assert.Equal(t, "R", jointPos.ReadWrite)
	assert.NotNil(t, jointPos.Safe)
	assert.True(t, *jointPos.Safe)

	eStop, ok := interactionMap["emergency_stop"]
	require.True(t, ok)
	assert.Equal(t, "W", eStop.ReadWrite)
}

func TestMCPGetDeviceDescNotFound(t *testing.T) {
	session := setupMCPClientSession(t)
	ctx := context.Background()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "get_device_desc",
		Arguments: map[string]any{"device_name": "nonexistent-device"},
	})
	require.NoError(t, err)
	require.True(t, result.IsError, "should return tool error for nonexistent device")

	text := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, text, "DEVICE_NOT_FOUND")
	assert.Contains(t, text, "nonexistent-device")
}

func TestMCPGetDeviceDescMissingArg(t *testing.T) {
	session := setupMCPClientSession(t)
	ctx := context.Background()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "get_device_desc",
		Arguments: map[string]any{},
	})
	require.NoError(t, err)
	require.True(t, result.IsError, "should return error for missing device_name")

	text := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, text, "device_name is required")
}
