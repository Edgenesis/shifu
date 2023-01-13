package tdengine

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/edgenesis/shifu/pkg/logger"
	_ "github.com/taosdata/driver-go/v3/taosRestful"
)

const (
	DeviceNameHeaderField = "device_name"
	EventTagHeaderField   = "event_tag"
)

type DBHelper struct {
	DB       *sql.DB
	Settings *v1alpha1.SQLConnectionSetting
}

type DeviceInfo struct {
	Name string // device name
	Tag  string // device tag
}

type StructureData struct {
	DeviceName string    `json:"deviceName"` // device name
	Timestamp  time.Time `json:"timestamp"`  // event time stamp
	Data       string    `json:"data"`       // event content
	Tag        string    `json:"tag"`        // event categories
}

func SendToTDengine(ctx context.Context, rawData []byte, sqlcs *v1alpha1.SQLConnectionSetting, header map[string][]string) error {
	db := &DBHelper{Settings: sqlcs}

	err := db.connectToTDengine(ctx)
	if err != nil {
		logger.Errorf("Error to Connect to tdengine, error %v", err.Error())
		return err
	}

	deviceInfo := &DeviceInfo{}
	if name, exists := header[DeviceNameHeaderField]; exists {
		deviceInfo.Name = name[0]
	} else {
		logger.Errorf("Error to get device name from http header")
	}
	if tag, exists := header[EventTagHeaderField]; exists {
		deviceInfo.Tag = tag[0]
	} else {
		logger.Errorf("Error to get device tag from http header")
	}

	err = db.insertDataToDB(ctx, rawData, deviceInfo)
	if err != nil {
		logger.Errorf("Error to Insert rawData to DB, errror: %v", err.Error())
		return err
	}

	return nil
}

func (db *DBHelper) connectToTDengine(ctx context.Context) error {
	var err error
	taosUri := constructTDengineUri(db.Settings)
	db.DB, err = sql.Open("taosRestful", taosUri)
	logger.Infof("Try connect to tdengine %v", *db.Settings.DBName)
	return err
}

func (db *DBHelper) createDBAndTable(ctx context.Context) error {
	// create Database
	_, err := db.DB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", *db.Settings.DBName))
	if err != nil {
		return err
	}
	// create STable, binary size is customized
	_, err = db.DB.Exec(fmt.Sprintf("CREATE STABLE IF NOT EXISTS %s.%s (timestamp TIMESTAMP, data BINARY(1024), tag BINARY(64)) tags (deviceName binary(64))", *db.Settings.DBName, *db.Settings.DBTable))
	if err != nil {
		return err
	}
	return nil
}

func (db *DBHelper) insertDataToDB(ctx context.Context, rawData []byte, deviceInfo *DeviceInfo) error {
	if err := db.createDBAndTable(ctx); err != nil {
		logger.Errorf("Error to create database or table, error: %v", err)
	}
	// STable is set in the yaml file, device name is used as Table name and Tag.
	result, err := db.DB.Exec(fmt.Sprintf("INSERT INTO %s.%s USING %s.%s TAGS('%s') VALUES('%s','%s','%s')", *db.Settings.DBName, deviceInfo.Name, *db.Settings.DBName, *db.Settings.DBTable, deviceInfo.Name, time.Now().Format("2006-01-02 15:04:05"), string(rawData), deviceInfo.Tag))
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

// constructTDengineUri  example: root:taosdata@http(localhost:6041)/test
func constructTDengineUri(sqlcs *v1alpha1.SQLConnectionSetting) string {
	return fmt.Sprintf("%s:%s@http(%s)/%s", *sqlcs.UserName, *sqlcs.Secret, *sqlcs.ServerAddress, *sqlcs.DBName)
}

// example: select timestamp, data, tag, deviceName from $(STable) where deviceName=$(deviceName)
func (db *DBHelper) query(sql string) ([]*StructureData, error) {
	StructureDatas := make([]*StructureData, 0)
	rows, err := db.DB.Query(sql)
	if err != nil {
		logger.Errorf("Error to query data from tdengine, sql code: %s", sql)
		return nil, err
	}
	for rows.Next() {
		s := &StructureData{}
		err := rows.Scan(s.Timestamp, s.Data, s.Tag, s.DeviceName)
		if err != nil {
			logger.Errorf("Error to scan result into structureData")
			return nil, err
		}
		StructureDatas = append(StructureDatas, s)
	}
	return StructureDatas, nil
}
