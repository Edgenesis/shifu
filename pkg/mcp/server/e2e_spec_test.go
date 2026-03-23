package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

// specDevices returns the three exact example devices from shifu_mcp_spec.md §4.4.
func specDevices() []v1alpha1.EdgeDevice {
	return []v1alpha1.EdgeDevice{
		// §4.4 MQTT Device — Robot Arm
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "edgedevice-robot-arm",
				Namespace: "devices",
			},
			Spec: v1alpha1.EdgeDeviceSpec{
				Sku:        strPtr("FANUC-M20iD"),
				Connection: (*v1alpha1.Connection)(strPtr("Ethernet")),
				Address:    strPtr("192.168.1.50"),
				Protocol:   protocolPtr(v1alpha1.ProtocolMQTT),
				Description: strPtr(`6-axis industrial robot arm (FANUC M-20iD) on the main assembly line.
Shifu translates PLC registers into MQTT topics.

**SAFETY:** Command interactions (` + "`robot-arm/commands/*`" + `) actuate real
joints and the gripper. Validate joint angles before publishing.`),
				ConnectionInfo: strPtr(`MQTT broker: mqtt://deviceshifu-robot-arm.deviceshifu.svc.cluster.local:1883
No authentication required. Use QoS 1 for commands, QoS 0 for status.

` + "```python" + `
import paho.mqtt.client as mqtt
client = mqtt.Client()
client.connect("deviceshifu-robot-arm.deviceshifu.svc.cluster.local", 1883)
` + "```"),
			},
			Status: v1alpha1.EdgeDeviceStatus{
				EdgeDevicePhase: phasePtr(v1alpha1.EdgeDeviceRunning),
			},
		},
		// §4.4 NATS Device — Sensor Array
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "edgedevice-sensor-array",
				Namespace: "devices",
			},
			Spec: v1alpha1.EdgeDeviceSpec{
				Protocol: (*v1alpha1.Protocol)(strPtr("NATS")),
				Address:  strPtr("/dev/ttyUSB0"),
				Description: strPtr(`Distributed sensor array across the warehouse floor. 24 sensor nodes.
Shifu translates proprietary RS-485 serial protocol into NATS subjects.`),
				ConnectionInfo: strPtr(`NATS server: nats://deviceshifu-sensor-array.deviceshifu.svc.cluster.local:4222
No authentication required. Use NATS wildcards for multiple sensors.

` + "```python" + `
import nats
nc = await nats.connect("nats://deviceshifu-sensor-array.deviceshifu.svc.cluster.local:4222")
` + "```"),
			},
			Status: v1alpha1.EdgeDeviceStatus{
				EdgeDevicePhase: phasePtr(v1alpha1.EdgeDeviceRunning),
			},
		},
		// §4.4 HTTP Device — Temperature Sensor
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "edgedevice-thermometer",
				Namespace: "devices",
			},
			Spec: v1alpha1.EdgeDeviceSpec{
				Protocol:    protocolPtr(v1alpha1.ProtocolHTTP),
				Address:     strPtr("192.168.1.100:502"),
				Description: strPtr(`Industrial temperature sensor. Calibrated for -40°C to 200°C range.`),
				ConnectionInfo: strPtr(`Base URL: http://deviceshifu-thermometer.deviceshifu.svc.cluster.local
No authentication required.`),
			},
			Status: v1alpha1.EdgeDeviceStatus{
				EdgeDevicePhase: phasePtr(v1alpha1.EdgeDeviceRunning),
			},
		},
	}
}

