package sqlserver

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/edgenesis/shifu/pkg/logger"
	"github.com/edgenesis/shifu/pkg/telemetryservice/sql/template"

	_ "github.com/microsoft/go-mssqldb"
)

type DBHelper struct {
	DB       *sql.DB
	Settings *v1alpha1.SQLConnectionSetting
}

var _ template.DBDriver = (*DBHelper)(nil)

func (db *DBHelper) SendToDB(ctx context.Context, rawData []byte) error {
	err := db.ConnectToDB(ctx)
	if err != nil {
		logger.Errorf("Error to Connect to SQL Server, error %v", err.Error())
		return err
	}

	err = db.InsertDataToDB(ctx, rawData)
	if err != nil {
		logger.Errorf("Error to Insert rawData to DB, errror: %v", err.Error())
		return err
	}

	return nil
}

func constructDBUri(sqlcs *v1alpha1.SQLConnectionSetting) string {
	return fmt.Sprintf("sqlserver://%s:%s@%s?database=%s", *sqlcs.UserName, *sqlcs.Secret, *sqlcs.ServerAddress, *sqlcs.DBName)
}

func (db *DBHelper) ConnectToDB(ctx context.Context) error {
	var err error
	sqlServerUri := constructDBUri(db.Settings)
	db.DB, err = sql.Open("sqlserver", sqlServerUri)
	logger.Infof("Try connect to sqlserver %v", *db.Settings.DBName)
	if err != nil {
		return err
	}
	return db.DB.Ping()
}

func (db *DBHelper) InsertDataToDB(ctx context.Context, rawData []byte) error {
	result, err := db.DB.Exec(fmt.Sprintf("Insert Into %s Values('%s','%s')", *db.Settings.DBTable, time.Now().Format("2006-01-02 15:04:05"), string(rawData)))
	if err != nil {
		logger.Errorf("Error to Insert RawData to db, error: %v", err)
		return err
	}

	id, err := result.RowsAffected()
	if err != nil {
		logger.Errorf("Error to get RowsAffected, error: %v", err)
		return err
	} else if id <= 0 {
		logger.Errorf("Data insert failed, for RowsAffected is equal or lower than 0")
		return errors.New("insert Failed")
	}

	logger.Infof("Successfully Insert Data %v to DB %v", string(rawData), db.Settings.DBName)
	return nil
}
