package deviceshifubase

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"edgenesis.io/shifu/k8s/crd/api/v1alpha1"
)

// HTTP header type:
// type Header map[string][]string
func CopyHeader(dst, src http.Header) {
	for header, headerValueList := range src {
		for _, value := range headerValueList {
			dst.Add(header, value)
		}
	}
}

func PushToHTTPTelemetryCollectionService(telemetryServiceProtocol v1alpha1.Protocol,
	message *http.Response, telemetryCollectionService string) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(DEVICE_TELEMETRY_TIMEOUT_MS)*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, telemetryCollectionService, message.Body)
	if err != nil {
		log.Printf("error creating request for telemetry service, error: %v\n" + err.Error())
		return
	}

	log.Printf("pushing %v to %v", message.Body, telemetryCollectionService)
	CopyHeader(req.Header, req.Header)
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("HTTP POST error for telemetry service %v, error: %v\n", telemetryCollectionService, err.Error())
		return
	}
}

func getTelemetryCollectionServiceMap(ds *DeviceShifuBase) (map[string]string, error) {
	serviceAddressMap := make(map[string]string)
	res := make(map[string]string)
	defaultPushToServer := false
	defaultTelemetryCollectionService := ""
	defaultTelemetryServiceAddress := ""
	telemetries := ds.DeviceShifuConfig.Telemetries
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
			Resource(TELEMETRYCOLLECTIONSERVICE_RESOURCE_STR).
			Name(defaultTelemetryCollectionService).
			Do(context.TODO()).
			Into(&telemetryService); err != nil {
			log.Printf("unable to get telemetry service %v, error: %v", defaultTelemetryCollectionService, err)
		}

		defaultTelemetryServiceAddress = *telemetryService.Spec.Address
	}

	for telemetryName, telemetries := range telemetries.DeviceShifuTelemetries {
		pushSettings := telemetries.DeviceShifuTelemetryProperties.PushSettings

		if pushSettings != nil {
			if pushSettings.DeviceShifuTelemetryPushToServer != nil {
				if !*pushSettings.DeviceShifuTelemetryPushToServer {
					continue
				}
			}

			if pushSettings.DeviceShifuTelemetryCollectionService != nil ||
				len(*pushSettings.DeviceShifuTelemetryCollectionService) != 0 {
				if telemetryServiceAddress, exist := serviceAddressMap[*pushSettings.DeviceShifuTelemetryCollectionService]; exist {
					res[telemetryName] = telemetryServiceAddress
				}

				var telemetryService v1alpha1.TelemetryService
				if err := ds.RestClient.Get().
					Namespace(ds.EdgeDevice.Namespace).
					Resource(TELEMETRYCOLLECTIONSERVICE_RESOURCE_STR).
					Name(*pushSettings.DeviceShifuTelemetryCollectionService).
					Do(context.TODO()).
					Into(&telemetryService); err != nil {
					log.Printf("unable to get telemetry service %v, error: %v", *pushSettings.DeviceShifuTelemetryCollectionService, err)
					continue
				}

				serviceAddressMap[*pushSettings.DeviceShifuTelemetryCollectionService] = *telemetryService.Spec.Address
				res[telemetryName] = *telemetryService.Spec.Address
				continue
			}
		}

		res[telemetryName] = defaultTelemetryServiceAddress
	}

	return res, nil
}
