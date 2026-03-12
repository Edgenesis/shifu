package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgenesis/shifu/pkg/deviceapi"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/fake"
)

// TestE2EStreamableHTTP tests the MCP server over real HTTP Streamable transport,
// simulating what an AI agent like Claude Code would do.
func TestE2EStreamableHTTP(t *testing.T) {
	// Setup fake K8s + device data.
	devices := testDevices()
	k8sObjects := testK8sObjects()
	fakeClient := fake.NewSimpleClientset(k8sObjects...)

	lister := func(ctx context.Context) ([]v1alpha1.EdgeDevice, error) {
		return devices, nil
	}

	resolver := deviceapi.NewResolver(fakeClient, lister)
	apiClient := deviceapi.NewClient(resolver)
	mcpServer := New(apiClient)

	// Start HTTP server with Streamable HTTP handler.
	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return mcpServer
	}, nil)
	httpServer := httptest.NewServer(handler)
	defer httpServer.Close()

	// Connect an MCP client via Streamable HTTP.
	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "e2e-client", Version: "v0.0.1"}, nil)
	transport := &mcp.StreamableClientTransport{Endpoint: httpServer.URL}
	session, err := client.Connect(ctx, transport, nil)
	require.NoError(t, err)
	defer session.Close()

	// Test 1: List tools
	t.Run("ListTools", func(t *testing.T) {
		var toolNames []string
		for tool, err := range session.Tools(ctx, nil) {
			require.NoError(t, err)
			toolNames = append(toolNames, tool.Name)
		}
		assert.Contains(t, toolNames, "list_devices")
		assert.Contains(t, toolNames, "get_device_desc")
	})

	// Test 2: list_devices via HTTP
	t.Run("ListDevicesHTTP", func(t *testing.T) {
		result, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: "list_devices",
		})
		require.NoError(t, err)
		require.False(t, result.IsError)

		text := result.Content[0].(*mcp.TextContent).Text
		var devices []deviceapi.DeviceSummary
		err = json.Unmarshal([]byte(text), &devices)
		require.NoError(t, err)
		assert.Len(t, devices, 2)
	})

	// Test 3: get_device_desc for HTTP device
	t.Run("GetDeviceDescHTTP", func(t *testing.T) {
		result, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "get_device_desc",
			Arguments: map[string]any{"device_name": "edgedevice-thermometer"},
		})
		require.NoError(t, err)
		require.False(t, result.IsError)

		text := result.Content[0].(*mcp.TextContent).Text
		var desc deviceapi.DeviceDesc
		err = json.Unmarshal([]byte(text), &desc)
		require.NoError(t, err)
		assert.Equal(t, "HTTP", desc.Protocol)
		assert.Contains(t, desc.ConnectionInfo, "Base URL")
		assert.GreaterOrEqual(t, len(desc.Interactions), 2)
	})

	// Test 4: get_device_desc for MQTT device
	t.Run("GetDeviceDescMQTT", func(t *testing.T) {
		result, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "get_device_desc",
			Arguments: map[string]any{"device_name": "edgedevice-robot-arm"},
		})
		require.NoError(t, err)
		require.False(t, result.IsError)

		text := result.Content[0].(*mcp.TextContent).Text
		var desc deviceapi.DeviceDesc
		err = json.Unmarshal([]byte(text), &desc)
		require.NoError(t, err)
		assert.Equal(t, "MQTT", desc.Protocol)
		assert.Contains(t, desc.ConnectionInfo, "MQTT broker")
		assert.GreaterOrEqual(t, len(desc.Interactions), 2)
	})

	// Test 5: Device not found via HTTP
	t.Run("DeviceNotFoundHTTP", func(t *testing.T) {
		result, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "get_device_desc",
			Arguments: map[string]any{"device_name": "nonexistent"},
		})
		require.NoError(t, err)
		require.True(t, result.IsError)

		text := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, text, "DEVICE_NOT_FOUND")
	})
}