// specK8sObjects returns all K8s resources (Deployments + ConfigMaps) matching spec §4.4.
func specK8sObjects() []runtime.Object {
	return []runtime.Object{
		// --- Services ---
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deviceshifu-robot-arm",
				Namespace: "deviceshifu",
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{"app": "deviceshifu-robot-arm"},
			},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deviceshifu-sensor-array",
				Namespace: "deviceshifu",
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{"app": "deviceshifu-sensor-array"},
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

		// --- Robot Arm (MQTT) ---
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
							Name:  "deviceshifu-http",
							Image: "edgehub/deviceshifu-http-mqtt:nightly",
							Env: []corev1.EnvVar{
								{Name: "EDGEDEVICE_NAME", Value: "edgedevice-robot-arm"},
								{Name: "EDGEDEVICE_NAMESPACE", Value: "devices"},
							},
						}},
						Volumes: []corev1.Volume{{
							Name: "deviceshifu-config",
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
      ## Topic
      ` + "`robot-arm/commands/move_joint`" + `
      ## Message format (JSON)
      ` + "```json" + `
      {"joint": 1, "angle": 45.0, "speed": 50}
      ` + "```" + `
      - ` + "`joint`" + `: 1-6 (axis number)
      - ` + "`angle`" + `: degrees. Safe ranges: J1 ±170, J2 -100/+75, J3 -70/+200, J4 ±190, J5 ±125, J6 ±360
      - ` + "`speed`" + `: 1-100 (% of max speed)
  gripper:
    readWrite: W
    safe: false
    description: |
      Open or close the gripper.
      ## Topic
      ` + "`robot-arm/commands/gripper`" + `
      ## Message format
      ` + "```json" + `
      {"action": "close", "force": 80}
      ` + "```" + `
  joint_positions:
    readWrite: R
    safe: true
    description: |
      Real-time joint positions. Subscribe to receive continuous updates.
      ## Topic
      ` + "`robot-arm/status/joint_positions`" + `
      Published every 100ms. Array is [J1..J6] in degrees.
  emergency_stop:
    readWrite: W
    safe: false
    description: |
      Immediately halt all motion.
      ## Topic
      ` + "`robot-arm/commands/emergency_stop`" + `
      Publish any message to trigger E-stop.
`,
			},
		},

		// --- Sensor Array (NATS) ---
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deviceshifu-sensor-array",
				Namespace: "deviceshifu",
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app": "deviceshifu-sensor-array"},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"app": "deviceshifu-sensor-array"},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name:  "deviceshifu-http",
							Image: "edgehub/deviceshifu-http-nats:nightly",
							Env: []corev1.EnvVar{
								{Name: "EDGEDEVICE_NAME", Value: "edgedevice-sensor-array"},
								{Name: "EDGEDEVICE_NAMESPACE", Value: "devices"},
							},
						}},
						Volumes: []corev1.Volume{{
							Name: "deviceshifu-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "deviceshifu-sensor-array-configmap",
									},
								},
							},
						}},
					},
				},
			},
		},
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deviceshifu-sensor-array-configmap",
				Namespace: "deviceshifu",
			},
			Data: map[string]string{
				"instructions": `instructions:
  temperature:
    readWrite: R
    safe: true
    description: |
      Temperature readings. Subject: ` + "`sensors.<node_id>.temperature`" + `
      Wildcard: ` + "`sensors.*.temperature`" + ` for all nodes.
      Published every 5 seconds per node.
  vibration:
    readWrite: R
    safe: true
    description: |
      Vibration readings. Subject: ` + "`sensors.<node_id>.vibration`" + `
      Values above 0.5g indicate potential failure.
  configure_interval:
    readWrite: W
    safe: false
    description: |
      Change reporting interval. Uses NATS request/reply.
      Subject: ` + "`sensors.<node_id>.config.interval`" + `
      Valid intervals: 1-60 seconds. Default is 5.
`,
			},
		},

		// --- Thermometer (HTTP) ---
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
							Name:  "deviceshifu-http",
							Image: "edgehub/deviceshifu-http-http:nightly",
							Env: []corev1.EnvVar{
								{Name: "EDGEDEVICE_NAME", Value: "edgedevice-thermometer"},
								{Name: "EDGEDEVICE_NAMESPACE", Value: "devices"},
							},
						}},
						Volumes: []corev1.Volume{{
							Name: "deviceshifu-config",
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
      Updates every 3 seconds.
  set_unit:
    readWrite: W
    safe: false
    description: |
      POST /set_unit {"unit": "fahrenheit"}
      Response: {"status": "ok", "unit": "fahrenheit"}
  status:
    readWrite: R
    safe: true
    description: |
      GET /status — returns plain text: ` + "`running`" + ` or ` + "`error: <message>`" + `.
`,
			},
		},
	}
}

