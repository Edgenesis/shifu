package rtspRecord

type Request interface {
	RegisterRequest | UnregisterRequest | UpdateRequest
}

type RegisterRequest struct {
	DeviceName    string `json:"deviceName"`
	SecretName    string `json:"secretName"`
	ServerAddress string `json:"serverAddress"`
	Record        bool   `json:"record"`
}

type UnregisterRequest struct {
	DeviceName string `json:"deviceName"`
}

type UpdateRequest struct {
	DeviceName string `json:"deviceName"`
	Record     bool   `json:"record"`
}
