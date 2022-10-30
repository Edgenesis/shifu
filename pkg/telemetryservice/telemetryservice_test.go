package telemetryservice

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/edgenesis/shifu/pkg/deviceshifu/unitest"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

const (
	unitTestServerAddress = "localhost:18927"
)

func TestStart(t *testing.T) {
	testCases := []struct {
		name     string
		mux      *http.ServeMux
		stopChan chan struct{}
		addr     string
	}{
		{
			name:     "case1 pass",
			mux:      http.NewServeMux(),
			addr:     unitTestServerAddress,
			stopChan: make(chan struct{}, 1),
		}, {
			name:     "case2 without mux",
			addr:     unitTestServerAddress,
			stopChan: make(chan struct{}, 1),
		},
	}

	for _, c := range testCases {
		go func() {
			time.Sleep(time.Microsecond * 100)
			c.stopChan <- struct{}{}
		}()
		err := Start(c.stopChan, c.mux, c.addr)
		assert.Nil(t, err)
	}
}

func TestNew(t *testing.T) {
	stop := make(chan struct{}, 1)
	go func() {
		time.After(time.Millisecond * 100)
		stop <- struct{}{}
	}()
	New(stop)
}

// target Url is unused but it must have a port
func TestMatchHandler(t *testing.T) {
	testCases := []struct {
		desc        string
		requestBody *v1alpha1.TelemetryRequest
		expectResp  string
	}{
		{
			desc:       "testCase1 RequestBody is not a JSON",
			expectResp: "unexpected end of JSON input\n",
		},
		{
			desc: "testCase2 MQTT Protocol",
			requestBody: &v1alpha1.TelemetryRequest{
				MQTTSetting: &v1alpha1.MQTTSetting{
					MQTTServerAddress: unitest.ToPointer("wrong address")},
			},
			expectResp: "no servers defined to connect to\n",
		},
		{
			desc: "testCase3 SQL Protocol",
			requestBody: &v1alpha1.TelemetryRequest{
				SQLConnectionSetting: &v1alpha1.SQLConnectionSetting{
					DBType: unitest.ToPointer(v1alpha1.DBType("default")),
				},
			},
			expectResp: "UnSupport DB Type\n",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "test:80", bytes.NewBuffer([]byte{}))
			if tC.requestBody != nil {
				requestBody, err := json.Marshal(tC.requestBody)
				assert.Nil(t, err)
				req = httptest.NewRequest(http.MethodPost, "test:80", bytes.NewBuffer(requestBody))
			}

			rr := httptest.NewRecorder()
			matchHandler(rr, req)
			body, err := io.ReadAll(rr.Result().Body)
			assert.Nil(t, err)
			assert.Equal(t, tC.expectResp, string(body))

		})
	}
}
