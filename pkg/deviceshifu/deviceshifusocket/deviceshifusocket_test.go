package deviceshifusocket

import (
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifubase"
	"github.com/edgenesis/shifu/pkg/deviceshifu/utils"

	"k8s.io/apimachinery/pkg/util/wait"
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

	resp, err := utils.RetryAndGetHTTP("http://localhost:8080/health", 3)
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

func TestPraseCommandWithEncode(t *testing.T) {
	input := "1230000abc"
	var outputHex = []byte{18, 48, 0, 10, 188}

	output, err := praseCommandWithEncode(input, EncodeHexStr)
	if err != nil {
		t.Errorf("%+v", err)
	}
	if !reflect.DeepEqual(output, outputHex) {
		t.Errorf("not match with current output, output: %v", output)
	}

	output, err = praseCommandWithEncode(input, EncodeUTF8Str)
	if err != nil {
		t.Errorf("%+v", err)
	}
	if input != string(output) {
		t.Errorf("not match with current output, output: %v", output)
	}
}

func TestEncodeMessageWithEncoding(t *testing.T) {
	var inputHex = []byte{18, 48, 0, 10, 188}
	var output = "1230000abc"

	output1, err := encodeMessageWithEncoding(inputHex, EncodeHexStr)
	if err != nil {
		t.Errorf("%+v", err)
	}
	if output1 != output {
		t.Errorf("not match with current output, output: %v", output)
	}

	var inputUtf8 = []byte{49, 50, 51, 48, 48, 48, 48, 97, 98, 99}
	output2, err := encodeMessageWithEncoding(inputUtf8, EncodeUTF8Str)
	if err != nil {
		t.Errorf("%+v", err)
	}
	if output2 != output {
		t.Errorf("not match with current output, output: %v", output)
	}
}
