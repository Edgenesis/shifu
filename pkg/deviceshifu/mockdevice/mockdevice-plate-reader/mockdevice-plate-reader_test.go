package main

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/edgenesis/shifu/pkg/deviceshifu/mockdevice/mockdevice"
	"github.com/edgenesis/shifu/pkg/deviceshifu/mockdevice/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstructionHandler(t *testing.T) {
	port := testutil.MustLocalhostPort(t)
	baseURL := "http://127.0.0.1:" + port

	availableFuncs := []string{
		"get_measurement",
		"get_status",
	}
	t.Setenv("MOCKDEVICE_NAME", "mockdevice_test")
	t.Setenv("MOCKDEVICE_PORT", port)
	mocks := []struct {
		name       string
		url        string
		StatusCode int
		expResult  interface{}
	}{
		{
			"case 1 get_status",
			baseURL + "/get_status",
			200,
			[]string{"Running", "Idle", "Busy", "Error"},
		},
		{
			"case 2 get_measurement",
			baseURL + "/get_measurement",
			200,
			true,
		},
	}

	go mockdevice.StartMockDevice(availableFuncs, instructionHandler)

	testutil.WaitForHTTPServer(t, mocks[0].url)

	for _, c := range mocks {
		t.Run(c.name, func(t *testing.T) {
			resp, err := http.Get(c.url)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, c.StatusCode, resp.StatusCode)
			body, _ := io.ReadAll(resp.Body)

			switch {
			case strings.Contains(c.url, "/get_measurement"):
				assert.Equal(t, c.expResult, check(body))
			case strings.Contains(c.url, "/get_status"):
				assert.Contains(t, c.expResult, string(body))
			}
		})
	}
}

func check(Result interface{}) bool {
	res := true
	if Result == nil {
		res = false
	}
	return res
}
