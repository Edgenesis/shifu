// Package tdengine contains tests for the TDengine database interactions.
// IMPORTANT: Do not run these tests with the -race flag due to a known issue.
// For more details, see the GitHub issue: https://github.com/taosdata/driver-go/issues/185
package telemetryservice

import (
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	testCases := []struct {
		name string
		mux  *http.ServeMux
	}{
		{
			name: "case1 pass",
			mux:  http.NewServeMux(),
		}, {
			name: "case2 without mux",
		},
	}

	for _, c := range testCases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			addr := getFreeLocalAddr(t)
			stopChan := make(chan struct{}, 1)
			go func(ch chan struct{}) {
				time.Sleep(10 * time.Millisecond)
				ch <- struct{}{}
			}(stopChan)
			err := Start(stopChan, c.mux, addr)
			assert.Nil(t, err)
		})
	}
}

func TestNew(t *testing.T) {
	stop := make(chan struct{}, 1)
	go func() {
		time.Sleep(100 * time.Millisecond)
		stop <- struct{}{}
	}()
	New(stop)
}

func getFreeLocalAddr(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	addr := l.Addr().String()
	if err := l.Close(); err != nil {
		t.Fatalf("failed to release listener: %v", err)
	}
	return addr
}
