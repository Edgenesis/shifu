package testutil

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func WaitForHTTPServer(t *testing.T, url string) {
	t.Helper()

	client := &http.Client{Timeout: 100 * time.Millisecond}
	require.Eventually(t, func() bool {
		resp, err := client.Get(url)
		if err != nil {
			return false
		}
		defer resp.Body.Close()

		return true
	}, 2*time.Second, 20*time.Millisecond)
}
