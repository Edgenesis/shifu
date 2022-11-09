package sql

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgenesis/shifu/pkg/deviceshifu/unitest"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestBindSQLServiceHandler(t *testing.T) {
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
			desc:       "testCase2 wrong db type",
			expectResp: "Error to send to server\n",
			requestBody: &v1alpha1.TelemetryRequest{
				SQLConnectionSetting: &v1alpha1.SQLConnectionSetting{
					DBType: unitest.ToPointer(v1alpha1.DBType("wrongType")),
				},
			},
		},
		{
			desc:       "testCase3 db Type TDengine",
			expectResp: "Error to send to server\n",
			requestBody: &v1alpha1.TelemetryRequest{
				SQLConnectionSetting: &v1alpha1.SQLConnectionSetting{
					ServerAddress: unitest.ToPointer("testAddr"),
					DBType:        unitest.ToPointer(v1alpha1.DBTypeTDengine),
					UserName:      unitest.ToPointer("test"),
					Secret:        unitest.ToPointer("test"),
					DBName:        unitest.ToPointer("test"),
					DBTable:       unitest.ToPointer("test"),
				},
				RawData: []byte("test"),
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "test:80", bytes.NewBuffer([]byte("")))
			if tC.requestBody != nil {
				requestBody, err := json.Marshal(tC.requestBody)
				assert.Nil(t, err)
				req = httptest.NewRequest(http.MethodPost, "test:80", bytes.NewBuffer(requestBody))
			}

			rr := httptest.NewRecorder()
			BindSQLServiceHandler(rr, req)
			body, err := io.ReadAll(rr.Result().Body)
			assert.Nil(t, err)
			assert.Equal(t, tC.expectResp, string(body))

		})
	}

}
