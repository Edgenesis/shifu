package tdengine

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/edgenesis/shifu/pkg/deviceshifu/unitest"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestConstructTDengineUri(t *testing.T) {
	testCases := []struct {
		desc   string
		Input  v1alpha1.SQLConnectionSetting
		output string
	}{
		{
			desc: "test",
			Input: v1alpha1.SQLConnectionSetting{
				UserName:      unitest.ToPointer("testUser"),
				Secret:        unitest.ToPointer("testPassword"),
				ServerAddress: unitest.ToPointer("testAddress"),
				DBName:        unitest.ToPointer("testDB"),
			},
			output: "testUser:testPassword@http(testAddress)/testDB",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			result := constructTDengineUri(&tC.Input)
			assert.Equal(t, tC.output, result)
		})
	}
}

func TestCreateDBAndTable(t *testing.T) {
	testCases := []struct {
		desc     string
		dbHelper *DBHelper
	}{
		{
			desc: "test",
			dbHelper: &DBHelper{
				Settings: &v1alpha1.SQLConnectionSetting{
					DBName:  unitest.ToPointer("testDB"),
					DBTable: unitest.ToPointer("testTable"),
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			db, sm, err := sqlmock.New()
			assert.Nil(t, err)

			sm.ExpectExec("CREATE DATABASE IF NOT EXISTS").WillReturnResult(sqlmock.NewResult(0, 0))
			sm.ExpectExec("CREATE STABLE IF NOT EXISTS").WillReturnResult(sqlmock.NewResult(0, 0))
			helper := DBHelper{DB: db, Settings: tC.dbHelper.Settings}
			err = helper.createDBAndTable(context.Background())
			assert.Nil(t, err)
		})
	}
}

// The SQL supported by TDengine has some differences from classic SQL, so we cannot use sqlmock for testing
func TestPost(t *testing.T) {
	testCases := []struct {
		desc         string
		rawData      []byte
		deviceInfo   *DeviceInfo
		dbHelper     *DBHelper
		expectSQL    string
		expectResult sql.Result
		expectErr    string
		preCloseDB   bool
	}{
		{
			desc:      "testCases 1 insert Successfully",
			expectSQL: "INSERT INTO testDB.device_no_1 USING testDB.testTable",
			rawData:   []byte("testData"),
			dbHelper: &DBHelper{
				Settings: &v1alpha1.SQLConnectionSetting{
					DBName:  unitest.ToPointer("testDB"),
					DBTable: unitest.ToPointer("testTable"),
				},
			},
			expectResult: sqlmock.NewResult(1, 1),
			deviceInfo: &DeviceInfo{
				Name: "device_no_1",
				Tag:  "open device",
			},
		},
		{
			desc:    "testCases2 without DBName",
			rawData: []byte("testData"),
			dbHelper: &DBHelper{
				Settings: &v1alpha1.SQLConnectionSetting{
					DBName:  unitest.ToPointer("testDB"),
					DBTable: unitest.ToPointer(""),
				},
			},
			preCloseDB: true,
			expectErr:  "sql: database is closed",
			deviceInfo: &DeviceInfo{
				Name: "device_no_2",
				Tag:  "close device",
			},
		},
		{
			desc:      "testCases 3 LastInsertId = 0",
			expectSQL: "INSERT INTO testDB.device_no_3 USING testDB.testTable",
			rawData:   []byte("testData"),
			dbHelper: &DBHelper{
				Settings: &v1alpha1.SQLConnectionSetting{
					DBName:  unitest.ToPointer("testDB"),
					DBTable: unitest.ToPointer("testTable"),
				},
			},
			expectResult: sqlmock.NewResult(0, 0),
			expectErr:    "insert Failed",
			deviceInfo: &DeviceInfo{
				Name: "device_no_3",
				Tag:  "restart device",
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			db, sm, err := sqlmock.New()
			assert.Nil(t, err)
			if tC.preCloseDB {
				db.Close()
			} else {
				defer db.Close()
			}

			sm.ExpectExec("CREATE DATABASE IF NOT EXISTS").WillReturnResult(sqlmock.NewResult(0, 0))
			sm.ExpectExec("CREATE STABLE IF NOT EXISTS").WillReturnResult(sqlmock.NewResult(0, 0))
			sm.ExpectExec(tC.expectSQL).WillReturnResult(tC.expectResult)
			helper := DBHelper{DB: db, Settings: tC.dbHelper.Settings}
			err = helper.post(context.TODO(), tC.rawData, tC.deviceInfo)
			if tC.expectErr == "" {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
				assert.Equal(t, tC.expectErr, err.Error())
			}
		})
	}
}

// just for cover the code
func TestConnectTdengine(t *testing.T) {
	db := &DBHelper{
		Settings: &v1alpha1.SQLConnectionSetting{
			UserName:      unitest.ToPointer("testUser"),
			Secret:        unitest.ToPointer("testSecret"),
			ServerAddress: unitest.ToPointer("testAddress"),
			DBName:        unitest.ToPointer("testDB"),
		},
	}
	err := db.connectToTDengine(context.TODO())
	assert.Nil(t, err)
}

func TestSendToTDengine(t *testing.T) {
	settings := &v1alpha1.SQLConnectionSetting{
		UserName:      unitest.ToPointer("testUser"),
		Secret:        unitest.ToPointer("testSecret"),
		ServerAddress: unitest.ToPointer("1.2.3.4"),
		DBName:        unitest.ToPointer("testDB"),
		DBTable:       unitest.ToPointer("testTable"),
	}
	expectErr := "invalid DSN: network address not terminated (missing closing brace)"
	err := SendToTDengine(context.TODO(), []byte("test"), settings, nil)
	assert.Equal(t, expectErr, err.Error())
}

func TestQuery(t *testing.T) {
	testCases := []struct {
		desc     string
		dbHelper *DBHelper
		query    string
	}{
		{
			desc: "query successfully",
			dbHelper: &DBHelper{
				Settings: &v1alpha1.SQLConnectionSetting{
					DBName:  unitest.ToPointer("testDB"),
					DBTable: unitest.ToPointer("testTable"),
				},
			},
			query: "SELECT ts, data, tg, devicename FROM testDB.testTable WHERE devicename='device_1'",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			db, sm, err := sqlmock.New()
			assert.Nil(t, err)

			sm.ExpectQuery("SELECT ts, data, tg, devicename").WillReturnRows(sqlmock.NewRows([]string{"ts", "data", "tg", "devicename"}))
			helper := DBHelper{DB: db, Settings: tC.dbHelper.Settings}
			_, err = helper.query(tC.query)
			assert.Nil(t, err)
		})
	}
}

