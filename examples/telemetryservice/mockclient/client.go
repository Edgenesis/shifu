package main

/*
mockHTTPClient, using this file will send a message to telemetryService
*/
import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
)

var (
	targetServer         = "http://" + os.Getenv("TARGET_SERVER_ADDRESS")
	targetMqttServer     = os.Getenv("TARGET_MQTT_SERVER_ADDRESS")
	targetTDengineServer = os.Getenv("TARGET_TDENGINE_SERVER_ADDRESS")
	targetMySQLServer    = os.Getenv("TARGET_MYSQL_SERVER_ADDRESS")
	targetMSSQLServer    = os.Getenv("TARGET_SQLSERVER_SERVER_ADDRESS")
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/mqtt", sendToMQTT)
	mux.HandleFunc("/tdengine", sendToTDengine)
	mux.HandleFunc("/mysql", sendToMySQL)
	mux.HandleFunc("/sqlserver", sendToSQLServer)

	_ = http.ListenAndServe(":9090", mux)

}

func sendToMQTT(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")

	defaultTopic := "/test"
	req := &v1alpha1.TelemetryRequest{
		RawData: []byte("testData"),
		MQTTSetting: &v1alpha1.MQTTSetting{
			MQTTTopic:          &defaultTopic,
			MQTTServerAddress:  &targetMqttServer,
			MQTTServerUserName: username,
			MQTTServerPassword: password,
		},
	}
	if err := sendRequest(req, "/mqtt"); err != nil {
		http.Error(w, "send failed", http.StatusInternalServerError)
	}
}

func sendToTDengine(w http.ResponseWriter, r *http.Request) {
	req := &v1alpha1.TelemetryRequest{
		Spec: &v1alpha1.EdgeDeviceSpec{
			Sku: toPointer("testDevice"),
		},
		SQLConnectionSetting: &v1alpha1.SQLConnectionSetting{
			ServerAddress: &targetTDengineServer,
			UserName:      toPointer("root"),
			Secret:        toPointer("taosdata"),
			DBName:        toPointer("shifu"),
			DBTable:       toPointer("testsubtable"),
			DBType:        toPointer(v1alpha1.DBTypeTDengine),
		},
		RawData: []byte("testData"),
	}
	_ = sendRequest(req, "/sql")
}

func sendToMySQL(w http.ResponseWriter, r *http.Request) {
	req := &v1alpha1.TelemetryRequest{
		Spec: &v1alpha1.EdgeDeviceSpec{
			Sku: toPointer("testDevice"),
		},
		SQLConnectionSetting: &v1alpha1.SQLConnectionSetting{
			ServerAddress: &targetMySQLServer,
			UserName:      toPointer("root"),
			Secret:        toPointer(""),
			DBName:        toPointer("shifu"),
			DBTable:       toPointer("testTable"),
			DBType:        toPointer(v1alpha1.DBTypeMySQL),
		},
		RawData: []byte("testData"),
	}
	_ = sendRequest(req, "/sql")
}

func sendToSQLServer(w http.ResponseWriter, r *http.Request) {
	req := &v1alpha1.TelemetryRequest{
		Spec: &v1alpha1.EdgeDeviceSpec{
			Sku: toPointer("testDevice"),
		},
		SQLConnectionSetting: &v1alpha1.SQLConnectionSetting{
			ServerAddress: &targetMSSQLServer,
			UserName:      toPointer("sa"),
			Secret:        toPointer("YourStrong@Passw0rd"),
			DBName:        toPointer("shifu"),
			DBTable:       toPointer("testTable"),
			DBType:        toPointer(v1alpha1.DBTypeSQLServer),
		},
		RawData: []byte("testData"),
	}
	_ = sendRequest(req, "/sql")
}

func sendRequest(request *v1alpha1.TelemetryRequest, path string) error {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return err
	}

	_, err = http.DefaultClient.Post(targetServer+path, "application/json", bytes.NewBuffer(requestBody))
	return err
}

func toPointer[T any](v T) *T {
	return &v
}
