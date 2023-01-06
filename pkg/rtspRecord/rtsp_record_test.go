package rtspRecord

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegister(t *testing.T) {
	testCases := []struct {
		desc        string
		requestBody *RegisterRequest
	}{
		{
			desc: "testCase0 valid request",
			requestBody: &RegisterRequest{
				DeviceName:    "test",
				SecretName:    "no",
				Username:      "admin",
				Password:      "HikVQDRQL",
				ServerAddress: "bj-hikcamera-01.saifai.cn:39999/capture",
				Recoding:      true,
				OutputPath:    "/Users/faraway/Downloads/output.mp4",
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "test:80", bytes.NewBuffer([]byte("")))
			if tC.requestBody != nil {
				requestBody, err := json.Marshal(tC.requestBody)
				assert.Nil(t, err)
				req = httptest.NewRequest(http.MethodGet, "test:80", bytes.NewBuffer(requestBody))
			}

			rr := httptest.NewRecorder()
			Register(rr, req)
		})
	}
}
