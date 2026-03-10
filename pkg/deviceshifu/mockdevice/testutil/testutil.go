package testutil

import (
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func MustLocalhostPort(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	_, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)

	return port
}

func WaitForHTTPServer(t *testing.T, url string) {
	t.Helper()

	client := &http.Client{Timeout: 100 * time.Millisecond}
	require.Eventually(t, func() bool {
		resp, err := client.Get(url)
		if err != nil {
			return false
		}
		defer resp.Body.Close()

		return resp.StatusCode == http.StatusOK
	}, 2*time.Second, 20*time.Millisecond)
}
