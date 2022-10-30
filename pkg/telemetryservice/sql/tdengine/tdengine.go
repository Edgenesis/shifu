package tdengine

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	_ "github.com/taosdata/driver-go/v3/taosRestful"
	"k8s.io/klog"
)

type DBHelper struct {
	DB       *sql.DB
	Settings *v1alpha1.SQLConnectionSetting
}

func SendToTDEngine(ctx context.Context, rawData []byte, sqlcs *v1alpha1.SQLConnectionSetting) error {
	db := &DBHelper{Settings: sqlcs}

	err := db.connectToTDEngine(ctx)
	if err != nil {
		klog.Errorf("Error to Connect to tdengine, error %v", err.Error())
		return err
	}

	err = db.insertDataToDB(ctx, rawData)
	if err != nil {
		klog.Errorf("Error to Insert rawData to DB, errror: %v", err.Error())
		return err
	}

	return nil
}

func (db *DBHelper) connectToTDEngine(ctx context.Context) error {
	var err error
	taosUri := constructTDEngineUri(db.Settings)
	db.DB, err = sql.Open("taosRestful", taosUri)
	klog.Infof("Try connect to tdengine %v", *db.Settings.DBName)
	return err
}

func (db *DBHelper) insertDataToDB(ctx context.Context, rawData []byte) error {
	result, err := db.DB.Exec(fmt.Sprintf("Insert Into %s Values('%s','%s')", *db.Settings.DBTable, time.Now().Format("2006-01-02 15:04:05"), string(rawData)))
	if err != nil {
		klog.Errorf("Error to Insert RawData to db, error: %v", err)
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		klog.Errorf("Error to get LastInsertId, error: %v", err)
		return err
	} else if id <= 0 {
		klog.Errorf("Data insert failed, for LastInsertId is equal or lower than 0")
		return errors.New("insert Failed")
	}

	klog.Infof("Successfully Insert Data %v to DB %v", string(rawData), db.Settings.DBName)
	return nil
}

// constructTDEngineUri  example: root:taosdata@http(localhost:6041)/test
func constructTDEngineUri(sqlcs *v1alpha1.SQLConnectionSetting) string {
	return fmt.Sprintf("%s:%s@http(%s)/%s", *sqlcs.UserName, *sqlcs.Secret, *sqlcs.ServerAddress, *sqlcs.DBName)
}
