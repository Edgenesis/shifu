package main

import (
	"github.com/edgenesis/shifu/pkg/deviceshifu/mockdevice/mockdevice"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestInstructionHandler(t *testing.T) {
	availableFuncs := []string{
		"get_position",
		"get_status",
	}
	os.Setenv("MOCKDEVICE_NAME", "mockdevice_test")
	os.Setenv("MOCKDEVICE_PORT", "12345")
	mocks := []struct {
		name string
		url  string
	}{
		{
			"case 1 prot 12345 get_status",
			"http://localhost:12345/get_status",
		},
		{
			"case 2 prot 12345 get_position",
			"http://localhost:12345/get_position",
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

			body, _ := io.ReadAll(resp.Body)
			assert.NotNil(t, body)
		})
	}
}
