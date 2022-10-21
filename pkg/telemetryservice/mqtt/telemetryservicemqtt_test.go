package mqtt

import (
	"context"
	"encoding/json"
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
	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"k8s.io/kubectl/pkg/scheme"
)

const (
	unitTestServerAddress = "localhost:18927"
)

func TestMain(m *testing.M) {
	stop := make(chan struct{}, 1)
	wg := sync.WaitGroup{}
	os.Setenv("SERVER_LISTEN_PORT", ":18926")
	wg.Add(1)
	go func() {
		mockMQTTServer(stop)
		wg.Done()
	}()
	m.Run()

	stop <- struct{}{}
	os.Unsetenv("SERVER_LISTEN_PORT")
	wg.Wait()
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
				MQTTServerAddress: unitest.ToPointer("localhost" + unitTestServerAddress),
			},
			isConnected: true,
			expectedErr: "",
		},
	}

	for _, c := range testCases {
		client, err := connectToMQTT(c.setting)
		if err != nil {
			assert.Equal(t, c.expectedErr, err.Error())
			return
		}

		connected := (*client).IsConnected()
		assert.Equal(t, c.isConnected, connected)
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

func mockMQTTServer(stop <-chan struct{}) {
	lis, err := net.Listen("tcp", unitTestServerAddress)
	if err != nil {
		klog.Fatalf("Error when Listen ad %v, error: %v", unitTestServerAddress, err)
	}
	klog.Infof("mockDevice listen at %v", unitTestServerAddress)
	svr := mqtt.NewServer(lis)
	svr.Start()

	select {
	case <-svr.Done:
	case <-stop:
	case <-time.After(time.Second * 10):
		klog.Fatalf("Timeout")
	}
	lis.Close()
}

func TestHandler(t *testing.T) {
	mockServer := mockUnitTestServer(t)
	client := mockRestClient(mockServer.URL, t)

	req1 := TelemetryRequest{
		RawData: []byte("test"),
		MQTTSetting: &v1alpha1.MQTTSetting{
			MQTTTopic:         unitest.ToPointer("/test"),
			MQTTServerAddress: unitest.ToPointer(unitTestServerAddress),
		},
	}
	req2 := TelemetryRequest{
		RawData: []byte("test"),
		MQTTSetting: &v1alpha1.MQTTSetting{
			MQTTTopic:         unitest.ToPointer("/test"),
			MQTTServerAddress: unitest.ToPointer("wrong address"),
		},
	}

	requestBody1, err := json.Marshal(req1)
	assert.Nil(t, err)
	requestBody2, err := json.Marshal(req2)
	assert.Nil(t, err)

	testCases := []struct {
		name      string
		req       *rest.Request
		expectErr string
	}{
		{
			name:      "case1 without request body",
			req:       client.Post(),
			expectErr: "the server rejected our request for an unknown reason",
		}, {
			name:      "case2 correct request body but not connect to server",
			req:       client.Post().Body(requestBody1),
			expectErr: "",
		}, {
			name:      "case3 wrong request body with wrong address",
			req:       client.Post().Body(requestBody2),
			expectErr: "an error on the server (\"no servers defined to connect to\") has prevented the request from succeeding",
		},
	}
	for _, c := range testCases {
		result := c.req.Do(context.TODO())
		if err := result.Error(); err != nil {
			assert.Equal(t, c.expectErr, err.Error())
		}
	}
}

func mockRestClient(url string, t *testing.T) *rest.RESTClient {
	c, _ := rest.RESTClientFor(&rest.Config{
		Host: url,
		ContentConfig: rest.ContentConfig{
			GroupVersion:         &v1.SchemeGroupVersion,
			NegotiatedSerializer: scheme.Codecs.WithoutConversion(),
		},
		Username: "user",
		Password: "pass",
	})

	return c
}

func mockUnitTestServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", BindMQTTServicehandler)
	return httptest.NewServer(mux)
}
