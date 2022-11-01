package mqtt

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/edgenesis/shifu/pkg/deviceshifu/unitest"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/jeffallen/mqtt"
	"github.com/stretchr/testify/assert"
	"k8s.io/klog"
)

const (
	unitTestServerAddress = "localhost:18927"
)

func TestMain(m *testing.M) {
	start := make(chan struct{}, 1)
	stop := make(chan struct{}, 1)
	wg := sync.WaitGroup{}
	os.Setenv("SERVER_LISTEN_PORT", ":18926")
	wg.Add(1)
	go func() {
		mockMQTTServer(stop, start)
		klog.Infof("Server Closed")
		wg.Done()
	}()
	<-start
	m.Run()
	stop <- struct{}{}
	wg.Wait()
	os.Unsetenv("SERVER_LISTEN_PORT")
}

func TestConnectToMQTT(t *testing.T) {
	testCases := []struct {
		name        string
		setting     *v1alpha1.MQTTSetting
		isConnected bool
		expectedErr string
	}{
		{
			name: "case1 wrong address",
			setting: &v1alpha1.MQTTSetting{
				MQTTTopic:         unitest.ToPointer("/default/topic"),
				MQTTServerAddress: unitest.ToPointer("wrong address"),
			},
			isConnected: false,
			expectedErr: "no servers defined to connect to",
		},
		{
			name: "case2 pass",
			setting: &v1alpha1.MQTTSetting{
				MQTTTopic:         unitest.ToPointer("/default/topic"),
				MQTTServerAddress: unitest.ToPointer(unitTestServerAddress),
			},
			isConnected: true,
			expectedErr: "",
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			client, err := connectToMQTT(c.setting)
			if err != nil {
				assert.Equal(t, c.expectedErr, err.Error())
				return
			}
			connected := (*client).IsConnected()
			assert.Equal(t, c.isConnected, connected)
		})
	}
}

func TestMessagePubHandler(t *testing.T) {
	messagePubHandler(nil, nil)
}

func TestConnectHandler(t *testing.T) {
	connectHandler(nil)
}

func TestConnectLostHandler(t *testing.T) {
	connectLostHandler(nil, nil)
}

func mockMQTTServer(stop <-chan struct{}, start chan<- struct{}) {
	lis, err := net.Listen("tcp", unitTestServerAddress)
	if err != nil {
		klog.Fatalf("Error when Listen ad %v, error: %v", unitTestServerAddress, err)
	}
	klog.Infof("mockDevice listen at %v", unitTestServerAddress)
	svr := mqtt.NewServer(lis)
	svr.Start()

	start <- struct{}{}
	select {
	case <-stop:
	case <-time.After(time.Second * 10):
		klog.Fatalf("Timeout")
	}
	lis.Close()
}

func TestBindMQTTServicehandler(t *testing.T) {
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
			desc:       "testCase2 wrong address",
			expectResp: "Error to connect to server\n",
			requestBody: &v1alpha1.TelemetryRequest{
				MQTTSetting: &v1alpha1.MQTTSetting{
					MQTTServerAddress: unitest.ToPointer("wrong address"),
				},
			},
		},
		{
			desc:       "testCase3 pass",
			expectResp: "",
			requestBody: &v1alpha1.TelemetryRequest{
				MQTTSetting: &v1alpha1.MQTTSetting{
					MQTTTopic:         unitest.ToPointer("/test/test"),
					MQTTServerAddress: unitest.ToPointer(unitTestServerAddress)},
				RawData: []byte("123"),
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
			BindMQTTServicehandler(rr, req)
			body, err := io.ReadAll(rr.Result().Body)
			assert.Nil(t, err)
			assert.Equal(t, tC.expectResp, string(body))

		})
	}

}
