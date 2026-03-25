package server

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/edgenesis/shifu/pkg/deviceapi"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// New creates a new MCP Server with list_devices and get_device_desc tools.
func New(apiClient *deviceapi.Client) *mcp.Server {
	s := mcp.NewServer(
		&mcp.Implementation{
			Name:    "shifu-mcp-server",
			Version: "v0.1.0",
		},
		nil,
	)

	// list_devices tool — no parameters.
	s.AddTool(
		&mcp.Tool{
			Name:        "list_devices",
			Description: "List all IoT devices in the Shifu cluster with their protocol, phase, and summary.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
		},
		listDevicesHandler(apiClient),
	)

	// get_device_desc tool — requires device_name parameter.
	s.AddTool(
		&mcp.Tool{
			Name:        "get_device_desc",
			Description: "Get the full documentation for a device — what it is, how to connect, and all interactions with usage examples.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"device_name":{"type":"string","description":"Name of the EdgeDevice"}},"required":["device_name"]}`),
		},
		getDeviceDescHandler(apiClient),
	)

	return s
}

func listDevicesHandler(apiClient *deviceapi.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		devices, err := apiClient.ListDevices(ctx)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: "Failed to list devices: " + err.Error()}},
			}, nil
		}

		data, err := json.MarshalIndent(devices, "", "  ")
		if err != nil {
			return nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil
	}
}

func getDeviceDescHandler(apiClient *deviceapi.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args struct {
			DeviceName string `json:"device_name"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: "Invalid arguments: " + err.Error()}},
			}, nil
		}

		if args.DeviceName == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: "device_name is required"}},
			}, nil
		}

		desc, err := apiClient.GetDeviceDesc(ctx, args.DeviceName)
		if err != nil {
			var notFound *deviceapi.DeviceNotFoundError
			if errors.As(err, &notFound) {
				errResp := map[string]string{
					"error":   "DEVICE_NOT_FOUND",
					"message": notFound.Error(),
				}
				data, _ := json.MarshalIndent(errResp, "", "  ")
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
				}, nil
			}
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: "Failed to get device description: " + err.Error()}},
			}, nil
		}

		data, err := json.MarshalIndent(desc, "", "  ")
		if err != nil {
			return nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil
	}
}
