package deviceshifubase

import (
	"errors"
	"testing"

	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidateTelemetryConfig(t *testing.T) {
	testCases := []struct {
		Name        string
		inputDevice *DeviceShifuBase
		expErrStr   string
	}{
		{
			"case 1 no setting",
			&DeviceShifuBase{
				DeviceShifuConfig: &DeviceShifuConfig{
					Telemetries: &DeviceShifuTelemetries{},
				},
			},
			"",
		},
		{
			"case 2 has pushsetting with negative interval",
			&DeviceShifuBase{
				Name: "test",
				DeviceShifuConfig: &DeviceShifuConfig{
					Telemetries: &DeviceShifuTelemetries{
						DeviceShifuTelemetrySettings: &DeviceShifuTelemetrySettings{
							DeviceShifuTelemetryDefaultPushToServer:          boolPointer(true),
							DeviceShifuTelemetryDefaultCollectionService:     strPointer("test_endpoint-1"),
							DeviceShifuTelemetryUpdateIntervalInMilliseconds: int64Pointer(-1),
						},
					},
				},
			},
			"error deviceShifuTelemetryInterval mustn't be negative number",
		},
		{
			"case 3 has pushsetting with negative initial delay",
			&DeviceShifuBase{
				Name: "test",
				DeviceShifuConfig: &DeviceShifuConfig{
					Telemetries: &DeviceShifuTelemetries{
						DeviceShifuTelemetrySettings: &DeviceShifuTelemetrySettings{
							DeviceShifuTelemetryDefaultPushToServer:        boolPointer(true),
							DeviceShifuTelemetryDefaultCollectionService:   strPointer("test_endpoint-1"),
							DeviceShifuTelemetryInitialDelayInMilliseconds: int64Pointer(-1),
						},
					},
				},
			},
			"error deviceShifuTelemetryInitialDelay mustn't be negative number",
		},
		{
			"case 4 has pushsetting with negative timeout",
			&DeviceShifuBase{
				Name: "test",
				DeviceShifuConfig: &DeviceShifuConfig{
					Telemetries: &DeviceShifuTelemetries{
						DeviceShifuTelemetrySettings: &DeviceShifuTelemetrySettings{
							DeviceShifuTelemetryDefaultPushToServer:      boolPointer(true),
							DeviceShifuTelemetryDefaultCollectionService: strPointer("test_endpoint-1"),
							DeviceShifuTelemetryTimeoutInMilliseconds:    int64Pointer(-1),
						},
					},
				},
			},
			"error deviceShifuTelemetryTimeout mustn't be negative number",
		},
	}

	for _, c := range testCases {
		t.Run(c.Name, func(t *testing.T) {
			err := c.inputDevice.ValidateTelemetryConfig()
			if len(c.expErrStr) > 0 {
				assert.Equal(t, c.expErrStr, err.Error())
			} else {
				assert.Nil(t, err)
			}

		})
	}

}

func TestStartTelemetryCollection(t *testing.T) {
	mockds := &DeviceShifuBase{
		Name: "test",
		EdgeDevice: &v1alpha1.EdgeDevice{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test_namespace",
			},
		},
		DeviceShifuConfig: &DeviceShifuConfig{
			Telemetries: &DeviceShifuTelemetries{
				DeviceShifuTelemetrySettings: &DeviceShifuTelemetrySettings{
					DeviceShifuTelemetryDefaultPushToServer:      boolPointer(true),
					DeviceShifuTelemetryDefaultCollectionService: strPointer("test_endpoint-1"),
				},
				DeviceShifuTelemetries: map[string]*DeviceShifuTelemetry{
					"device_healthy": {
						DeviceShifuTelemetryProperties: DeviceShifuTelemetryProperties{
							PushSettings: &DeviceShifuTelemetryPushSettings{
								DeviceShifuTelemetryPushToServer:      boolPointer(false),
								DeviceShifuTelemetryCollectionService: strPointer("test_endpoint-1"),
							},
							InitialDelayMs: intPointer(1),
						},
					},
				},
			},
		},
		RestClient: mockRestClientFor("{\"spec\": {\"address\": \"http://192.168.15.48:12345/test_endpoint-1\",\"type\": \"HTTP\"}}", t),
	}

	testCases := []struct {
		Name        string
		inputDevice *DeviceShifuBase
		fn          func() (bool, error)
		expErrStr   string
	}{
		{
			"case 1 fn true with nil error",
			mockds,
			func() (bool, error) {
				return true, nil
			},
			"",
		},
		{
			"case 2 fn false with nil error",
			mockds,
			func() (bool, error) {
				return false, nil
			},
			"",
		},
		{
			"case 3 fn false with error",
			mockds,
			func() (bool, error) {
				return false, errors.New("exit")
			},
			"",
		},
	}

	for _, c := range testCases {
		t.Run(c.Name, func(t *testing.T) {
			err := c.inputDevice.telemetryCollection(c.fn)
			if len(c.expErrStr) > 0 {
				assert.Equal(t, c.expErrStr, err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}

}
