package deviceapi

// DeviceSummary is a brief summary of a device, returned by ListDevices.
type DeviceSummary struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Description string `json:"description,omitempty"`
	Protocol    string `json:"protocol,omitempty"`
	Phase       string `json:"phase,omitempty"`
	Service     string `json:"service,omitempty"`
}

// DeviceDesc is the full description of a device, returned by GetDeviceDesc.
type DeviceDesc struct {
	Name           string        `json:"name"`
	Description    string        `json:"description,omitempty"`
	Protocol       string        `json:"protocol,omitempty"`
	Phase          string        `json:"phase,omitempty"`
	Service        string        `json:"service,omitempty"`
	ConnectionInfo string        `json:"connectionInfo,omitempty"`
	Interactions   []Interaction `json:"interactions,omitempty"`
}

// Interaction represents a single device instruction/interaction.
type Interaction struct {
	Name        string `json:"name"`
	ReadWrite   string `json:"readWrite,omitempty"`
	Safe        *bool  `json:"safe,omitempty"`
	Description string `json:"description,omitempty"`
}
