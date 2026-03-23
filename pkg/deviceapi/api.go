package deviceapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
)

// Client provides the device API — ListDevices and GetDeviceDesc.
type Client struct {
	resolver *Resolver
}

// NewClient creates a new device API Client.
func NewClient(resolver *Resolver) *Client {
	return &Client{resolver: resolver}
}

// ListDevices returns a summary of all EdgeDevice resources in the cluster.
func (c *Client) ListDevices(ctx context.Context) ([]DeviceSummary, error) {
	devices, err := c.resolver.edgeDeviceLister(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing edge devices: %w", err)
	}

	var summaries []DeviceSummary
	for i := range devices {
		ed := &devices[i]
		summary := DeviceSummary{
			Name:      ed.Name,
			Namespace: ed.Namespace,
		}

		if ed.Spec.Protocol != nil {
			summary.Protocol = string(*ed.Spec.Protocol)
		}
		if ed.Status.EdgeDevicePhase != nil {
			summary.Phase = string(*ed.Status.EdgeDevicePhase)
		}
		if ed.Spec.Description != nil {
			// Use only the first line for the summary.
			desc := strings.TrimSpace(*ed.Spec.Description)
			if idx := strings.Index(desc, "\n"); idx > 0 {
				desc = desc[:idx]
			}
			summary.Description = desc
		}

		// Resolve the DeviceShifu service for this device.
		info, _ := c.resolver.resolveDeployment(ctx, ed.Name)
		if info != nil {
			summary.Service = info.ServiceName
		}

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// GetDeviceDesc returns the full description for a specific device.
func (c *Client) GetDeviceDesc(ctx context.Context, deviceName string) (*DeviceDesc, error) {
	devices, err := c.resolver.edgeDeviceLister(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing edge devices: %w", err)
	}

	var ed *v1alpha1.EdgeDevice
	for i := range devices {
		if devices[i].Name == deviceName {
			ed = &devices[i]
			break
		}
	}

	if ed == nil {
		return nil, &DeviceNotFoundError{Name: deviceName}
	}

	desc := &DeviceDesc{
		Name: ed.Name,
	}

	if ed.Spec.Protocol != nil {
		desc.Protocol = string(*ed.Spec.Protocol)
	}
	if ed.Status.EdgeDevicePhase != nil {
		desc.Phase = string(*ed.Status.EdgeDevicePhase)
	}
	if ed.Spec.Description != nil {
		desc.Description = strings.TrimSpace(*ed.Spec.Description)
	}
	if ed.Spec.ConnectionInfo != nil {
		desc.ConnectionInfo = strings.TrimSpace(*ed.Spec.ConnectionInfo)
	}

	// Resolve the DeviceShifu deployment, service, and ConfigMap.
	info, err := c.resolver.resolveDeployment(ctx, ed.Name)
	if err != nil {
		return desc, nil // Return what we have even if resolution fails.
	}
	if info != nil {
		desc.Service = info.ServiceName
		interactions, err := c.resolver.parseInstructions(ctx, info.ConfigMapName)
		if err != nil {
			return desc, nil
		}
		desc.Interactions = interactions
	}

	return desc, nil
}

// DeviceNotFoundError is returned when a device is not found.
type DeviceNotFoundError struct {
	Name string
}

func (e *DeviceNotFoundError) Error() string {
	return fmt.Sprintf("EdgeDevice '%s' not found in any namespace", e.Name)
}