// setupSpecSession creates an MCP client session backed by all 3 spec example devices
// over a real Streamable HTTP transport.
func setupSpecSession(t *testing.T) *mcp.ClientSession {
	t.Helper()

	devices := specDevices()
	k8sObjects := specK8sObjects()
	fakeClient := fake.NewSimpleClientset(k8sObjects...)

	lister := func(ctx context.Context) ([]v1alpha1.EdgeDevice, error) {
		return devices, nil
	}

	resolver := deviceapi.NewResolver(fakeClient, lister)
	apiClient := deviceapi.NewClient(resolver)
	mcpServer := New(apiClient)

	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return mcpServer
	}, nil)
	httpServer := httptest.NewServer(handler)
	t.Cleanup(httpServer.Close)

	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "spec-test-client", Version: "v0.0.1"}, nil)
	transport := &mcp.StreamableClientTransport{Endpoint: httpServer.URL}
	session, err := client.Connect(ctx, transport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { session.Close() })

	return session
}

// TestE2ESpecListDevicesAll verifies list_devices returns all 3 spec devices
// with correct fields per §5 list_devices.
func TestE2ESpecListDevicesAll(t *testing.T) {
	session := setupSpecSession(t)
	ctx := context.Background()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "list_devices"})
	require.NoError(t, err)
	require.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var devices []deviceapi.DeviceSummary
	require.NoError(t, json.Unmarshal([]byte(text), &devices))
	require.Len(t, devices, 3, "spec defines 3 example devices")

	deviceMap := make(map[string]deviceapi.DeviceSummary)
	for _, d := range devices {
		deviceMap[d.Name] = d
	}

	// §5: list_devices response fields: name, namespace, description, protocol, phase, service
	robotArm := deviceMap["edgedevice-robot-arm"]
	assert.Equal(t, "devices", robotArm.Namespace)
	assert.Equal(t, "MQTT", robotArm.Protocol)
	assert.Equal(t, "Running", robotArm.Phase)
	assert.Contains(t, robotArm.Description, "robot arm")
	assert.Contains(t, robotArm.Service, "deviceshifu-robot-arm")

	sensorArray := deviceMap["edgedevice-sensor-array"]
	assert.Equal(t, "devices", sensorArray.Namespace)
	assert.Equal(t, "NATS", sensorArray.Protocol)
	assert.Equal(t, "Running", sensorArray.Phase)
	assert.Contains(t, sensorArray.Description, "sensor array")
	assert.Contains(t, sensorArray.Service, "deviceshifu-sensor-array")

	thermometer := deviceMap["edgedevice-thermometer"]
	assert.Equal(t, "devices", thermometer.Namespace)
	assert.Equal(t, "HTTP", thermometer.Protocol)
	assert.Equal(t, "Running", thermometer.Phase)
	assert.Contains(t, thermometer.Description, "temperature sensor")
	assert.Contains(t, thermometer.Service, "deviceshifu-thermometer")
}

// TestE2ESpecRobotArmMQTT verifies get_device_desc for the MQTT robot arm
// matches §4.4 and §5 get_device_desc MQTT example.
func TestE2ESpecRobotArmMQTT(t *testing.T) {
	session := setupSpecSession(t)
	ctx := context.Background()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "get_device_desc",
		Arguments: map[string]any{"device_name": "edgedevice-robot-arm"},
	})
	require.NoError(t, err)
	require.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var desc deviceapi.DeviceDesc
	require.NoError(t, json.Unmarshal([]byte(text), &desc))

	// §5 get_device_desc MQTT example response fields
	assert.Equal(t, "edgedevice-robot-arm", desc.Name)
	assert.Equal(t, "MQTT", desc.Protocol)
	assert.Equal(t, "Running", desc.Phase)
	assert.Contains(t, desc.Service, "deviceshifu-robot-arm")

	// §4.4: description contains FANUC and SAFETY note
	assert.Contains(t, desc.Description, "FANUC M-20iD")
	assert.Contains(t, desc.Description, "SAFETY")

	// §4.4: connectionInfo contains MQTT broker URL and Python example
	assert.Contains(t, desc.ConnectionInfo, "mqtt://deviceshifu-robot-arm")
	assert.Contains(t, desc.ConnectionInfo, "paho.mqtt")

	// §4.4: 4 interactions — move_joint, gripper, joint_positions, emergency_stop
	require.Len(t, desc.Interactions, 4, "robot arm spec has exactly 4 interactions")

	interactionMap := make(map[string]deviceapi.Interaction)
	for _, intr := range desc.Interactions {
		interactionMap[intr.Name] = intr
	}

	// move_joint: W, unsafe, topic info
	moveJoint := interactionMap["move_joint"]
	assert.Equal(t, "W", moveJoint.ReadWrite)
	require.NotNil(t, moveJoint.Safe)
	assert.False(t, *moveJoint.Safe)
	assert.Contains(t, moveJoint.Description, "robot-arm/commands/move_joint")
	assert.Contains(t, moveJoint.Description, "joint")
	assert.Contains(t, moveJoint.Description, "angle")

	// gripper: W, unsafe
	gripper := interactionMap["gripper"]
	assert.Equal(t, "W", gripper.ReadWrite)
	require.NotNil(t, gripper.Safe)
	assert.False(t, *gripper.Safe)
	assert.Contains(t, gripper.Description, "robot-arm/commands/gripper")

	// joint_positions: R, safe
	jointPos := interactionMap["joint_positions"]
	assert.Equal(t, "R", jointPos.ReadWrite)
	require.NotNil(t, jointPos.Safe)
	assert.True(t, *jointPos.Safe)
	assert.Contains(t, jointPos.Description, "robot-arm/status/joint_positions")

	// emergency_stop: W, unsafe
	eStop := interactionMap["emergency_stop"]
	assert.Equal(t, "W", eStop.ReadWrite)
	require.NotNil(t, eStop.Safe)
	assert.False(t, *eStop.Safe)
	assert.Contains(t, eStop.Description, "robot-arm/commands/emergency_stop")
}

