package mockdevice

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/edgenesis/shifu/pkg/deviceshifu/mockdevice/testutil"
	"github.com/edgenesis/shifu/pkg/logger"
	"github.com/stretchr/testify/require"
)

func TestStartMockDevice(t *testing.T) {
	port := testutil.MustLocalhostPort(t)
	baseURL := "http://127.0.0.1:" + port

	t.Setenv("MOCKDEVICE_NAME", "mockdevice_test")
	t.Setenv("MOCKDEVICE_PORT", port)
	availableFuncs := []string{
		"get_position",
		"get_status",
	}

	instructionHandler := func(functionName string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			logger.Infof("Handling: %v", functionName)
			switch functionName {
			case "get_status":
				fmt.Fprintf(w, "Running")
			}
		}
	}

	go StartMockDevice(availableFuncs, instructionHandler)

	testutil.WaitForHTTPServer(t, baseURL+"/get_status")

	resp, err := http.Get(baseURL + "/get_status")
	require.NoError(t, err)

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	if string(body) != "Running" {
		t.Errorf("Body is not running: %+v", string(body))
	}
}
