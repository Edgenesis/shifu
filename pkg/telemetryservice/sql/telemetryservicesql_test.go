package sql

import (
	"testing"

	"github.com/edgenesis/shifu/pkg/deviceshifu/unitest"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
)

func TestSendToTDEngine(t *testing.T) {
	t.Skip()
	rawData := []byte("Hello")
	sqlcs := &v1alpha1.SQLConnectionSetting{
		ServerAddress: unitest.ToPointer("192.168.14.163:6041"),
		UserName:      unitest.ToPointer("root"),
		Secret:        unitest.ToPointer("taosdata"),
		DBName:        unitest.ToPointer("shifu"),
		DBTable:       unitest.ToPointer("testTable2"),
	}
	err := sendToTDEngine(rawData, sqlcs)
	if err != nil {
		panic(err)
	}
}
