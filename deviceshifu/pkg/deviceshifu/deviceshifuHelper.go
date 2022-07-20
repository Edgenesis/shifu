package deviceshifu

import (
	"errors"
)

func (ds *DeviceShifu) ValidateTelemetryConfig() error {
	if ds.deviceShifuConfig.Telemetries.DeviceShifuTelemetrySettings == nil {
		ds.deviceShifuConfig.Telemetries.DeviceShifuTelemetrySettings = &DeviceShifuTelemetrySettings{}
	}

	var dsTelemetrySettings = ds.deviceShifuConfig.Telemetries.DeviceShifuTelemetrySettings
	if initial := dsTelemetrySettings.DeviceShifuTelemetryInitialDelayInMilliseconds; initial == nil {
		var telemetryInitialDelayInMilliseconds = DEVICE_TELEMETRY_INITIAL_DELAY_MS
		dsTelemetrySettings.DeviceShifuTelemetryInitialDelayInMilliseconds = &telemetryInitialDelayInMilliseconds
	} else if *initial < 0 {
		return errors.New("error deviceShifuTelemetryInitialDelay mustn't be negative number")
	}

	if timeout := dsTelemetrySettings.DeviceShifuTelemetryTimeoutInMilliseconds; timeout == nil {
		var telemetryTimeoutInMilliseconds = DEVICE_TELEMETRY_TIMEOUT_MS
		dsTelemetrySettings.DeviceShifuTelemetryTimeoutInMilliseconds = &telemetryTimeoutInMilliseconds
	} else if *timeout < 0 {
		return errors.New("error deviceShifuTelemetryTimeout mustn't be negative number")
	}

	if interval := dsTelemetrySettings.DeviceShifuTelemetryUpdateIntervalInMilliseconds; interval == nil {
		var telemetryUpdateIntervalInMilliseconds = DEVICE_TELEMETRY_UPDATE_INTERVAL_MS
		dsTelemetrySettings.DeviceShifuTelemetryUpdateIntervalInMilliseconds = &telemetryUpdateIntervalInMilliseconds
	} else if *interval < 0 {
		return errors.New("error deviceShifuTelemetryInterval mustn't be negative number")
	}

	return nil
}