func TestQueryFromDeviceName(t *testing.T) {
	testCases := []struct {
		desc        string
		dbHelper    *DBHelper
		deviceName  string
		expectedErr string
	}{
		{
			desc: "query successfully",
			dbHelper: &DBHelper{
				Settings: &v1alpha1.SQLConnectionSetting{
					DBName:  unitest.ToPointer("testDB"),
					DBTable: unitest.ToPointer("testTable"),
				},
			},
			deviceName: "device_no_1",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			db, sm, err := sqlmock.New()
			assert.Nil(t, err)

			sm.ExpectQuery("SELECT ts, data, tg FROM testDB.testTable WHERE devicename=").WillReturnRows(sqlmock.NewRows([]string{"ts", "data", "tg", "devicename"}))
			helper := DBHelper{DB: db, Settings: tC.dbHelper.Settings}
			_, err = helper.queryFromDeviceName(tC.deviceName)
			if tC.expectedErr == "" {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
				assert.Equal(t, tC.expectedErr, err.Error())
			}
		})
	}
}

func TestQueryFromTag(t *testing.T) {
	testCases := []struct {
		desc        string
		dbHelper    *DBHelper
		Tag         string
		expectedErr string
	}{
		{
			desc: "query successfully",
			dbHelper: &DBHelper{
				Settings: &v1alpha1.SQLConnectionSetting{
					DBName:  unitest.ToPointer("testDB"),
					DBTable: unitest.ToPointer("testTable"),
				},
			},
			Tag: "open service",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			db, sm, err := sqlmock.New()
			assert.Nil(t, err)

			sm.ExpectQuery("SELECT ts, data, tg FROM testDB.testTable WHERE tg").WillReturnRows(sqlmock.NewRows([]string{"ts", "data", "tg", "devicename"}))
			helper := DBHelper{DB: db, Settings: tC.dbHelper.Settings}
			_, err = helper.queryFromTag(tC.Tag)
			if tC.expectedErr == "" {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
				assert.Equal(t, tC.expectedErr, err.Error())
			}
		})
	}
}

func TestQueryFromTime(t *testing.T) {
	testCases := []struct {
		desc        string
		dbHelper    *DBHelper
		startTime   time.Time
		endTime     time.Time
		expectedErr string
	}{
		{
			desc: "query successfully",
			dbHelper: &DBHelper{
				Settings: &v1alpha1.SQLConnectionSetting{
					DBName:  unitest.ToPointer("testDB"),
					DBTable: unitest.ToPointer("testTable"),
				},
			},
			startTime: time.Now().Add(-3 * time.Hour),
			endTime:   time.Now().Add(-time.Hour),
		},
		{
			desc: "query failed",
			dbHelper: &DBHelper{
				Settings: &v1alpha1.SQLConnectionSetting{
					DBName:  unitest.ToPointer("testDB"),
					DBTable: unitest.ToPointer("testTable"),
				},
			},
			startTime:   time.Now().Add(-time.Hour),
			endTime:     time.Now().Add(-2 * time.Hour),
			expectedErr: "start time is after the end time",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			db, sm, err := sqlmock.New()
			assert.Nil(t, err)

			sm.ExpectQuery("SELECT ts, data, tg FROM testDB.testTable WHERE ts").WillReturnRows(sqlmock.NewRows([]string{"ts", "data", "tg", "devicename"}))
			helper := DBHelper{DB: db, Settings: tC.dbHelper.Settings}
			_, err = helper.queryFromTime(tC.startTime, tC.endTime)
			if tC.expectedErr == "" {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
				assert.Equal(t, tC.expectedErr, err.Error())
			}
		})
	}
}
