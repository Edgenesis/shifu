package sql

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/edgenesis/shifu/pkg/logger"
	"github.com/edgenesis/shifu/pkg/telemetryservice/sql/tdengine"
)

func BindSQLServiceHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Errorf("Error when Read Data From Body, error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger.Infof("requestBody: %s", string(body))
	request := v1alpha1.TelemetryRequest{}

	err = json.Unmarshal(body, &request)
	if err != nil {
		logger.Errorf("Error to Unmarshal request body to struct")
		http.Error(w, "unexpected end of JSON input", http.StatusBadRequest)
		return
	}

	switch *request.SQLConnectionSetting.DBType {
	case v1alpha1.DBTypeTDengine:
		err = tdengine.SendToTDengine(context.TODO(), request.RawData, request.SQLConnectionSetting)
	default:
		err = fmt.Errorf("UnSupport DB Type")
	}

	if err != nil {
		logger.Errorf("Error to Send to SQL Server, error: %s", err.Error())
		http.Error(w, "Error to send to server", http.StatusBadRequest)
	}
}
