package main

import (
	"io"
	"net/http"
	"strconv"
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
		"read_value",
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
			"case 1 read_value",
			baseURL + "/read_value",
			200,
			true,
		},
		{
			"case 2 get_status",
			baseURL + "/get_status",
			200,
			[]string{"Running", "Idle", "Busy", "Error"},
		},
	}

	go mockdevice.StartMockDevice(availableFuncs, instructionHandler)

	testutil.WaitForHTTPServer(t, mocks[1].url)

	for _, c := range mocks {
		t.Run(c.name, func(t *testing.T) {
			resp, err := http.Get(c.url)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, c.StatusCode, resp.StatusCode)
			body, _ := io.ReadAll(resp.Body)

			switch {
			case strings.Contains(c.url, "/read_value"):
				assert.Equal(t, c.expResult, check(string(body)))
			case strings.Contains(c.url, "/get_status"):
				assert.Contains(t, c.expResult, string(body))
			}
		})
	}
}

func check(Result string) bool {
	res := true
	if _, err := strconv.Atoi(Result); err != nil {
		res = false
	}
	return res
}
