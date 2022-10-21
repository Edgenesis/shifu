package main

import (
	"github.com/edgenesis/shifu/pkg/deviceshifu/mockdevice/mockdevice"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
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
	os.Setenv("MOCKDEVICE_NAME", "mockdevice_test")
	os.Setenv("MOCKDEVICE_PORT", "12345")
	mocks := []struct {
		name       string
		url        string
		StatusCode int
		expResult  string
	}{
		{
			"case 1 getcontent rootaddress nil",
			"http://localhost:12345/getcontent",
			400,
			"Nonexistent memory area",
		},
		{
			"case 2 getcontent rootaddress=Q",
			"http://localhost:12345/getcontent?rootaddress=Q",
			200,
			dataStorage["Q"],
		},
		{
			"case 3 sendsinglebit rootaddress nil address=0 digit=1",
			"http://localhost:12345/sendsinglebit?digit=1&address=0",
			400,
			"Nonexistent memory area",
		},
		{
			"case 4 sendsinglebit rootaddress=Q address=0 digit=1",
			"http://localhost:12345/sendsinglebit?rootaddress=Q&digit=1&address=0&value=1",
			200,
			"0b0000000000000010",
		},
		{
			"case 5 get_status",
			"http://localhost:12345/get_status",
			200,
			"",
		},
	}

	go mockdevice.StartMockDevice(availableFuncs, instructionHandler)

	time.Sleep(1 * time.Second)

	for _, c := range mocks {
		t.Run(c.name, func(t *testing.T) {
			resp, err := http.Get(c.url)
			if err != nil {
				t.Fatalf("HTTP returns an error %v", err.Error())
			}
			assert.Equal(t, c.StatusCode, resp.StatusCode)

			defer resp.Body.Close()
			assert.Nil(t, err)

			body, _ := io.ReadAll(resp.Body)

			if c.name == "case 5 get_status" {
				assert.Contains(t, []string{"Running", "Idle", "Busy", "Error"}, string(body))
				return
			}
			assert.Equal(t, c.expResult, string(body))

		})
	}
}
