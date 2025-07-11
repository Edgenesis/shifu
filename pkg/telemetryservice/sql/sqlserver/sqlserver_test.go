package sqlserver

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/edgenesis/shifu/pkg/deviceshifu/unitest"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/edgenesis/shifu/pkg/telemetryservice/sql/template"
	"github.com/stretchr/testify/assert"
)

func TestConstructDBUri(t *testing.T) {
	testCases := []struct {
		desc   string
		Input  v1alpha1.SQLConnectionSetting
		output string
	}{
		{
			desc: "mysql test",
			Input: v1alpha1.SQLConnectionSetting{
				UserName:      unitest.ToPointer("testUser"),
				Secret:        unitest.ToPointer("testPassword"),
				ServerAddress: unitest.ToPointer("testAddress"),
				DBName:        unitest.ToPointer("testDB"),
			},
			output: "sqlserver://testUser:testPassword@testAddress?database=testDB",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			result := constructDBUri(&tC.Input)
			assert.Equal(t, tC.output, result)
		})
	}
}

func TestInsertDataToDB(t *testing.T) {
	testCases := []struct {
		desc         string
		expectSQL    string
		deviceName   string
		rawData      []byte
		dbHelper     *DBHelper
		expectResult sql.Result
		expectErr    string
		preCloseDB   bool
	}{
		{
			desc:       "testCases 1 insert Successfully",
			expectSQL:  "Insert Into testTable",
			deviceName: "testDevice",
			rawData:    []byte("testData"),
			dbHelper: &DBHelper{
				Settings: &v1alpha1.SQLConnectionSetting{
					DBName:  unitest.ToPointer("testDB"),
					DBTable: unitest.ToPointer("testTable"),
				},
			},
			expectResult: sqlmock.NewResult(1, 1),
		},
		{
			desc:       "testCases2 without DBName",
			deviceName: "testDevice",
			rawData:    []byte("testData"),
			dbHelper: &DBHelper{
				Settings: &v1alpha1.SQLConnectionSetting{
					DBName:  unitest.ToPointer("testDB"),
					DBTable: unitest.ToPointer(""),
				},
			},
			preCloseDB: true,
			expectErr:  "sql: database is closed",
		},
		{
			desc:       "testCases 3 LastInsertId = 0",
			deviceName: "testDevice",
			expectSQL:  "Insert Into testTable",
			rawData:    []byte("testData"),
			dbHelper: &DBHelper{
				Settings: &v1alpha1.SQLConnectionSetting{
					DBName:  unitest.ToPointer("testDB"),
					DBTable: unitest.ToPointer("testTable"),
				},
			},
			expectResult: sqlmock.NewResult(0, 0),
			expectErr:    "insert Failed",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			db, sm, err := sqlmock.New()
			assert.Nil(t, err)
			if tC.preCloseDB {
				_ = db.Close()
			} else {
				defer func() { _ = db.Close() }()
			}

			sm.ExpectExec(tC.expectSQL).WillReturnResult(tC.expectResult)
			helper := DBHelper{DB: db, Settings: tC.dbHelper.Settings}
			err = helper.InsertDataToDB(context.TODO(), tC.deviceName, tC.rawData)
			if tC.expectErr == "" {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
				assert.Equal(t, tC.expectErr, err.Error())
			}
		})
	}
}

func TestConnectT0DB(t *testing.T) {
	db := &DBHelper{
		Settings: &v1alpha1.SQLConnectionSetting{
			UserName:      unitest.ToPointer("testUser"),
			Secret:        unitest.ToPointer("testSecret"),
			ServerAddress: unitest.ToPointer("127.0.0.1:1234"),
			DBName:        unitest.ToPointer("testDB"),
		},
	}
	expectErr := "unable to open tcp connection with host '127.0.0.1:1234': dial tcp 127.0.0.1:1234: connect: connection refused"
	err := db.ConnectToDB(context.TODO())
	assert.Equal(t, expectErr, err.Error())
}

func TestSendToSQLServer(t *testing.T) {
	settings := &v1alpha1.SQLConnectionSetting{
		UserName:      unitest.ToPointer("testUser"),
		Secret:        unitest.ToPointer("testSecret"),
		ServerAddress: unitest.ToPointer("127.0.0.1:1234"),
		DBName:        unitest.ToPointer("testDB"),
		DBTable:       unitest.ToPointer("testTable"),
	}
	var dbDriver template.DBDriver

	expectErr := "unable to open tcp connection with host '127.0.0.1:1234': dial tcp 127.0.0.1:1234: connect: connection refused"
	dbDriver = &DBHelper{Settings: settings}
	err := dbDriver.SendToDB(context.TODO(), "testDevice", []byte("test"))
	assert.Equal(t, expectErr, err.Error())
}
