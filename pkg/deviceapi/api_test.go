package deviceapi

import (
	"context"
	"testing"

	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func strPtr(s string) *string { return &s }

func protocolPtr(p v1alpha1.Protocol) *v1alpha1.Protocol { return &p }
func phasePtr(p v1alpha1.EdgeDevicePhase) *v1alpha1.EdgeDevicePhase { return &p }

func newTestDevices() []v1alpha1.EdgeDevice {
	return []v1alpha1.EdgeDevice{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "edgedevice-thermometer",
				Namespace: "devices",
			},
			Spec: v1alpha1.EdgeDeviceSpec{
				Protocol:    protocolPtr(v1alpha1.ProtocolHTTP),
				Address:     strPtr("192.168.1.100:502"),
				Description: strPtr("Industrial temperature sensor.\nCalibrated for -40C to 200C."),
				ConnectionInfo: strPtr("Base URL: http://deviceshifu-thermometer.deviceshifu.svc.cluster.local\nNo auth required."),
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
				Protocol:    protocolPtr(v1alpha1.ProtocolMQTT),
				Description: strPtr("6-axis robot arm"),
			},
			Status: v1alpha1.EdgeDeviceStatus{
				EdgeDevicePhase: phasePtr(v1alpha1.EdgeDeviceRunning),
			},
		},
	}
}

func newTestDeploymentAndConfigMap() []runtime.Object {
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
						Containers: []corev1.Container{
							{
								Name:  "deviceshifu",
								Image: "edgehub/deviceshifu-http-http:latest",
								Env: []corev1.EnvVar{
									{Name: "EDGEDEVICE_NAME", Value: "edgedevice-thermometer"},
									{Name: "EDGEDEVICE_NAMESPACE", Value: "devices"},
								},
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "edgedevice-config",
								VolumeSource: corev1.VolumeSource{
									ConfigMap: &corev1.ConfigMapVolumeSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "deviceshifu-thermometer-configmap",
										},
									},
								},
							},
						},
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
  set_unit:
    readWrite: W
    safe: false
    description: |
      POST /set_unit {"unit": "fahrenheit"}
  status:
    readWrite: R
    safe: true
`,
			},
		},
		// Robot arm deployment (no ConfigMap for testing graceful degradation)
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
						Containers: []corev1.Container{
							{
								Name:  "deviceshifu",
								Image: "edgehub/deviceshifu-mqtt:latest",
								Env: []corev1.EnvVar{
									{Name: "EDGEDEVICE_NAME", Value: "edgedevice-robot-arm"},
								},
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "edgedevice-config",
								VolumeSource: corev1.VolumeSource{
									ConfigMap: &corev1.ConfigMapVolumeSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "deviceshifu-robot-arm-configmap",
										},
									},
								},
							},
						},
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
      Topic: robot-arm/commands/move_joint
  joint_positions:
    readWrite: R
    safe: true
    description: |
      Real-time joint positions.
      Topic: robot-arm/status/joint_positions
`,
			},
		},
	}
}

func setupTestClient() *Client {
	devices := newTestDevices()
	k8sObjects := newTestDeploymentAndConfigMap()
	fakeClient := fake.NewSimpleClientset(k8sObjects...)

	lister := func(ctx context.Context) ([]v1alpha1.EdgeDevice, error) {
		return devices, nil
	}

	resolver := NewResolver(fakeClient, lister)
	return NewClient(resolver)
}

func TestListDevices(t *testing.T) {
	client := setupTestClient()
	ctx := context.Background()

	summaries, err := client.ListDevices(ctx)
	require.NoError(t, err)
	assert.Len(t, summaries, 2)

	// Find thermometer
	var thermo *DeviceSummary
	for i := range summaries {
		if summaries[i].Name == "edgedevice-thermometer" {
			thermo = &summaries[i]
			break
		}
	}
	require.NotNil(t, thermo)
	assert.Equal(t, "devices", thermo.Namespace)
	assert.Equal(t, "HTTP", thermo.Protocol)
	assert.Equal(t, "Running", thermo.Phase)
	assert.Equal(t, "Industrial temperature sensor.", thermo.Description) // First line only
	assert.Equal(t, "deviceshifu-thermometer.deviceshifu.svc.cluster.local", thermo.Service)
}

