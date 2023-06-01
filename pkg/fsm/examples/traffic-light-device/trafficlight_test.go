package main

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/edgenesis/shifu/pkg/deviceshifu/mockdevice/mockdevice"
	"github.com/stretchr/testify/assert"
)

func TestInstructionHandler(t *testing.T) {
	trafficLight = NewTrafficLight(RED)

	availableFuncs := []string{
		"stop",
		"caution",
		"proceed",
		"get_color",
		"get_status",
	}
	t.Setenv("MOCKDEVICE_NAME", "mockdevice_test")
	t.Setenv("MOCKDEVICE_PORT", "12345")
	mocks := []struct {
		name       string
		url        string
		StatusCode int
		expResult  interface{}
	}{
		{
			"case 1 get_color",
			"http://localhost:12345/get_color",
			200,
			"traffic light current state: RED",
		},
		{
			"case 2 RED stop",
			"http://localhost:12345/stop",
			200,
			"Disable transition from RED to RED, must be YELLOW",
		},
		{
			"case 3 RED proceed",
			"http://localhost:12345/proceed",
			200,
			"Transition from RED to GREEN",
		},
		{
			"case 4 GREEN caution",
			"http://localhost:12345/caution",
			200,
			"Transition from GREEN to YELLOW",
		},
		{
			"case 5 YELLOW stop",
			"http://localhost:12345/stop",
			200,
			"Transition from YELLOW to RED",
		},
		{
			"case 5 get_status",
			"http://localhost:12345/get_status",
			200,
			[]string{"Running", "Idle", "Busy", "Error"},
		},
	}

	go mockdevice.StartMockDevice(availableFuncs, instructionHandler)

	time.Sleep(100 * time.Microsecond)

	for _, c := range mocks {
		t.Run(c.name, func(t *testing.T) {
			resp, err := http.Get(c.url)
			assert.Nil(t, err)
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			switch {
			case strings.Contains(c.url, "/get_color"):
				assert.Equal(t, c.expResult, string(body))
			case strings.Contains(c.url, "/stop"):
				assert.Equal(t, c.expResult, string(body))
			case strings.Contains(c.url, "/caution"):
				assert.Equal(t, c.expResult, string(body))
			case strings.Contains(c.url, "/proceed"):
				assert.Equal(t, c.expResult, string(body))
			case strings.Contains(c.url, "/get_status"):
				assert.Contains(t, c.expResult, string(body))
			}

		})
	}
}