// TestE2ESpecSensorArrayNATS verifies get_device_desc for the NATS sensor array
// matches §4.4 NATS example.
func TestE2ESpecSensorArrayNATS(t *testing.T) {
	session := setupSpecSession(t)
	ctx := context.Background()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "get_device_desc",
		Arguments: map[string]any{"device_name": "edgedevice-sensor-array"},
	})
	require.NoError(t, err)
	require.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var desc deviceapi.DeviceDesc
	require.NoError(t, json.Unmarshal([]byte(text), &desc))

	// §4.4 NATS sensor array
	assert.Equal(t, "edgedevice-sensor-array", desc.Name)
	assert.Equal(t, "NATS", desc.Protocol)
	assert.Equal(t, "Running", desc.Phase)
	assert.Contains(t, desc.Service, "deviceshifu-sensor-array")

	// Description mentions 24 sensor nodes and RS-485
	assert.Contains(t, desc.Description, "24 sensor nodes")
	assert.Contains(t, desc.Description, "RS-485")

	// ConnectionInfo has NATS server URL and Python example
	assert.Contains(t, desc.ConnectionInfo, "nats://deviceshifu-sensor-array")
	assert.Contains(t, desc.ConnectionInfo, "nats.connect")

	// §4.4: 3 interactions — temperature, vibration, configure_interval
	require.Len(t, desc.Interactions, 3, "sensor array spec has exactly 3 interactions")

	interactionMap := make(map[string]deviceapi.Interaction)
	for _, intr := range desc.Interactions {
		interactionMap[intr.Name] = intr
	}

	// temperature: R, safe, NATS subject pattern
	temp := interactionMap["temperature"]
	assert.Equal(t, "R", temp.ReadWrite)
	require.NotNil(t, temp.Safe)
	assert.True(t, *temp.Safe)
	assert.Contains(t, temp.Description, "sensors.")
	assert.Contains(t, temp.Description, "temperature")

	// vibration: R, safe
	vib := interactionMap["vibration"]
	assert.Equal(t, "R", vib.ReadWrite)
	require.NotNil(t, vib.Safe)
	assert.True(t, *vib.Safe)
	assert.Contains(t, vib.Description, "vibration")

	// configure_interval: W, unsafe, request/reply
	cfgInterval := interactionMap["configure_interval"]
	assert.Equal(t, "W", cfgInterval.ReadWrite)
	require.NotNil(t, cfgInterval.Safe)
	assert.False(t, *cfgInterval.Safe)
	assert.Contains(t, cfgInterval.Description, "request/reply")
}

