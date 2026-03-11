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
	availableFuncs := []string{
		"get_position",
		"get_status",
	}
	md, err := mockdevice.New("mockdevice_test", "0", availableFuncs, instructionHandler)
	require.NoError(t, err)

	stopCh := make(chan struct{})
	t.Cleanup(func() {
		close(stopCh)
	})

	require.NoError(t, md.Start(stopCh))

	baseURL := md.URL()
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
			"case 2 get_position",
			baseURL + "/get_position",
			200,
			true,
		},
	}

	testutil.WaitForHTTPServer(t, baseURL+"/health")

	for _, c := range mocks {
		t.Run(c.name, func(t *testing.T) {
			resp, err := http.Get(c.url)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, c.StatusCode, resp.StatusCode)
			body, _ := io.ReadAll(resp.Body)

			switch {
			case strings.Contains(c.url, "/get_position"):
				assert.Equal(t, c.expResult, check(string(body)))
			case strings.Contains(c.url, "/get_status"):
				assert.Contains(t, c.expResult, string(body))
			}
		})
	}
}

func check(Result string) bool {
	res := true
	expResult := []string{"xpos", "ypos"}
	for _, v := range expResult {
		if !strings.Contains(Result, v) {
			res = false
			break
		}
	}
	return res
}
