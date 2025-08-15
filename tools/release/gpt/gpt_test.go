package gpt

import (
	"os"
	"strings"
	"testing"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid configuration",
			config: Config{
				APIKey:         "test-key",
				Host:           "https://test.openai.azure.com/",
				DeploymentName: "gpt-4",
				Version:        "v1.0.0",
			},
			expectError: false,
		},
		{
			name: "Missing API key",
			config: Config{
				Host:           "https://test.openai.azure.com/",
				DeploymentName: "gpt-4",
				Version:        "v1.0.0",
			},
			expectError: true,
			errorMsg:    "AZURE_OPENAI_APIKEY",
		},
		{
			name: "Invalid host URL - HTTP instead of HTTPS",
			config: Config{
				APIKey:         "test-key",
				Host:           "http://invalid.com/",
				DeploymentName: "gpt-4",
				Version:        "v1.0.0",
			},
			expectError: true,
			errorMsg:    "host must use HTTPS scheme",
		},
		{
			name: "Invalid host URL - malformed URL",
			config: Config{
				APIKey:         "test-key",
				Host:           "://invalid-url",
				DeploymentName: "gpt-4",
				Version:        "v1.0.0",
			},
			expectError: true,
			errorMsg:    "invalid host URL format",
		},
		{
			name: "Invalid host URL - missing hostname",
			config: Config{
				APIKey:         "test-key",
				Host:           "https://",
				DeploymentName: "gpt-4",
				Version:        "v1.0.0",
			},
			expectError: true,
			errorMsg:    "host URL must have a valid hostname",
		},
		{
			name: "Missing deployment name",
			config: Config{
				APIKey:  "test-key",
				Host:    "https://test.openai.azure.com/",
				Version: "v1.0.0",
			},
			expectError: true,
			errorMsg:    "DEPLOYMENT_NAME",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got: %s", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %s", err.Error())
				}
			}
		})
	}
}

func TestNewConfigFromEnv(t *testing.T) {
	// Save original env vars
	originalVars := map[string]string{
		"AZURE_OPENAI_APIKEY": os.Getenv("AZURE_OPENAI_APIKEY"),
		"AZURE_OPENAI_HOST":   os.Getenv("AZURE_OPENAI_HOST"),
		"DEPLOYMENT_NAME":     os.Getenv("DEPLOYMENT_NAME"),
		"VERSION":             os.Getenv("VERSION"),
	}

	// Clean up function
	defer func() {
		for key, value := range originalVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	// Test with valid environment
	os.Setenv("AZURE_OPENAI_APIKEY", "test-key")
	os.Setenv("AZURE_OPENAI_HOST", "https://test.openai.azure.com/")
	os.Setenv("DEPLOYMENT_NAME", "gpt-4")
	os.Setenv("VERSION", "v1.0.0")

	config, err := NewConfigFromEnv()
	if err != nil {
		t.Errorf("Expected no error with valid environment, got: %s", err.Error())
		return
	}

	if config.APIKey != "test-key" {
		t.Errorf("Expected APIKey 'test-key', got: %s", config.APIKey)
	}
	if config.APIVersion != DefaultAPIVersion {
		t.Errorf("Expected APIVersion '%s', got: %s", DefaultAPIVersion, config.APIVersion)
	}
}

func TestProcessContent(t *testing.T) {
	generator := &ChangelogGenerator{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Remove code blocks",
			input:    "```markdown\n# Test\n```",
			expected: "markdown\n\n# Test",
		},
		{
			name:     "Remove empty lines",
			input:    "Line 1\n\n\nLine 2\n\n",
			expected: "Line 1\n\nLine 2",
		},
		{
			name:     "Complex content",
			input:    "```\n# Header\n\nContent\n\n\n\nMore content\n```",
			expected: "# Header\n\nContent\n\nMore content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.processContent(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildMessages(t *testing.T) {
	generator := &ChangelogGenerator{}

	releaseNotes := "Test release notes"
	messages := generator.buildMessages(releaseNotes)

	if len(messages) != 5 {
		t.Errorf("Expected 5 messages, got %d", len(messages))
	}

	// The last message should be the release notes
	// Note: We can't easily test the exact content due to the opaque message structure
	// but we can verify the count and that the function doesn't panic
}

// Benchmark tests
func BenchmarkProcessContent(b *testing.B) {
	generator := &ChangelogGenerator{}
	content := strings.Repeat("```\nTest content\n\n\nMore content\n```\n", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator.processContent(content)
	}
}