// TestE2ESpecThermometerHTTP verifies get_device_desc for the HTTP thermometer
// matches §4.4 HTTP example.
func TestE2ESpecThermometerHTTP(t *testing.T) {
	session := setupSpecSession(t)
	ctx := context.Background()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "get_device_desc",
		Arguments: map[string]any{"device_name": "edgedevice-thermometer"},
	})
	require.NoError(t, err)
	require.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var desc deviceapi.DeviceDesc
	require.NoError(t, json.Unmarshal([]byte(text), &desc))

	// §4.4 HTTP thermometer
	assert.Equal(t, "edgedevice-thermometer", desc.Name)
	assert.Equal(t, "HTTP", desc.Protocol)
	assert.Equal(t, "Running", desc.Phase)
	assert.Contains(t, desc.Service, "deviceshifu-thermometer")

	assert.Contains(t, desc.Description, "temperature sensor")
	assert.Contains(t, desc.ConnectionInfo, "http://deviceshifu-thermometer")

	// §4.4: 3 interactions — get_temperature, set_unit, status
	require.Len(t, desc.Interactions, 3, "thermometer spec has exactly 3 interactions")

	interactionMap := make(map[string]deviceapi.Interaction)
	for _, intr := range desc.Interactions {
		interactionMap[intr.Name] = intr
	}

	// get_temperature: R, safe, GET endpoint
	getTemp := interactionMap["get_temperature"]
	assert.Equal(t, "R", getTemp.ReadWrite)
	require.NotNil(t, getTemp.Safe)
	assert.True(t, *getTemp.Safe)
	assert.Contains(t, getTemp.Description, "GET /get_temperature")

	// set_unit: W, unsafe, POST endpoint
	setUnit := interactionMap["set_unit"]
	assert.Equal(t, "W", setUnit.ReadWrite)
	require.NotNil(t, setUnit.Safe)
	assert.False(t, *setUnit.Safe)
	assert.Contains(t, setUnit.Description, "POST /set_unit")

	// status: R, safe
	status := interactionMap["status"]
	assert.Equal(t, "R", status.ReadWrite)
	require.NotNil(t, status.Safe)
	assert.True(t, *status.Safe)
	assert.Contains(t, status.Description, "GET /status")
}

// TestE2ESpecDeviceNotFound verifies error handling per §10.
func TestE2ESpecDeviceNotFound(t *testing.T) {
	session := setupSpecSession(t)
	ctx := context.Background()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "get_device_desc",
		Arguments: map[string]any{"device_name": "camera1"},
	})
	require.NoError(t, err)
	require.True(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	// §10: error response format
	assert.Contains(t, text, "DEVICE_NOT_FOUND")
	assert.Contains(t, text, "camera1")
}

