package lwm2m

import (
	"reflect"
	"testing"

	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/pion/dtls/v3"
)

func TestCipherSuiteStringsToCodes(t *testing.T) {
	tests := []struct {
		name           string
		input          []v1alpha1.CipherSuite
		expectedOutput []dtls.CipherSuiteID
		expectError    bool
	}{
		{
			name: "Valid cipher suites",
			input: []v1alpha1.CipherSuite{
				v1alpha1.CipherSuite_TLS_ECDHE_ECDSA_WITH_AES_128_CCM,
				v1alpha1.CipherSuite_TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
			expectedOutput: []dtls.CipherSuiteID{
				TLS_ECDHE_ECDSA_WITH_AES_128_CCM,
				TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
			expectError: false,
		},
		{
			name: "Invalid cipher suite",
			input: []v1alpha1.CipherSuite{
				v1alpha1.CipherSuite("INVALID_CIPHER_SUITE"),
			},
			expectedOutput: nil,
			expectError:    true,
		},
		{
			name: "Mixed valid and invalid cipher suites",
			input: []v1alpha1.CipherSuite{
				v1alpha1.CipherSuite_TLS_ECDHE_ECDSA_WITH_AES_128_CCM,
				v1alpha1.CipherSuite("INVALID_CIPHER_SUITE"),
			},
			expectedOutput: nil,
			expectError:    true,
		},
		{
			name:           "Empty input",
			input:          []v1alpha1.CipherSuite{},
			expectedOutput: []dtls.CipherSuiteID{},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := CipherSuiteStringsToCodes(tt.input)
			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v", tt.expectError, err)
			}
			if !tt.expectError && !reflect.DeepEqual(output, tt.expectedOutput) {
				t.Errorf("expected output: %v, got: %v", tt.expectedOutput, output)
			}
		})
	}
}
