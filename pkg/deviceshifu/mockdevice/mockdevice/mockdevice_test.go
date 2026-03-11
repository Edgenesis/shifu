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

	md, err := New("mockdevice_test", "0", availableFuncs, instructionHandler)
	require.NoError(t, err)

	stopCh := make(chan struct{})
	t.Cleanup(func() {
		close(stopCh)
	})

	require.NoError(t, md.Start(stopCh))
	testutil.WaitForHTTPServer(t, md.URL()+"/health")

	resp, err := http.Get(md.URL() + "/get_status")
	require.NoError(t, err)

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	if string(body) != "Running" {
		t.Errorf("Body is not running: %+v", string(body))
	}
}
