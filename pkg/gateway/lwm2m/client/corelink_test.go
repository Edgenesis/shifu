package client

import (
	"testing"
)

// Test for the NewLink function
func TestNewLink(t *testing.T) {
	// Test case: valid resource path and attributes
	resourcePath := "/1/0"
	attributes := map[string]string{
		"rt": "oma.lwm2m",
		"ct": "0",
	}
	link := NewLink(resourcePath, attributes)

	if link.ResourcePath != resourcePath {
		t.Errorf("expected ResourcePath to be %s, got %s", resourcePath, link.ResourcePath)
	}

	if len(link.Attributes) != len(attributes) {
		t.Errorf("expected Attributes length to be %d, got %d", len(attributes), len(link.Attributes))
	}

	for key, value := range attributes {
		if link.Attributes[key] != value {
			t.Errorf("expected attribute %s to be %s, got %s", key, value, link.Attributes[key])
		}
	}
}

// Test for the String function
func TestLink_String(t *testing.T) {
	tests := []struct {
		name         string
		resourcePath string
		attributes   map[string]string
		expected     string
	}{
		{
			name:         "Single attribute",
			resourcePath: "/1/0",
			attributes: map[string]string{
				"rt": "oma.lwm2m",
			},
			expected: `</1/0>;rt="oma.lwm2m"`,
		},
		{
			name:         "Multiple attributes",
			resourcePath: "/1/0",
			attributes: map[string]string{
				"rt": "oma.lwm2m",
				"ct": "0",
			},
			expected: `</1/0>;rt="oma.lwm2m",ct="0"`,
		},
		{
			name:         "No attributes",
			resourcePath: "/1/0",
			attributes:   map[string]string{},
			expected:     `</1/0>;`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			link := NewLink(test.resourcePath, test.attributes)
			result := link.String()

			if result != test.expected {
				t.Errorf("expected %s, got %s", test.expected, result)
			}
		})
	}
}
