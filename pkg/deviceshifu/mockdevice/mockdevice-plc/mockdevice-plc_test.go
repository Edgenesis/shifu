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
	dataStorage = make(map[string]string)
	for _, v := range memoryArea {
		dataStorage[v] = originalCharacter
	}

	availableFuncs := []string{
		"getcontent",
		"sendsinglebit",
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
			"case 1 getcontent rootaddress nil",
			baseURL + "/getcontent",
			400,
			"Nonexistent memory area",
		},
		{
			"case 2 getcontent rootaddress=Q",
			baseURL + "/getcontent?rootaddress=Q",
			200,
			dataStorage["Q"],
		},
		{
			"case 3 sendsinglebit rootaddress nil address=0 digit=1",
			baseURL + "/sendsinglebit?digit=1&address=0",
			400,
			"Nonexistent memory area",
		},
		{
			"case 4 sendsinglebit rootaddress=Q address=0 digit=1",
			baseURL + "/sendsinglebit?rootaddress=Q&digit=1&address=0&value=1",
			200,
			"0b0000000000000010",
		},
		{
			"case 5 get_status",
			baseURL + "/get_status",
			200,
			[]string{"Running", "Idle", "Busy", "Error"},
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
			case strings.Contains(c.url, "/getcontent"):
				assert.Equal(t, c.expResult, string(body))
			case strings.Contains(c.url, "/get_status"):
				assert.Contains(t, c.expResult, string(body))
			}

		})
	}
}
