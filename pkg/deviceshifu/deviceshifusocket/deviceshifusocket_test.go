package deviceshifusocket

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/edgenesis/shifu/pkg/deviceshifu/unitest"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

func TestMain(m *testing.M) {
	err := GenerateConfigMapFromSnippet(MockDeviceCmStr, MockDeviceConfigFolder)
	if err != nil {
		klog.Errorf("error when generateConfigmapFromSnippet,err: %v", err)
		os.Exit(-1)
	}
	m.Run()
	err = os.RemoveAll(MockDeviceConfigPath)
	if err != nil {
		klog.Fatal(err)
	}
}

func TestStart(t *testing.T) {
	deviceShifuMetadata := &deviceshifubase.DeviceShifuMetaData{
		Name:           "TestStart",
		ConfigFilePath: "etc/edgedevice/config",
		KubeConfigPath: deviceshifubase.DeviceKubeconfigDoNotLoadStr,
	}

	mockds, err := New(deviceShifuMetadata)
	if err != nil {
		t.Errorf("Failed creating new deviceshifu")
	}

	if err := mockds.Start(wait.NeverStop); err != nil {
		t.Errorf("DeviceShifu.Start failed due to: %v", err.Error())
	}

	if err := mockds.Stop(); err != nil {
		t.Errorf("unable to stop mock deviceShifu, error: %+v", err)
	}
}

func TestDeviceHealthHandler(t *testing.T) {
	deviceShifuMetadata := &deviceshifubase.DeviceShifuMetaData{
		Name:           "TeststartHTTPServer",
		ConfigFilePath: "etc/edgedevice/config",
		KubeConfigPath: deviceshifubase.DeviceKubeconfigDoNotLoadStr,
	}

	mockds, err := New(deviceShifuMetadata)
	if err != nil {
		t.Errorf("Failed creating new deviceshifu")
	}

	if err := mockds.Start(wait.NeverStop); err != nil {
		t.Errorf("DeviceShifu.Start failed due to: %v", err.Error())
	}

	resp, err := unitest.RetryAndGetHTTP("http://localhost:8080/health", 3)
	if err != nil {
		t.Errorf("HTTP GET returns an error %v", err.Error())
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("unable to read response body, error: %v", err.Error())
	}

	if string(body) != deviceshifubase.DeviceIsHealthyStr {
		t.Errorf("%+v", body)
	}

	if err := mockds.Stop(); err != nil {
		t.Errorf("unable to stop mock deviceShifu, error: %+v", err)
	}
}

func TestDecodeCommand(t *testing.T) {
	input := "1230000abc"
	var outputHex = []byte{18, 48, 0, 10, 188}

	output, err := decodeCommand(input, v1alpha1.HEX)
	if err != nil {
		t.Errorf("Error when decodeCommand on test1, error:%v", err)
	}
	if !reflect.DeepEqual(output, outputHex) {
		t.Errorf("not match with current output, output: %v", output)
	}

	output, err = decodeCommand(input, v1alpha1.UTF8)
	if err != nil {
		t.Errorf("Error when decodeCommand on test2, error: %v", err)
	}
	if input != string(output) {
		t.Errorf("not match with current output, output: %v", output)
	}
}

func TestEncodeMessage(t *testing.T) {
	var inputHex = []byte{18, 48, 0, 10, 188}
	var output = "1230000abc"

	output1, err := encodeMessage(inputHex, v1alpha1.HEX)
	if err != nil {
		t.Errorf("Error when decodeCommand on test1, error: %v", err)
	}
	if output1 != output {
		t.Errorf("not match with current output, output: %v", output)
	}

	var inputUtf8 = []byte{49, 50, 51, 48, 48, 48, 48, 97, 98, 99}
	output2, err := encodeMessage(inputUtf8, v1alpha1.UTF8)
	if err != nil {
		t.Errorf("Error when decodeCommand on test1, error: %v", err)
	}
	if output2 != output {
		t.Errorf("not match with current output, output: %v", output)
	}
}

