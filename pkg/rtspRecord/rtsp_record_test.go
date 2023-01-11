package rtspRecord

import (
	"bytes"
	"encoding/json"
	"github.com/edgenesis/shifu/pkg/telemetryservice/utils"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// no mock rtsp server
func TestRecord(t *testing.T) {
	// won't persist the map
	InitPersistMap("")
	VideoSavePath = "/tmp"
	const testNamespace = "shifu-app"
	client := testclient.NewSimpleClientset(&v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: testNamespace,
		},
		Data: map[string][]byte{
			"username": []byte("admin"),
			"password": []byte("password"),
		},
	})
	utils.SetClient(client, testNamespace)
	testCases := []struct {
		desc          string
		requestBodies []any
		specCodes     []int
	}{
		{
			desc: "testCase0 valid request but no rtsp server",
			requestBodies: []any{
				&RegisterRequest{
					DeviceName:    "test",
					SecretName:    "test-secret",
					ServerAddress: "address:12345/capture",
					Record:        true,
				},
				&UnregisterRequest{
					DeviceName: "test",
				},
			},
			specCodes: []int{http.StatusOK, http.StatusOK},
		},
		{
			desc: "testCase1 valid request with update but no rtsp server",
			requestBodies: []any{
				&RegisterRequest{
					DeviceName:    "test",
					SecretName:    "test-secret",
					ServerAddress: "address:12345/capture",
					Record:        true,
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
					SecretName:    "test-secret",
					ServerAddress: "address:12345/capture",
					Record:        true,
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
