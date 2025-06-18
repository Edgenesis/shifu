package utils

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/edgenesis/shifu/pkg/k8s/crd/usermetrics/types"
	"github.com/edgenesis/shifu/pkg/logger"
)

const (
	URL_SHIFU_TELEMETRY    = "https://telemetry.shifu.dev/shifu-telemetry/"
	URL_DEFAULT_PUBLIC_IP  = "0.0.0.0"
	TASK_RUN_DEMO_KIND     = "run_shifu_release"
	DEFAULT_SOURCE         = "default"
	HTTP_CONTENT_TYPE_JSON = "application/json"
)

var TelemetryIntervalInSecond int

func SendUserMetrics(telemetry types.UserMetricsResponse) error {
	postBodyJson, err := json.Marshal(telemetry)
	if err != nil {
		logger.Errorf("Error marshaling telemetry")
		return err
	}

	resp, err := http.Post(URL_SHIFU_TELEMETRY, HTTP_CONTENT_TYPE_JSON, bytes.NewBuffer(postBodyJson))
	if err != nil {
		logger.Errorln("error posting telemetry, errors: ", err)
		return err
	}

	defer func() { _ = resp.Body.Close() }()
	return nil
}
