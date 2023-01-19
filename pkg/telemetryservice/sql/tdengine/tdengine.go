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
	Timestamp  time.Time `json:"ts"`         // event time stamp
	Data       string    `json:"data"`       // event content
	Tag        string    `json:"tg"`         // event categories
}

func SendToTDengine(ctx context.Context, rawData []byte, sqlcs *v1alpha1.SQLConnectionSetting, header map[string][]string) error {
	db := &DBHelper{Settings: sqlcs}

	err := db.connectToTDengine(ctx)
	if err != nil {
		logger.Errorf("Error to Connect to tdengine, error %s", err.Error())
		return err
	}

	deviceInfo := &DeviceInfo{}
	if name, exists := header[DeviceNameHeaderField]; exists {
		deviceInfo.Name = name[0]
	} else {
		logger.Infof("Error to get device name from http header")
		deviceInfo.Name = "default_device_name"
	}
	if tag, exists := header[EventTagHeaderField]; exists {
		deviceInfo.Tag = tag[0]
	} else {
		logger.Infof("Error to get device tag from http header")
		deviceInfo.Tag = "default_event_tag"
	}

	err = db.post(ctx, rawData, deviceInfo)
	if err != nil {
		logger.Errorf("Error to Insert rawData to DB, error: %s", err.Error())
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

// Create Database and STable if not exists
func (db *DBHelper) createDBAndTable(ctx context.Context) error {
	// create Database
	_, err := db.DB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", *db.Settings.DBName))
	if err != nil {
		return err
	}
	// create STable, binary size is customized
	_, err = db.DB.Exec(fmt.Sprintf("CREATE STABLE IF NOT EXISTS %s.%s (ts TIMESTAMP, data BINARY(1024), tg BINARY(64)) tags (devicename binary(64))", *db.Settings.DBName, *db.Settings.DBTable))
	if err != nil {
		return err
	}
	return nil
}

func (db *DBHelper) post(ctx context.Context, rawData []byte, deviceInfo *DeviceInfo) error {
	if err := db.createDBAndTable(ctx); err != nil {
		logger.Errorf("Error to create database or table, error: %v", err)
	}
	// STable is set in the yaml file, device name is used as Table name and Tag.
	insert := fmt.Sprintf("INSERT INTO %s.%s USING %s.%s TAGS('%s') VALUES('%s','%s','%s')", *db.Settings.DBName, deviceInfo.Name, *db.Settings.DBName, *db.Settings.DBTable, deviceInfo.Name, time.Now().Format("2006-01-02 15:04:05"), string(rawData), deviceInfo.Tag)
	result, err := db.DB.Exec(insert)
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

func (db *DBHelper) query(querySql string) ([]*StructureData, error) {
	StructureDatas := make([]*StructureData, 0)
	rows, err := db.DB.Query(querySql)
	if err != nil {
		logger.Errorf("Error to query data from tdengine, sql code: %s", querySql)
		return nil, err
	}
	for rows.Next() {
		s := &StructureData{}
		err := rows.Scan(s.Timestamp, s.Data, s.Tag)
		if err != nil {
			logger.Errorf("Error to scan result into structureData")
			return nil, err
		}
		StructureDatas = append(StructureDatas, s)
	}
	return StructureDatas, nil
}

func (db *DBHelper) queryFromDeviceName(devicename string) ([]*StructureData, error) {
	querySql := fmt.Sprintf("SELECT ts, data, tg FROM %s.%s WHERE devicename='%s'", *db.Settings.DBName, *db.Settings.DBTable, devicename)
	return db.query(querySql)
}

func (db *DBHelper) queryFromTag(tag string) ([]*StructureData, error) {
	querySql := fmt.Sprintf("SELECT ts, data, tg FROM %s.%s WHERE tg='%s'", *db.Settings.DBName, *db.Settings.DBTable, tag)
	return db.query(querySql)
}

func (db *DBHelper) queryFromTime(start, end time.Time) ([]*StructureData, error) {
	if start.After(end) {
		return nil, errors.New("start time is after the end time")
	}
	querySql := fmt.Sprintf("SELECT ts, data, tg FROM %s.%s WHERE ts>'%s' AND ts<'%s'", *db.Settings.DBName, *db.Settings.DBTable, start.Format("2006-01-02 15:04:05"), end.Format("2006-01-02 15:04:05"))
	return db.query(querySql)
}
