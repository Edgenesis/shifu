package main

import (
	"github.com/edgenesis/shifu/pkg/deviceshifu/mockdevice/mockdevice"
	"github.com/stretchr/testify/assert"
	"io"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"
)

func TestInstructionHandler(t *testing.T) {
	availableFuncs := []string{
		"read_value",
		"get_status",
	}
	os.Setenv("MOCKDEVICE_NAME", "mockdevice_test")
	os.Setenv("MOCKDEVICE_PORT", "12345")
	mocks := []struct {
		name       string
		url        string
		StatusCode int
		expResult  interface{}
	}{
		{
			"case 1 prot 12345 read_value",
			"http://localhost:12345/read_value",
			200,
			"",
		},
		{
			"case 2 prot 12345 get_status",
			"http://localhost:12345/get_status",
			200,
			[]string{"Running", "Idle", "Busy", "Error"},
		},
	}

	go mockdevice.StartMockDevice(availableFuncs, instructionHandler)

	time.Sleep(1 * time.Second)

	for _, c := range mocks {
		t.Run(c.name, func(t *testing.T) {
			resp, err := http.Get(c.url)
			if err != nil {
				t.Fatalf("HTTP GET returns an error %v", err.Error())
			}
			defer resp.Body.Close()
			assert.Nil(t, err)
			body, err := io.ReadAll(resp.Body)

			if c.name == mocks[len(mocks)-1].name {
				assert.Contains(t, c.expResult, string(body))
				return
			}
			if !regexp.MustCompile("\\d").Match(body) {
				klog.Error("read_value Body error")
			}
		})
	}
}
