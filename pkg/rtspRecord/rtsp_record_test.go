package rtspRecord

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRecord(t *testing.T) {
	testCases := []struct {
		desc          string
		requestBodies []any
		specCodes     []int
	}{
		{
			desc: "testCase0 valid request",
			requestBodies: []any{
				&RegisterRequest{
					DeviceName: "test",
					SecretName: "no",
					Username:   "admin",
					Password:   "HikVQDRQL",
					// TODO: need to change the url to mock device
					ServerAddress: "bj-hikcamera-01.saifai.cn:39999/capture",
					Recoding:      true,
					OutDir:        "/tmp",
				},
				&UnregisterRequest{
					DeviceName: "test",
				},
			},
			specCodes: []int{http.StatusOK, http.StatusOK},
		},
		{
			desc: "testCase1 valid request with update",
			requestBodies: []any{
				&RegisterRequest{
					DeviceName:    "test",
					SecretName:    "no",
					Username:      "admin",
					Password:      "HikVQDRQL",
					ServerAddress: "bj-hikcamera-01.saifai.cn:39999/capture",
					Recoding:      true,
					OutDir:        "/tmp",
				},
				&UpdateRequest{
					DeviceName: "test",
					Record:     false,
				},
				&UpdateRequest{
					DeviceName: "test",
					Record:     true,
				},
				&UnregisterRequest{
					DeviceName: "test",
				},
			},
			specCodes: []int{http.StatusOK, http.StatusOK, http.StatusOK, http.StatusOK},
		},
		{
			desc: "testCase2 device not found",
			requestBodies: []any{
				&RegisterRequest{
					DeviceName:    "test",
					SecretName:    "no",
					Username:      "admin",
					Password:      "HikVQDRQL",
					ServerAddress: "bj-hikcamera-01.saifai.cn:39999/capture",
					Recoding:      true,
					OutDir:        "/tmp",
				},
				&UnregisterRequest{
					DeviceName: "test-2",
				},
				&UnregisterRequest{
					DeviceName: "test",
				},
			},
			specCodes: []int{http.StatusOK, http.StatusBadRequest, http.StatusOK},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			for i, body := range tC.requestBodies {
				req := httptest.NewRequest(http.MethodGet, "test:80", bytes.NewBuffer([]byte("")))
				if body != nil {
					requestBody, err := json.Marshal(body)
					assert.Nil(t, err)
					req = httptest.NewRequest(http.MethodGet, "test:80", bytes.NewBuffer(requestBody))
				}
				rr := httptest.NewRecorder()
				switch b := body.(type) {
				case *RegisterRequest:
					Register(rr, req)
					time.Sleep(5 * time.Second)
				case *UnregisterRequest:
					Unregister(rr, req)
				case *UpdateRequest:
					Update(rr, req)
					if b.Record {
						time.Sleep(5 * time.Second)
					}
				}
				assert.Equal(t, tC.specCodes[i], rr.Result().StatusCode)
			}
		})
	}
}