// TestE2ESpecGracefulDegradation verifies §4.7 — devices without extended fields
// still return instruction names with minimal metadata.
func TestE2ESpecGracefulDegradation(t *testing.T) {
	// A device with no description/connectionInfo and a ConfigMap with
	// instructions that have no extended fields (just instruction names).
	devices := []v1alpha1.EdgeDevice{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "edgedevice-legacy",
				Namespace: "devices",
			},
			Spec: v1alpha1.EdgeDeviceSpec{
				Protocol: protocolPtr(v1alpha1.ProtocolHTTP),
			},
			Status: v1alpha1.EdgeDeviceStatus{
				EdgeDevicePhase: phasePtr(v1alpha1.EdgeDeviceRunning),
			},
		},
	}

	k8sObjects := []runtime.Object{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deviceshifu-legacy",
				Namespace: "deviceshifu",
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app": "deviceshifu-legacy"},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"app": "deviceshifu-legacy"},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name:  "deviceshifu-http",
							Image: "edgehub/deviceshifu-http-http:nightly",
							Env: []corev1.EnvVar{
								{Name: "EDGEDEVICE_NAME", Value: "edgedevice-legacy"},
							},
						}},
						Volumes: []corev1.Volume{{
							Name: "deviceshifu-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "legacy-configmap",
									},
								},
							},
						}},
					},
				},
			},
		},
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "legacy-configmap",
				Namespace: "deviceshifu",
			},
			Data: map[string]string{
				"instructions": `instructions:
  read_value:
  write_config:
`,
			},
		},
	}

	fakeClient := fake.NewSimpleClientset(k8sObjects...)
	lister := func(ctx context.Context) ([]v1alpha1.EdgeDevice, error) {
		return devices, nil
	}

	resolver := deviceapi.NewResolver(fakeClient, lister)
	apiClient := deviceapi.NewClient(resolver)
	mcpServer := New(apiClient)

	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return mcpServer
	}, nil)
	httpServer := httptest.NewServer(handler)
	defer httpServer.Close()

	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "degradation-test", Version: "v0.0.1"}, nil)
	transport := &mcp.StreamableClientTransport{Endpoint: httpServer.URL}
	session, err := client.Connect(ctx, transport, nil)
	require.NoError(t, err)
	defer session.Close()

	// list_devices: should work, description and service should be populated (or empty)
	listResult, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "list_devices"})
	require.NoError(t, err)
	require.False(t, listResult.IsError)

	listText := listResult.Content[0].(*mcp.TextContent).Text
	var summaries []deviceapi.DeviceSummary
	require.NoError(t, json.Unmarshal([]byte(listText), &summaries))
	require.Len(t, summaries, 1)
	assert.Equal(t, "edgedevice-legacy", summaries[0].Name)
	assert.Equal(t, "HTTP", summaries[0].Protocol)
	assert.Equal(t, "", summaries[0].Description, "no description field on legacy device")

	// get_device_desc: should return instruction names with minimal metadata
	descResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "get_device_desc",
		Arguments: map[string]any{"device_name": "edgedevice-legacy"},
	})
	require.NoError(t, err)
	require.False(t, descResult.IsError)

	descText := descResult.Content[0].(*mcp.TextContent).Text
	var desc deviceapi.DeviceDesc
	require.NoError(t, json.Unmarshal([]byte(descText), &desc))

	assert.Equal(t, "edgedevice-legacy", desc.Name)
	assert.Equal(t, "HTTP", desc.Protocol)
	assert.Equal(t, "", desc.Description)
	assert.Equal(t, "", desc.ConnectionInfo)
	assert.Len(t, desc.Interactions, 2, "should return 2 instruction names even without extended fields")

	interactionNames := make(map[string]bool)
	for _, intr := range desc.Interactions {
		interactionNames[intr.Name] = true
		assert.Equal(t, "", intr.ReadWrite, "legacy instructions have no readWrite")
		assert.Nil(t, intr.Safe, "legacy instructions have no safe field")
		assert.Equal(t, "", intr.Description, "legacy instructions have no description")
	}
	assert.True(t, interactionNames["read_value"])
	assert.True(t, interactionNames["write_config"])
}

// TestE2ESpecMixedProtocolWorkflow simulates §6 Mixed-Protocol App workflow:
// list_devices to discover all protocols, then get_device_desc for each.
func TestE2ESpecMixedProtocolWorkflow(t *testing.T) {
	session := setupSpecSession(t)
	ctx := context.Background()

	// Step 1: Agent calls list_devices()
	listResult, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "list_devices"})
	require.NoError(t, err)
	require.False(t, listResult.IsError)

	listText := listResult.Content[0].(*mcp.TextContent).Text
	var summaries []deviceapi.DeviceSummary
	require.NoError(t, json.Unmarshal([]byte(listText), &summaries))
	require.Len(t, summaries, 3)

	// Verify we have all 3 protocols
	protocols := make(map[string]bool)
	for _, s := range summaries {
		protocols[s.Protocol] = true
	}
	assert.True(t, protocols["HTTP"], "should have HTTP device")
	assert.True(t, protocols["MQTT"], "should have MQTT device")
	assert.True(t, protocols["NATS"], "should have NATS device")

	// Step 2: Agent calls get_device_desc for each device (as in §6 Mixed-Protocol)
	for _, summary := range summaries {
		t.Run(summary.Name, func(t *testing.T) {
			result, err := session.CallTool(ctx, &mcp.CallToolParams{
				Name:      "get_device_desc",
				Arguments: map[string]any{"device_name": summary.Name},
			})
			require.NoError(t, err)
			require.False(t, result.IsError)

			text := result.Content[0].(*mcp.TextContent).Text
			var desc deviceapi.DeviceDesc
			require.NoError(t, json.Unmarshal([]byte(text), &desc))

			// Every device should have non-empty protocol, phase, description, connectionInfo, interactions
			assert.NotEmpty(t, desc.Protocol)
			assert.Equal(t, "Running", desc.Phase)
			assert.NotEmpty(t, desc.Description)
			assert.NotEmpty(t, desc.ConnectionInfo)
			assert.NotEmpty(t, desc.Interactions)
			assert.Contains(t, desc.Service, "deviceshifu-")
		})
	}
}
