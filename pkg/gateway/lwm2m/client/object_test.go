package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock ObjectAPI for testing
type MockObjectAPI struct{}

func (m *MockObjectAPI) Read() (interface{}, error) {
	return "mock-value", nil
}

func (m *MockObjectAPI) Write(interface{}) error {
	return nil
}

func (m *MockObjectAPI) Execute() error {
	return nil
}

// Test NewObject
func TestNewObject(t *testing.T) {
	mockAPI := &MockObjectAPI{}
	obj := NewObject("1", mockAPI)

	assert.Equal(t, "1", obj.Id, "Object Id should be 1")
	assert.NotNil(t, obj.Child, "Object Child map should be initialized")
	assert.Equal(t, mockAPI, obj.ObjectAPI, "Object API should be set to mock API")
}

// Test GetChildObject
func TestObject_GetChildObject(t *testing.T) {
	mockAPI := &MockObjectAPI{}
	root := NewObject(rootObjectId, mockAPI)

	// Add child objects
	root.AddObject("/1/0", mockAPI)
	root.AddObject("/1/1", mockAPI)

	child := root.GetChildObject("/1/0")
	assert.NotNil(t, child, "Child object /1/0 should exist")
	assert.Equal(t, "0", child.Id, "Child object ID should be 0")

	nonExistentChild := root.GetChildObject("/2/0")
	assert.Nil(t, nonExistentChild, "Non-existent child should return nil")
}

// Test AddObject
func TestObject_AddObject(t *testing.T) {
	mockAPI := &MockObjectAPI{}
	root := NewObject("root", mockAPI)

	// Add child objects
	root.AddObject("/1/0", mockAPI)
	root.AddObject("/1/1", mockAPI)

	child := root.GetChildObject("/1/0")
	assert.NotNil(t, child, "Child object /1/0 should exist")
	assert.Equal(t, "0", child.Id, "Child object ID should be 0")

	child1 := root.GetChildObject("/1/1")
	assert.NotNil(t, child1, "Child object /1/1 should exist")
	assert.Equal(t, "1", child1.Id, "Child object ID should be 1")
}

// Test ReadAll
func TestObject_ReadAll(t *testing.T) {
	mockAPI := &MockObjectAPI{}
	root := NewObject(rootObjectId, mockAPI)

	// Add child objects with values
	root.AddObject("/1/0", mockAPI)
	root.AddObject("/1/1", mockAPI)

	resource, err := root.ReadAll("/1")
	assert.NoError(t, err, "Reading all data should not return an error")
	assert.Equal(t, "/1/", resource.BaseName, "BaseName should match")
	assert.Len(t, resource.ResourceArray, 2, "There should be two resource entries")
}

// Test GetAllChildPaths
func TestObject_GetAllChildPaths(t *testing.T) {
	mockAPI := &MockObjectAPI{}
	root := NewObject(rootObjectId, mockAPI)

	// Add child objects
	root.AddObject("/1/0", mockAPI)
	root.AddObject("/1/1", mockAPI)

	paths := root.GetAllChildPaths()
	expectedPaths := []string{"/1", "/1/0", "/1/1"}
	assert.ElementsMatch(t, expectedPaths, paths, "Child paths should match expected paths")
}

// Test GetCoRELinkString
func TestObject_GetCoRELinkString(t *testing.T) {
	mockAPI := &MockObjectAPI{}
	root := NewObject(rootObjectId, mockAPI)

	// Add child objects
	root.AddObject("/1/0", mockAPI)
	root.AddObject("/1/1", mockAPI)

	coreLinkString := root.GetCoRELinkString()
	expected := `</>;rt="oma.lwm2m";ct="11543",</1>,</1/0>,</1/1>`
	assert.Equal(t, expected, coreLinkString, "CoRE Link format string should match expected format")
}
