package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string { return &s }

func TestEdgeDeviceSpecDeepCopy(t *testing.T) {
	spec := &EdgeDeviceSpec{
		Description: strPtr("test device"),
	}

	copied := spec.DeepCopy()
	require.NotNil(t, copied)

	// Values match.
	assert.Equal(t, "test device", *copied.Description)

	// Pointers are distinct (true deep copy).
	assert.NotSame(t, spec.Description, copied.Description)

	// Mutating the copy does not affect the original.
	*copied.Description = "mutated"
	assert.Equal(t, "test device", *spec.Description)
}

func TestEdgeDeviceSpecDeepCopyNilFields(t *testing.T) {
	spec := &EdgeDeviceSpec{}

	copied := spec.DeepCopy()
	require.NotNil(t, copied)
	assert.Nil(t, copied.Description)
}
