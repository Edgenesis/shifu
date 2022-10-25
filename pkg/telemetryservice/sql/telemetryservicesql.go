package sql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	_ "github.com/taosdata/driver-go/v3/taosRestful"
	"k8s.io/klog"
)

func BindSQLServiceHandler(request v1alpha1.TelemetryRequest) error {
	err := sendToTDEngine(request.RawData, request.SQLConnectionSetting)
	if err != nil {
		klog.Errorf("Error when send message to TDEngine, error: %v", err.Error())
		return err
	}
	return nil
}

func sendToTDEngine(rawData []byte, sqlcs *v1alpha1.SQLConnectionSetting) error {
	taosUri := constructTDEngineUri(sqlcs)
	taos, err := sql.Open("taosRestful", taosUri)
	if err != nil {
		klog.Errorf("Failed to connect TDengine, err: %v", err.Error())
		return err
	}
	defer taos.Close()

	// Insert Into testtable2 Using testtable TAGS('aaaaa',1) Values('2018-10-03 14:32:11',"123");
	result, err := taos.Exec(fmt.Sprintf("Insert Into %s Values('%s','%s')", *sqlcs.DBTable, time.Now().Format("2006-01-02 15:04:05"), string(rawData)))
	if err != nil {
		klog.Errorf("failed to insert, err: %v", err.Error())
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		klog.Errorf("failed to get affected rows, err: %v", err.Error())
		return err
	}
	klog.Infof("RowsAffected %v", rowsAffected)
	klog.Infof("successfully to write %s to table %s", string(rawData), *sqlcs.DBTable)
	return nil
}

// constructTDEngineUri  example: root:taosdata@http(localhost:6041)/test
func constructTDEngineUri(sqlcs *v1alpha1.SQLConnectionSetting) string {
	return fmt.Sprintf("%s:%s@http(%s)/%s", *sqlcs.UserName, *sqlcs.Secret, *sqlcs.ServerAddress, *sqlcs.DBName)
}
