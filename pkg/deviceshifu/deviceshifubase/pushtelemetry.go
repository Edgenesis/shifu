package deviceshifubase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/edgenesis/shifu/pkg/deviceshifu/utils"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"k8s.io/klog/v2"
)

type TelemetryRequest struct {
	RawData     []byte                `json:"rawData,omitempty"`
	MQTTSetting *v1alpha1.MQTTSetting `json:"mqttSetting,omitempty"`
}

func PushTelemetryCollectionService(tss *v1alpha1.TelemetryServiceSpec, message *http.Response) error {
	var err error
	switch *tss.Protocol {
	case v1alpha1.ProtocolHTTP:
		err = pushToHTTPTelemetryCollectionService(*tss.Protocol, message, *tss.Address)
	case v1alpha1.ProtocolMQTT:
		err = pushToMQTTTelemetryCollectionService(message, tss)
	default:
		return fmt.Errorf("unsupported protocol")
	}
	return err
}

// PushToHTTPTelemetryCollectionService push telemetry data to Collection Service
func pushToHTTPTelemetryCollectionService(telemetryServiceProtocol v1alpha1.Protocol,
	message *http.Response, telemetryCollectionService string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(DeviceTelemetryTimeoutInMS)*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, telemetryCollectionService, message.Body)
	if err != nil {
		klog.Errorf("error creating request for telemetry service, error: %v" + err.Error())
		return err
	}

	klog.Infof("pushing %v to %v", message.Body, telemetryCollectionService)
	utils.CopyHeader(req.Header, req.Header)
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		klog.Errorf("HTTP POST error for telemetry service %v, error: %v", telemetryCollectionService, err.Error())
		return err
	}
	return nil
}

func pushToMQTTTelemetryCollectionService(message *http.Response, settings *v1alpha1.TelemetryServiceSpec) error {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(DeviceTelemetryTimeoutInMS)*time.Millisecond)
	defer cancel()

	rawData, err := io.ReadAll(message.Body)
	if err != nil {
		klog.Errorf("Error when Read Info From RequestBody, error: %v", err)
		return err
	}
	request := TelemetryRequest{
		RawData:     rawData,
		MQTTSetting: settings.ServiceSettings.MQTTSetting,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		klog.Errorf("Error when marshal request to []byte, error: %v", err)
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, *settings.Address, bytes.NewBuffer(requestBody))
	if err != nil {
		klog.Errorf("Error when build request with requestBody, error: %v", err)
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		klog.Errorf("Error when send request to Server, error: %v", err)
		return err
	}
	klog.Infof("successfully sent message %v to telemetry service address %v", string(rawData), *settings.Address)
	err = resp.Body.Close()
	if err != nil {
		klog.Errorf("Error when Close response Body, error: %v", err)
		return err
	}

	return nil
}

func getTelemetryCollectionServiceMap(ds *DeviceShifuBase) (map[string]v1alpha1.TelemetryServiceSpec, error) {
	serviceAddressCache := make(map[string]v1alpha1.TelemetryServiceSpec)
	res := make(map[string]v1alpha1.TelemetryServiceSpec)
	defaultPushToServer := false
	defaultTelemetryCollectionService := ""
	defaultTelemetryServiceAddress := ""
	defaultTelemetryProtocol := v1alpha1.ProtocolHTTP
	defaultTelemetryServiceSpec := &v1alpha1.TelemetryServiceSpec{
		Protocol: &defaultTelemetryProtocol,
		Address:  &defaultTelemetryServiceAddress,
	}

	telemetries := ds.DeviceShifuConfig.Telemetries
	if telemetries == nil {
		return res, nil
	}

	if telemetries.DeviceShifuTelemetrySettings == nil {
		telemetries.DeviceShifuTelemetrySettings = &DeviceShifuTelemetrySettings{}
	}

	if telemetries.DeviceShifuTelemetrySettings.DeviceShifuTelemetryDefaultPushToServer != nil {
		defaultPushToServer = *telemetries.DeviceShifuTelemetrySettings.DeviceShifuTelemetryDefaultPushToServer
	}

	if defaultPushToServer {
		if telemetries.DeviceShifuTelemetrySettings.DeviceShifuTelemetryDefaultCollectionService == nil ||
			len(*telemetries.DeviceShifuTelemetrySettings.DeviceShifuTelemetryDefaultCollectionService) == 0 {
			return nil, fmt.Errorf("you need to configure defaultTelemetryCollectionService if setting defaultPushToServer to true")

		}

		defaultTelemetryCollectionService = *telemetries.DeviceShifuTelemetrySettings.DeviceShifuTelemetryDefaultCollectionService
		var telemetryService v1alpha1.TelemetryService
		if err := ds.RestClient.Get().
			Namespace(ds.EdgeDevice.Namespace).
			Resource(TelemetryCollectionServiceResourceStr).
			Name(defaultTelemetryCollectionService).
			Do(context.TODO()).
			Into(&telemetryService); err != nil {
			klog.Errorf("unable to get telemetry service %v, error: %v", defaultTelemetryCollectionService, err)
		}

		serviceAddressCache[defaultTelemetryCollectionService] = telemetryService.Spec
	}

	for telemetryName, telemetry := range telemetries.DeviceShifuTelemetries {
		if telemetry == nil {
			continue
		}

		pushSettings := telemetry.DeviceShifuTelemetryProperties.PushSettings
		if pushSettings == nil {
			res[telemetryName] = *defaultTelemetryServiceSpec
			continue
		}

		if pushSettings.DeviceShifuTelemetryPushToServer != nil {
			if !*pushSettings.DeviceShifuTelemetryPushToServer {
				continue
			}
		}

		if pushSettings.DeviceShifuTelemetryCollectionService != nil &&
			len(*pushSettings.DeviceShifuTelemetryCollectionService) != 0 {
			if telemetryServiceAddress, exist := serviceAddressCache[*pushSettings.DeviceShifuTelemetryCollectionService]; exist {
				res[telemetryName] = telemetryServiceAddress
				continue
			}

			var telemetryService v1alpha1.TelemetryService
			if err := ds.RestClient.Get().
				Namespace(ds.EdgeDevice.Namespace).
				Resource(TelemetryCollectionServiceResourceStr).
				Name(*pushSettings.DeviceShifuTelemetryCollectionService).
				Do(context.TODO()).
				Into(&telemetryService); err != nil {
				klog.Errorf("unable to get telemetry service %v, error: %v", *pushSettings.DeviceShifuTelemetryCollectionService, err)
				continue
			}

			serviceAddressCache[*pushSettings.DeviceShifuTelemetryCollectionService] = telemetryService.Spec
			res[telemetryName] = telemetryService.Spec
			continue
		}

		res[telemetryName] = *defaultTelemetryServiceSpec
	}

	return res, nil
}
