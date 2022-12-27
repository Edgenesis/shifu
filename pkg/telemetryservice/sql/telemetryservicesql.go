package sql

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/edgenesis/shifu/pkg/telemetryservice/utils"
	"go.uber.org/zap"
	"io"
	"net/http"

	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/edgenesis/shifu/pkg/telemetryservice/sql/tdengine"
	"k8s.io/klog"
)

var zlog *zap.SugaredLogger

func init() {
	logger, _ := zap.NewProduction()
	zlog = logger.Sugar()
}

func BindSQLServiceHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		klog.Errorf("Error when Read Data From Body, error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	klog.Infof("requestBody: %s", string(body))
	request := v1alpha1.TelemetryRequest{}

	err = json.Unmarshal(body, &request)
	if err != nil {
		klog.Errorf("Error to Unmarshal request body to struct")
		http.Error(w, "unexpected end of JSON input", http.StatusBadRequest)
		return
	}

	InjectSecret(request.SQLConnectionSetting)

	switch *request.SQLConnectionSetting.DBType {
	case v1alpha1.DBTypeTDengine:
		err = tdengine.SendToTDengine(context.TODO(), request.RawData, request.SQLConnectionSetting)
	default:
		err = fmt.Errorf("UnSupport DB Type")
	}

	if err != nil {
		klog.Errorf("Error to Send to SQL Server, error: %s", err.Error())
		http.Error(w, "Error to send to server", http.StatusBadRequest)
	}
}

func InjectSecret(setting *v1alpha1.SQLConnectionSetting) {
	if setting == nil {
		zlog.Warnf("empty telemetry service setting.")
		return
	}
	pwd, err := utils.GetPasswordFromSecret(*setting.Secret)
	if err != nil {
		zlog.Errorf("unable to get secret for telemetry %v, error: %v", *setting.Secret, err)
		return
	}
	*setting.Secret = pwd
	zlog.Infof("SQLSetting.Secret load from secret")
}
