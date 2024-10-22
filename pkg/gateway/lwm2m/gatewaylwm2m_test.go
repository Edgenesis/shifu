package lwm2m

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock HTTP server to simulate responses
func mockServer(response string, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(response))
	}))
}

// Test for the Read method
func TestShifuInstruction_Read(t *testing.T) {
	tests := []struct {
		dataType       string
		response       string
		expectedResult interface{}
		expectedError  bool
		statusCode     int
	}{
		{"int", "123", 123, false, http.StatusOK},
		{"float", "123.45", 123.45, false, http.StatusOK},
		{"bool", "true", true, false, http.StatusOK},
		{"string", "test", "test", false, http.StatusOK},
		{"int", "not-an-int", nil, true, http.StatusOK}, // Invalid integer
		{"", "test", "test", false, http.StatusOK},      // Default to string
		{"", "test", nil, true, http.StatusBadRequest},  // Invalid status code
	}

	for _, test := range tests {
		server := mockServer(test.response, test.statusCode)
		defer server.Close()

		si := ShifuInstruction{
			Endpoint: server.URL,
			DataType: test.dataType,
		}

		result, err := si.Read()

		if test.expectedError {
			if err == nil {
				t.Errorf("expected an error but got none")
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != test.expectedResult {
				t.Errorf("expected %v but got %v", test.expectedResult, result)
			}
		}
	}
}

// Test for the Write method
func TestShifuInstruction_Write(t *testing.T) {
	server := mockServer("", http.StatusOK)
	defer server.Close()

	si := ShifuInstruction{
		Endpoint: server.URL,
	}

	err := si.Write(123) // Writing an integer as an example
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Simulate a failure
	server = mockServer("", http.StatusBadRequest)
	defer server.Close()

	si.Endpoint = server.URL
	err = si.Write(123)
	if err == nil {
		t.Errorf("expected an error but got none")
	}
}

// Test for the Execute method
func TestShifuInstruction_Execute(t *testing.T) {
	server := mockServer("", http.StatusOK)
	defer server.Close()

	si := ShifuInstruction{
		Endpoint: server.URL,
	}

	err := si.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Simulate a failure
	server = mockServer("", http.StatusBadRequest)
	defer server.Close()

	si.Endpoint = server.URL
	err = si.Execute()
	if err == nil {
		t.Errorf("expected an error but got none")
	}
}
