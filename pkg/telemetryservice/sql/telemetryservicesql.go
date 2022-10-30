package sql

import (
	"context"
	"fmt"

	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/edgenesis/shifu/pkg/telemetryservice/sql/tdengine"
)

func BindSQLServiceHandler(ctx context.Context, request v1alpha1.TelemetryRequest) error {
	var err error
	switch *request.SQLConnectionSetting.DBType {
	case v1alpha1.DBTypeTDEngine:
		err = tdengine.SendToTDEngine(ctx, request.RawData, request.SQLConnectionSetting)
	default:
		err = fmt.Errorf("UnSupport DB Type")
	}

	return err
}