func TestCollectSocketTelemetry(t *testing.T) {
	socketProtocol := v1alpha1.ProtocolSocket
	httpProtocol := v1alpha1.ProtocolHTTP
	address := "localhost:3000"
	emptyAddress := ""
	db := &deviceshifubase.DeviceShifuBase{
		Name: "Unit Test",
		EdgeDevice: &v1alpha1.EdgeDevice{
			Spec: v1alpha1.EdgeDeviceSpec{
				Protocol: &socketProtocol,
				Address:  &address,
			},
		},
	}
	ds := &DeviceShifu{
		base: db,
	}

	listener, err := net.Listen("tcp", "localhost:3000")
	if err != nil {
		t.Errorf("Cannot Listen at port 3000")
	}

	go listener.Accept()

	// testcase pass
	ok, err := ds.collectSocketTelemetry()
	if err != nil {
		t.Errorf("Error when collectSocketTelemetry")
	}
	if !ok {
		t.Errorf("Fail to conn to mock server")
	}

	// testcase Address is nil
	ds.base.EdgeDevice.Spec.Address = nil
	ok, err = ds.collectSocketTelemetry()
	if err == nil || ok {
		t.Errorf("Error this case2 should return err but passed")
	}
	ds.base.EdgeDevice.Spec.Address = &address

	// testcase Protocol is not Socket
	ds.base.EdgeDevice.Spec.Protocol = &httpProtocol
	ok, err = ds.collectSocketTelemetry()
	if ok {
		t.Errorf("Error this case3 should return false but passed")
	}
	ds.base.EdgeDevice.Spec.Protocol = &socketProtocol

	// testcase Wrong ip
	ds.base.EdgeDevice.Spec.Address = &emptyAddress
	ok, err = ds.collectSocketTelemetry()
	if err == nil || ok {
		t.Errorf("Error this case4 should return err but passed")
	}

	// testcase Protocol is nil
	ds.base.EdgeDevice.Spec.Protocol = nil
	ok, err = ds.collectSocketTelemetry()
	if !ok {
		t.Errorf("Error this case5 should pass")
	}
}

func TestDeviceCommandHandlerSocket(t *testing.T) {
	hexEncoding := v1alpha1.HEX
	bufferLength := 10
	readBuffer := make([]byte, bufferLength)
	ds := &DeviceShifu{}
	server, client := net.Pipe()
	_ = ds
	go func() {
		for {
			_, err := server.Read(readBuffer)
			if err != nil {
				t.Error("Error when Read from pipe")
			}
			_, err = server.Write(readBuffer)
			if err != nil {
				t.Error("Error when Write to pipe")
			}
		}
	}()
	metadata := &HandlerMetaData{
		connection: &client,
		edgeDeviceSpec: v1alpha1.EdgeDeviceSpec{
			ProtocolSettings: &v1alpha1.ProtocolSettings{
				SocketSetting: &v1alpha1.SocketSetting{
					Encoding:     &hexEncoding,
					BufferLength: &bufferLength,
				},
			},
		},
	}

	requestBody := &RequestBody{
		Command: "1234567890",
		Timeout: 1,
	}
	failRequestBody := &RequestBody{
		Command: "a",
		Timeout: 1,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Errorf("Error when marshal request body to []byte, error: %v", err)
	}
	failBody, err := json.Marshal(failRequestBody)
	if err != nil {
		t.Errorf("Error when marshal failRequestBody to []byte, error: %v", err)
	}

	hs := httptest.NewServer(deviceCommandHandlerSocket(metadata))
	defer hs.Close()

	dc := mockRestClient(hs.URL, "testing")
	log.Println(dc.APIVersion())

	// testcase without Set Header Content-Type
	rs := dc.Post().Do(context.TODO())
	if rs.Error() == nil {
		t.Errorf("case should return Error but passed")
	}

	req := dc.Post().SetHeader("Content-Type", "application/json")

	// testcase requestBody is empty
	rs = req.Do(context.TODO())
	if rs.Error() == nil {
		t.Errorf("case should return Error but passed")
	}

	// testcase requestBody is not hex
	rs = req.Body(failBody).Do(context.TODO())
	if rs.Error() == nil {
		t.Errorf("case should return Error but passed")
	}

	// testcase pass
	rs = req.Body(body).Do(context.TODO())
	if rs.Error() != nil {
		t.Errorf("case should passed but return error: %v", rs.Error())
	}
}

func mockRestClient(host string, path string) *rest.RESTClient {
	c, err := rest.RESTClientFor(
		&rest.Config{
			Host:    host,
			APIPath: path,
			ContentConfig: rest.ContentConfig{
				GroupVersion:         &v1.SchemeGroupVersion,
				NegotiatedSerializer: scheme.Codecs.WithoutConversion(),
			},
		},
	)
	if err != nil {
		klog.Errorf("mock client for host %s, apipath: %s failed,", host, path)
		return nil
	}

	return c
}
