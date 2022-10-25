package sql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	_ "github.com/taosdata/driver-go/v3/taosRestful"
	"k8s.io/klog"
)

func BindSQLServiceHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		klog.Errorf("Error when Read Body from request, error: %v", err.Error())
		http.Error(w, "Error when Read Body from reuqest", http.StatusInternalServerError)
		return
	}
	var request *v1alpha1.TelemetryRequest

	err = json.Unmarshal(body, request)
	if err != nil {
		klog.Errorf("Error when Unmarshal requestBody to TelemetryRequest, error: %v", err.Error())
		http.Error(w, "Error when Unmarshal requestBody to TelemetryRequest", http.StatusBadRequest)
		return
	}

	err = sendToTDEngine(request.RawData, request.SQLConnectionSetting)
	if err != nil {
		klog.Errorf("Error when send message to TDEngine, error: %v", err.Error())
		http.Error(w, "Error when send message to TDEngine", http.StatusInternalServerError)
	}
}

func sendToTDEngine(rawData []byte, sqlcs *v1alpha1.SQLConnectionSetting) error {
	taosUri := constructTDEngineUri(sqlcs)
	taos, err := sql.Open("taosRestful", taosUri)
	if err != nil {
		klog.Error("failed to connect TDengine, err:", err)
		return err
	}
	defer taos.Close()

	// Insert Into testtable2 Using testtable TAGS('aaaaa',1) Values('2018-10-03 14:32:11',"123");
	result, err := taos.Exec(fmt.Sprintf("Insert Into %s Values('%s','%s')", *sqlcs.DBTable, time.Now().Format("2006-01-02 15:04:05"), string(rawData)))
	if err != nil {
		log.Fatalln("failed to insert, err:", err.Error())
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatalln("failed to get affected rows, err:", err)
		return err
	}
	fmt.Println("RowsAffected", rowsAffected)
	return nil
}

// constructTDEngineUri  example: root:taosdata@http(localhost:6041)/test
func constructTDEngineUri(sqlcs *v1alpha1.SQLConnectionSetting) string {
	return fmt.Sprintf("%s:%s@http(%s)/%s", *sqlcs.UserName, *sqlcs.Secret, *sqlcs.ServerAddress, *sqlcs.DBName)
}