func TestGetDeviceDesc(t *testing.T) {
	client := setupTestClient()
	ctx := context.Background()

	desc, err := client.GetDeviceDesc(ctx, "edgedevice-thermometer")
	require.NoError(t, err)
	require.NotNil(t, desc)

	assert.Equal(t, "edgedevice-thermometer", desc.Name)
	assert.Equal(t, "HTTP", desc.Protocol)
	assert.Equal(t, "Running", desc.Phase)
	assert.Contains(t, desc.Description, "Industrial temperature sensor.")
	assert.Contains(t, desc.ConnectionInfo, "Base URL")
	assert.Equal(t, "deviceshifu-thermometer.deviceshifu.svc.cluster.local", desc.Service)

	// Check interactions
	assert.GreaterOrEqual(t, len(desc.Interactions), 2)

	interactionMap := make(map[string]Interaction)
	for _, intr := range desc.Interactions {
		interactionMap[intr.Name] = intr
	}

	getTemp, ok := interactionMap["get_temperature"]
	require.True(t, ok, "get_temperature interaction should exist")
	assert.Equal(t, "R", getTemp.ReadWrite)
	assert.NotNil(t, getTemp.Safe)
	assert.True(t, *getTemp.Safe)
	assert.Contains(t, getTemp.Description, "GET /get_temperature")

	setUnit, ok := interactionMap["set_unit"]
	require.True(t, ok, "set_unit interaction should exist")
	assert.Equal(t, "W", setUnit.ReadWrite)
	assert.NotNil(t, setUnit.Safe)
	assert.False(t, *setUnit.Safe)
}

func TestGetDeviceDescNotFound(t *testing.T) {
	client := setupTestClient()
	ctx := context.Background()

	_, err := client.GetDeviceDesc(ctx, "nonexistent-device")
	require.Error(t, err)

	var notFound *DeviceNotFoundError
	assert.ErrorAs(t, err, &notFound)
	assert.Equal(t, "nonexistent-device", notFound.Name)
}

func TestGetDeviceDescRobotArm(t *testing.T) {
	client := setupTestClient()
	ctx := context.Background()

	desc, err := client.GetDeviceDesc(ctx, "edgedevice-robot-arm")
	require.NoError(t, err)
	require.NotNil(t, desc)

	assert.Equal(t, "MQTT", desc.Protocol)
	assert.Equal(t, "Running", desc.Phase)
	assert.Equal(t, "6-axis robot arm", desc.Description)

	interactionMap := make(map[string]Interaction)
	for _, intr := range desc.Interactions {
		interactionMap[intr.Name] = intr
	}

	moveJoint, ok := interactionMap["move_joint"]
	require.True(t, ok)
	assert.Equal(t, "W", moveJoint.ReadWrite)
	assert.NotNil(t, moveJoint.Safe)
	assert.False(t, *moveJoint.Safe)

	jointPos, ok := interactionMap["joint_positions"]
	require.True(t, ok)
	assert.Equal(t, "R", jointPos.ReadWrite)
	assert.NotNil(t, jointPos.Safe)
	assert.True(t, *jointPos.Safe)
}

func TestListDevicesNoDevices(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()
	lister := func(ctx context.Context) ([]v1alpha1.EdgeDevice, error) {
		return nil, nil
	}
	resolver := NewResolver(fakeClient, lister)
	client := NewClient(resolver)

	summaries, err := client.ListDevices(context.Background())
	require.NoError(t, err)
	assert.Empty(t, summaries)
}

func TestListDevicesNoDescriptionField(t *testing.T) {
	// Device without description — graceful degradation.
	devices := []v1alpha1.EdgeDevice{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "edgedevice-sensor",
				Namespace: "devices",
			},
			Spec: v1alpha1.EdgeDeviceSpec{
				Protocol: protocolPtr(v1alpha1.ProtocolHTTP),
			},
		},
	}

	fakeClient := fake.NewSimpleClientset()
	lister := func(ctx context.Context) ([]v1alpha1.EdgeDevice, error) {
		return devices, nil
	}
	resolver := NewResolver(fakeClient, lister)
	client := NewClient(resolver)

	summaries, err := client.ListDevices(context.Background())
	require.NoError(t, err)
	assert.Len(t, summaries, 1)
	assert.Equal(t, "", summaries[0].Description)
	assert.Equal(t, "HTTP", summaries[0].Protocol)
}
